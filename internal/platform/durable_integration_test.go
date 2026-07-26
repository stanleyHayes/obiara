//go:build integration

package platform_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/platform/idempotency"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
)

const integrationTimeout = 3 * time.Minute

func startMongo(t *testing.T) (context.Context, *mongo.Client) {
	t.Helper()
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
	// ConnectionString already carries ?replicaSet=...; append with the
	// right separator. directConnection keeps the driver on the host-mapped
	// port of the single-node replica set.
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
	return ctx, client
}

func testRecord(id string) outbox.Record {
	return outbox.Record{
		ID:            id,
		AggregateType: "member",
		AggregateID:   "member-1",
		EventType:     "member.registered",
		Payload:       []byte(`{"memberId":"member-1"}`),
		OccurredAt:    time.Now(),
	}
}

func TestOutboxCommitsAndAbortsWithDomainChange(t *testing.T) {
	ctx, client := startMongo(t)
	database := client.Database("obiara_durable_test")
	store := outbox.NewStore(database, time.Now)
	members := database.Collection("members")

	// Abort: neither the domain write nor the outbox record may survive.
	abortErr := errors.New("domain invariant failed")
	err := apimongo.WithTransaction(ctx, client, func(sc context.Context) error {
		if _, err := members.InsertOne(sc, bson.M{"_id": "member-abort"}); err != nil {
			return err
		}
		if err := store.Append(sc, testRecord("evt-abort")); err != nil {
			return err
		}
		return abortErr
	})
	if err == nil {
		t.Fatal("transaction should have aborted")
	}
	if count, _ := members.CountDocuments(ctx, bson.M{"_id": "member-abort"}); count != 0 {
		t.Fatal("aborted domain write survived")
	}
	if pending, _ := store.Pending(ctx, 10); len(pending) != 0 {
		t.Fatalf("aborted outbox record survived: %d pending", len(pending))
	}

	// Commit: both survive atomically.
	err = apimongo.WithTransaction(ctx, client, func(sc context.Context) error {
		if _, err := members.InsertOne(sc, bson.M{"_id": "member-commit"}); err != nil {
			return err
		}
		return store.Append(sc, testRecord("evt-commit"))
	})
	if err != nil {
		t.Fatalf("commit transaction: %v", err)
	}
	pending, err := store.Pending(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].ID != "evt-commit" {
		t.Fatalf("pending = %#v", pending)
	}

	if err := store.MarkPublished(ctx, "evt-commit"); err != nil {
		t.Fatal(err)
	}
	if pending, _ := store.Pending(ctx, 10); len(pending) != 0 {
		t.Fatal("published record still pending")
	}
}

func TestInboxDeduplicatesRedelivery(t *testing.T) {
	ctx, client := startMongo(t)
	store := inbox.NewStore(client.Database("obiara_durable_test"), time.Now)

	first, err := store.AlreadyProcessed(ctx, "relay", "msg-1")
	if err != nil || first {
		t.Fatalf("first delivery = %v, %v; want false, nil", first, err)
	}
	for i := 0; i < 3; i++ {
		again, err := store.AlreadyProcessed(ctx, "relay", "msg-1")
		if err != nil || !again {
			t.Fatalf("redelivery %d = %v, %v; want true, nil", i, again, err)
		}
	}
	other, err := store.AlreadyProcessed(ctx, "other-consumer", "msg-1")
	if err != nil || other {
		t.Fatal("dedup is scoped per consumer")
	}
}

func TestIdempotencyClaimCompleteReplay(t *testing.T) {
	ctx, client := startMongo(t)
	store := idempotency.NewStore(client.Database("obiara_durable_test"), time.Now)

	claimed, err := store.Claim(ctx, "member.register", "key-1")
	if err != nil || !claimed {
		t.Fatalf("first claim = %v, %v; want true, nil", claimed, err)
	}
	replay, err := store.Claim(ctx, "member.register", "key-1")
	if err != nil || replay {
		t.Fatalf("replay claim = %v, %v; want false, nil", replay, err)
	}

	if err := store.Complete(ctx, "member.register", "key-1", "member-1"); err != nil {
		t.Fatal(err)
	}
	ref, err := store.ResultRef(ctx, "member.register", "key-1")
	if err != nil || ref != "member-1" {
		t.Fatalf("result ref = %q, %v", ref, err)
	}

	if err := store.Complete(ctx, "member.register", "never-claimed", "x"); err == nil {
		t.Fatal("completing an unclaimed key must fail")
	}
}
