//go:build integration

package mongodb

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryBoundariesReplayAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := mongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, container)
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()
	repository := NewRepository(client.Database("obiara_pace"))
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	key := func(value int) string { return fmt.Sprintf("%064x", value) }
	now := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	pace, _ := domain.New("pace_room_test", []string{key(1), key(2)}, "command_start", key(1), now)
	if err = repository.Create(ctx, pace); err != nil {
		t.Fatal(err)
	}
	rested, _ := pace.Advance(domain.Command{ID: "command_rest", ExpectedRevision: 1, At: pace.ResponseAt()})
	if err = repository.Save(ctx, rested, 1, "command_rest"); err != nil {
		t.Fatal(err)
	}
	loaded, err := repository.Find(ctx, pace.ID())
	if err != nil || loaded.Status() != domain.StatusResting || loaded.ArchiveAt().IsZero() {
		t.Fatalf("loaded = %s, %v", loaded.Status(), err)
	}
	raw, _ := repository.collection.FindOne(ctx, map[string]any{"_id": pace.ID()}).Raw()
	if strings.Contains(string(raw), "member-raw") || strings.Contains(string(raw), "room-raw") {
		t.Fatal("raw identifiers persisted")
	}
}
