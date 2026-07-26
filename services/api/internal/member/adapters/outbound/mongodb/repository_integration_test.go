//go:build integration

package mongodb

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

func TestRepositoryPersistsMemberInMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), twoMinutes)
	defer cancel()

	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("MongoDB connection string: %v", err)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(connectionString))
	if err != nil {
		t.Fatalf("connect to MongoDB: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Errorf("disconnect MongoDB: %v", err)
		}
	})
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping MongoDB: %v", err)
	}

	repository := NewRepository(client.Database("obiara_test"))
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	createdAt := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	expected, err := domain.NewMember("member-1", "member@example.com", createdAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, expected); err != nil {
		t.Fatalf("create member: %v", err)
	}

	actual, err := repository.FindByID(ctx, expected.ID())
	if err != nil {
		t.Fatalf("find member: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("member mismatch: got %#v want %#v", actual, expected)
	}

	duplicate, err := domain.NewMember("member-2", expected.Email(), createdAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, duplicate); !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("duplicate email error = %v, want duplicate-key error", err)
	}

	_, err = repository.FindByID(ctx, "missing")
	if !errors.Is(err, ErrMemberNotFound) {
		t.Fatalf("missing member error = %v, want %v", err, ErrMemberNotFound)
	}
}

const twoMinutes = 2 * time.Minute
