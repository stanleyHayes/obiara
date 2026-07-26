//go:build integration

package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"strings"
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
	// A single-node replica set advertises its container-internal address;
	// directConnection keeps the driver on the host-mapped port while still
	// allowing transactions against the replica set. ConnectionString
	// already carries ?replicaSet=..., so append with the right separator.
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return ctx, client
}

func TestRunnerAppliesMigrationsOnce(t *testing.T) {
	ctx, client := startMongo(t)
	database := client.Database("obiara_migrate_test")

	migrations := []Migration{
		{ID: "0001_widgets_unique_name", Apply: func(ctx context.Context, db *mongo.Database) error {
			_, err := db.Collection("widgets").Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys:    bson.D{{Key: "name", Value: 1}},
				Options: options.Index().SetName("widgets_name_unique").SetUnique(true),
			})
			return err
		}},
		{ID: "0002_seed_widget", Apply: func(ctx context.Context, db *mongo.Database) error {
			_, err := db.Collection("widgets").InsertOne(ctx, bson.M{"_id": "w1", "name": "akoben"})
			return err
		}},
	}

	runner := NewRunner(database, time.Now)
	if err := runner.Run(ctx, migrations); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if err := runner.Run(ctx, migrations); err != nil {
		t.Fatalf("second Run must be a no-op: %v", err)
	}

	count, err := database.Collection("schema_migrations").CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("schema_migrations count = %d, want 2", count)
	}

	// The unique index from migration 0001 must actually reject duplicates.
	_, err = database.Collection("widgets").InsertOne(ctx, bson.M{"_id": "w2", "name": "akoben"})
	if !IsDuplicateKey(err) {
		t.Fatalf("duplicate insert error = %v, want duplicate key", err)
	}
}

func TestWithTransactionCommitAndAbort(t *testing.T) {
	ctx, client := startMongo(t)
	database := client.Database("obiara_tx_test")
	ledger := database.Collection("ledger_entries")

	// Abort path: the write must not survive a returned error.
	abortErr := errors.New("invariant violated")
	err := WithTransaction(ctx, client, func(sc context.Context) error {
		if _, err := ledger.InsertOne(sc, bson.M{"_id": "aborted", "amount": 50}); err != nil {
			return err
		}
		return abortErr
	})
	if !errors.Is(err, abortErr) && err == nil {
		t.Fatalf("abort transaction error = %v, want wrapped %v", err, abortErr)
	}
	if count, _ := ledger.CountDocuments(ctx, bson.M{"_id": "aborted"}); count != 0 {
		t.Fatalf("aborted write survived; count = %d", count)
	}

	// Commit path.
	err = WithTransaction(ctx, client, func(sc context.Context) error {
		_, err := ledger.InsertOne(sc, bson.M{"_id": "committed", "amount": 50})
		return err
	})
	if err != nil {
		t.Fatalf("commit transaction: %v", err)
	}
	if count, _ := ledger.CountDocuments(ctx, bson.M{"_id": "committed"}); count != 1 {
		t.Fatalf("committed write missing; count = %d", count)
	}
}
