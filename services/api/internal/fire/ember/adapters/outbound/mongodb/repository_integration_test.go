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
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/domain"
)

const emberIntegrationTimeout = 3 * time.Minute

func TestEmbersEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), emberIntegrationTimeout)
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

	database := client.Database("obiara_ember_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	// Attendance: m-1 and m-2 attended, m-3 did not.
	for _, memberID := range []string{"m-1", "m-2"} {
		if _, err := database.Collection("fire_attendance").InsertOne(ctx, bson.M{
			"_id": "fire_1|" + memberID, "fireId": "fire_1", "memberId": memberID,
			"status": "going", "position": 0, "version": 1, "createdAt": time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
	}

	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "ember_" + strings.Repeat("x", counter)
		}
	}()
	service := application.NewEmberService(repository, repository, nil, time.Now, ids)

	// Non-attendee cannot give or receive.
	if _, err := service.Issue(ctx, "fire_1", "m-1", "m-3"); err != application.ErrNotCoAttendee {
		t.Fatalf("non-attendee recipient = %v, want ErrNotCoAttendee", err)
	}

	// One-way ember issues.
	oneWay, err := service.Issue(ctx, "fire_1", "m-1", "m-2")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if oneWay.Status() != domain.StatusIssued {
		t.Fatalf("status = %q", oneWay.Status())
	}

	// FR-402: one ember per attendee per fire.
	if _, err := service.Issue(ctx, "fire_1", "m-1", "m-2"); err != application.ErrEmberAlreadyGiven {
		t.Fatalf("second ember = %v, want ErrEmberAlreadyGiven", err)
	}

	// Only the recipient redeems.
	if _, err := service.Redeem(ctx, oneWay.ID(), "m-1"); err != application.ErrNotRecipient {
		t.Fatalf("giver redeem = %v, want ErrNotRecipient", err)
	}

	// Reverse ember flips both to mutual.
	reverse, err := service.Issue(ctx, "fire_1", "m-2", "m-1")
	if err != nil {
		t.Fatalf("reverse issue: %v", err)
	}
	if reverse.Status() != domain.StatusMutual {
		t.Fatalf("reverse status = %q, want mutual", reverse.Status())
	}
	reloaded, err := repository.FindByID(ctx, oneWay.ID())
	if err != nil || reloaded.Status() != domain.StatusMutual {
		t.Fatalf("forward status = %q, want mutual", reloaded.Status())
	}

	// Recipient redeems inside the window.
	redeemed, err := service.Redeem(ctx, oneWay.ID(), "m-2")
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if redeemed.Status() != domain.StatusRedeemed || redeemed.RedeemedAt() == nil {
		t.Fatalf("redeemed = %#v", redeemed)
	}
}
