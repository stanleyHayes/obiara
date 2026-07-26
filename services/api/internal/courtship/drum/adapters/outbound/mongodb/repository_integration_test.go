//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/domain"
)

func TestRepositoryAlternationReplayConcurrencyAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uri, "?") {
		uri += "&directConnection=true"
	} else {
		uri += "?directConnection=true"
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	database := client.Database("obiara_courtship_drum_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	stage, err := domain.Open("stage-1", []string{key(1), key(2)}, key(9), command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, stage); err != nil {
		t.Fatal(err)
	}
	var raw bson.M
	if err := database.Collection("courtship_drum_stages").FindOne(ctx, bson.M{"_id": "stage-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{"member-private", "voice-private", "text-private", "public", "popularity", "online", "presence"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("private stage leaked %q: %s", forbidden, encoded)
		}
	}

	if _, err := stage.Add(domain.MediumText, key(10), command("double", key(1), 1, now.Add(time.Second))); !errors.Is(err, domain.ErrNotTurn) {
		t.Fatalf("same actor error=%v", err)
	}
	first, _ := stage.Add(domain.MediumText, key(10), command("turn-a", key(2), 1, now.Add(time.Second)))
	racing, _ := stage.Add(domain.MediumVoice, key(11), command("turn-b", key(2), 1, now.Add(time.Second)))
	results := make(chan error, 2)
	go func() { results <- repository.Append(ctx, first, 1, "turn-a") }()
	go func() { results <- repository.Append(ctx, racing, 1, "turn-b") }()
	successes, conflicts := 0, 0
	for range 2 {
		switch err := <-results; {
		case err == nil:
			successes++
		case errors.Is(err, application.ErrOptimisticConflict):
			conflicts++
		default:
			t.Fatalf("append error=%v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	stored, err := repository.Find(ctx, "stage-1")
	if err != nil || stored.Revision() != 2 || stored.Beats()[0].Medium != domain.MediumVoice {
		t.Fatalf("stage revision=%d error=%v", stored.Revision(), err)
	}
	winner := stored.Commands()[1].ID
	if err := repository.Append(ctx, stored, 1, winner); !errors.Is(err, application.ErrCommandApplied) {
		t.Fatalf("replay error=%v", err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
