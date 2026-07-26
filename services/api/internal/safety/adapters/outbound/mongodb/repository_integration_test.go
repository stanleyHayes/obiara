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
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/application"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/domain"
)

const safetyIntegrationTimeout = 3 * time.Minute

func TestSafetyIntakeEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), safetyIntegrationTimeout)
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

	database := client.Database("obiara_safety_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	outboxStore := outbox.NewStore(database, time.Now)
	if err := outboxStore.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "rep_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewSafetyService(repository, repository, outboxStore, time.Now, ids)

	// File a fraud report: Tier A, persisted, queue event emitted.
	id, tier, err := service.File(ctx, "m-1", "m-2", domain.CategoryFraud, domain.SurfaceRoom, "room_1", "emergency money story")
	if err != nil {
		t.Fatalf("file: %v", err)
	}
	if tier != domain.TierA {
		t.Fatalf("tier = %s, want A", tier)
	}
	stored, err := repository.FindByID(ctx, id)
	if err != nil || stored.ReporterID() != "m-1" || stored.Category() != domain.CategoryFraud {
		t.Fatalf("stored = %#v, %v", stored, err)
	}
	eventCount, err := database.Collection("outbox").CountDocuments(ctx, bson.M{"eventType": "safety.report_filed"})
	if err != nil || eventCount != 1 {
		t.Fatalf("queue events = %d, want 1", eventCount)
	}

	// Self-report is rejected without persistence.
	if _, _, err := service.File(ctx, "m-1", "m-1", domain.CategorySpam, domain.SurfaceProfile, "", ""); err != domain.ErrSelfReport {
		t.Fatalf("self report = %v, want rejected", err)
	}

	// Block lifecycle with uniqueness.
	if err := service.Block(ctx, "m-1", "m-2"); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := service.Block(ctx, "m-1", "m-2"); err != application.ErrBlockExists {
		t.Fatalf("re-block = %v, want ErrBlockExists", err)
	}
	exists, err := service.IsBlocked(ctx, "m-1", "m-2")
	if err != nil || !exists {
		t.Fatalf("exists = %v, %v", exists, err)
	}
	if err := service.Unblock(ctx, "m-1", "m-2"); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if err := service.Unblock(ctx, "m-1", "m-2"); err != application.ErrBlockNotFound {
		t.Fatalf("re-unblock = %v, want ErrBlockNotFound", err)
	}
}
