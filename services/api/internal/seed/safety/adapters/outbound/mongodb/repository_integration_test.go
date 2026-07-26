//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	safetymongo "github.com/stanleyHayes/obiara/services/api/internal/seed/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func safetyKey(n int) string { return fmt.Sprintf("%064x", n) }

func TestConcurrentThrottleAndPseudonymousCareSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, _ := container.ConnectionString(ctx)
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("seed_safety_test")
	repository := safetymongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 10, 1, 0, 0, time.UTC)
	bucket, _ := domain.New(safetyKey(1), now)
	if err = repository.Create(ctx, bucket); err != nil {
		t.Fatal(err)
	}
	var allowed, stale int
	var mutex sync.Mutex
	var wait sync.WaitGroup
	for range 12 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			current, findErr := repository.Find(ctx, safetyKey(1))
			if findErr != nil {
				return
			}
			changed, decision, changeErr := current.Evaluate(domain.OperationSow, current.Revision, now)
			if changeErr != nil {
				return
			}
			saveErr := repository.Save(ctx, changed, current.Revision)
			mutex.Lock()
			defer mutex.Unlock()
			if saveErr == nil && decision.Allowed {
				allowed++
			} else if saveErr == domain.ErrStaleRevision {
				stale++
			}
		}()
	}
	wait.Wait()
	if allowed < 1 || allowed > 6 || stale < 1 {
		t.Fatalf("allowed=%d stale=%d", allowed, stale)
	}
	if err = repository.AppendCareSignal(ctx, application.CareSignal{ActorKey: safetyKey(1), Code: "repeated_seed_demand", WindowRevision: 9}); err != nil {
		t.Fatal(err)
	}
	if err = repository.AppendCareSignal(ctx, application.CareSignal{ActorKey: safetyKey(1), Code: "repeated_seed_demand", WindowRevision: 9}); err != nil {
		t.Fatal(err)
	}
	var raw bson.M
	if err = database.Collection("seed_care_signals").FindOne(ctx, bson.M{}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{"raw-member", "content", "accusation", "score", "globalGraph"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("care signal leak %q", forbidden)
		}
	}
}
