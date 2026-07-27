//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/application"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/domain"
	livekitapp "github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
)

const integrationTimeout = 3 * time.Minute

type recordingIssuer struct {
	requests []livekitapp.JoinRequest
}

func (issuer *recordingIssuer) Issue(_ context.Context, request livekitapp.JoinRequest) (livekitapp.JoinToken, error) {
	issuer.requests = append(issuer.requests, request)
	return livekitapp.JoinToken{Signed: "signed:" + request.ParticipantKey[:8], ExpiresAt: request.ServerTime.Add(request.TTL)}, nil
}

func TestCallsEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_calls_test")
	if _, err := database.Collection("rooms").InsertOne(ctx, bson.M{
		"_id": "room_1", "members": bson.A{"m-1", "m-2"}, "createdAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	issuer := &recordingIssuer{}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "call_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewCallService(repository, mongodb.NewRoomMembership(database), issuer, time.Now, ids)

	// Non-member cannot call.
	if _, err := service.Initiate(ctx, "room_1", "m-1", "m-9"); err != domain.ErrNotRoomMember {
		t.Fatalf("non-member = %v, want ErrNotRoomMember", err)
	}

	// Initiate: persisted, one opaque speaker token per participant.
	issued, err := service.Initiate(ctx, "room_1", "m-1", "m-2")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	if len(issued.Tokens) != 2 || len(issuer.requests) != 2 {
		t.Fatalf("tokens = %#v", issued.Tokens)
	}
	for _, request := range issuer.requests {
		if len(request.ParticipantKey) != 64 || len(request.RoomKey) != 64 {
			t.Fatalf("identifier not 64-hex opaque: %#v", request)
		}
		if request.Role != livekitapp.RoleSpeaker {
			t.Fatalf("role = %v, want speaker", request.Role)
		}
	}
	stored, err := repository.FindByID(ctx, issued.Call.ID())
	if err != nil || stored.Status() != domain.StatusRinging {
		t.Fatalf("stored = %#v, %v", stored, err)
	}

	// Outsider cannot end; participant ends.
	if err := service.End(ctx, issued.Call.ID(), "m-9"); err != application.ErrNotParticipant {
		t.Fatalf("outsider end = %v, want ErrNotParticipant", err)
	}
	if err := service.End(ctx, issued.Call.ID(), "m-2"); err != nil {
		t.Fatalf("end: %v", err)
	}
	stored, _ = repository.FindByID(ctx, issued.Call.ID())
	if stored.Status() != domain.StatusEnded || stored.EndedAt() == nil {
		t.Fatalf("ended = %#v", stored)
	}
}
