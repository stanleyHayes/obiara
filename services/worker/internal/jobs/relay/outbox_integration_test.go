//go:build integration

package relay

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

const integrationTimeout = 3 * time.Minute

func TestRelayAndDeadLettersAgainstRealMongo(t *testing.T) {
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

	database := client.Database("obiara_worker_test")
	store := outbox.NewStore(database, time.Now)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure outbox indexes: %v", err)
	}

	// Append a record, run one relay pass with a capturing publisher, and
	// confirm the record is published and no longer pending.
	if err := store.Append(ctx, outbox.Record{
		ID:            "evt-relay-1",
		AggregateType: "member",
		AggregateID:   "member-1",
		EventType:     "member.registered",
		Payload:       []byte(`{"memberId":"member-1"}`),
		OccurredAt:    time.Now(),
	}); err != nil {
		t.Fatalf("append outbox record: %v", err)
	}

	published := make(chan outbox.Record, 1)
	publisher := publishFunc(func(_ context.Context, record outbox.Record) error {
		published <- record
		return nil
	})
	job := NewOutboxJob(store, publisher, 100, time.Second)
	if err := job.Run(ctx); err != nil {
		t.Fatalf("relay run: %v", err)
	}
	select {
	case record := <-published:
		if record.ID != "evt-relay-1" {
			t.Fatalf("published record = %q", record.ID)
		}
	default:
		t.Fatal("publisher never called")
	}
	if pending, _ := store.Pending(ctx, 10); len(pending) != 0 {
		t.Fatalf("record still pending after relay: %d", len(pending))
	}

	// Dead-letter persistence round trip.
	letters := mongodb.NewDeadLetterStore(database, time.Now)
	if err := letters.Record(ctx, application.DeadLetter{
		JobName:    "outbox.relay",
		Reason:     "provider down",
		Failures:   5,
		OccurredAt: time.Now(),
	}); err != nil {
		t.Fatalf("record dead letter: %v", err)
	}
	count, err := database.Collection("job_dead_letters").CountDocuments(ctx, bson.M{"jobName": "outbox.relay"})
	if err != nil || count != 1 {
		t.Fatalf("dead letters count = %d, %v; want 1", count, err)
	}
}

type publishFunc func(context.Context, outbox.Record) error

func (fn publishFunc) Publish(ctx context.Context, record outbox.Record) error {
	return fn(ctx, record)
}
