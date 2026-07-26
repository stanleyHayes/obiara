//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	sproutmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMutualConcurrentSproutAndThreeExchangeDoorway(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("seed_sprout_test")
	repository := sproutmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	alice, _ := domain.NewIntent("intent-a", "alice-key", "bob-key", "seed-key", "sprout-a", "fp-a", now)
	bob, _ := domain.NewIntent("intent-b", "bob-key", "alice-key", "seed-key", "sprout-b", "fp-b", now.Add(time.Millisecond))
	type sproutResult struct {
		doorway *domain.Doorway
		err     error
	}
	results := make(chan sproutResult, 2)
	var wg sync.WaitGroup
	for _, intent := range []domain.Intent{alice, bob} {
		wg.Add(1)
		go func(i domain.Intent) {
			defer wg.Done()
			doorway, _, e := repository.RecordIntent(ctx, i)
			results <- sproutResult{doorway, e}
		}(intent)
	}
	wg.Wait()
	close(results)
	var doorway *domain.Doorway
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.doorway != nil {
			copy := *result.doorway
			doorway = &copy
		}
	}
	if doorway == nil {
		t.Fatal("mutual sprout did not create private doorway")
	}
	// A unilateral intent for another pair remains intent-only.
	unilateral, _ := domain.NewIntent("intent-c", "carol-key", "dave-key", "other-seed", "sprout-c", "fp-c", now)
	if got, _, e := repository.RecordIntent(ctx, unilateral); e != nil || got != nil {
		t.Fatalf("unilateral doorway=%#v err=%v", got, e)
	}

	current, err := repository.FindDoorway(ctx, doorway.ID())
	if err != nil {
		t.Fatal(err)
	}
	firstActor := current.NextActorKey()
	otherActor := current.Participants()[0]
	if otherActor == firstActor {
		otherActor = current.Participants()[1]
	}
	firstA, _, _ := current.Exchange(firstActor, "message-1a", "exchange-1a", "fp-1a", now)
	firstB, _, _ := current.Exchange(firstActor, "message-1b", "exchange-1b", "fp-1b", now)
	type exchangeResult struct {
		doorway domain.Doorway
		err     error
	}
	exchangeResults := make(chan exchangeResult, 2)
	for _, candidate := range []domain.Doorway{firstA, firstB} {
		wg.Add(1)
		go func(next domain.Doorway) {
			defer wg.Done()
			stored, _, e := repository.AppendExchange(ctx, next, current.Revision())
			exchangeResults <- exchangeResult{stored, e}
		}(candidate)
	}
	wg.Wait()
	close(exchangeResults)
	successes, conflicts := 0, 0
	for result := range exchangeResults {
		if result.err == nil {
			successes++
			current = result.doorway
		} else if errors.Is(result.err, application.ErrConcurrentChange) {
			conflicts++
		} else {
			t.Fatal(result.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflict=%d", successes, conflicts)
	}
	second, _, err := current.Exchange(otherActor, "message-2", "exchange-2", "fp-2", now)
	if err != nil {
		t.Fatal(err)
	}
	current, _, err = repository.AppendExchange(ctx, second, current.Revision())
	if err != nil {
		t.Fatal(err)
	}
	third, _, err := current.Exchange(firstActor, "message-3", "exchange-3", "fp-3", now)
	if err != nil {
		t.Fatal(err)
	}
	current, _, err = repository.AppendExchange(ctx, third, current.Revision())
	if err != nil {
		t.Fatal(err)
	}
	if !current.Sealed() || len(current.Exchanges()) != 3 {
		t.Fatalf("doorway not exactly sealed: %#v", current)
	}
	replayed, replay, err := repository.AppendExchange(ctx, current, current.Revision()-1)
	if err != nil || !replay || len(replayed.Exchanges()) != 3 {
		t.Fatalf("replay=%v err=%v", replay, err)
	}
	var document bson.M
	if err = database.Collection("seed_sprout_events").FindOne(ctx, bson.M{"doorwayId": doorway.ID()}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"raw-alice", "raw-bob", "public", "room"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("leaked %q in %s", forbidden, raw)
		}
	}
}
