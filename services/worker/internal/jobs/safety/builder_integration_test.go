//go:build integration

package safety_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
	safetyjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/safety"
)

const integrationTimeout = 3 * time.Minute

func TestCaseBuilderEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_cases_test")
	reportRepository := mongodb.NewRepository(database)
	caseRepository := mongodb.NewCaseRepository(database)
	if err := reportRepository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	if err := caseRepository.EnsureCaseIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	outboxStore := outbox.NewStore(database, time.Now)
	if err := outboxStore.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "id_" + strings.Repeat("c", counter)
		}
	}()
	safetyService := application.NewSafetyService(reportRepository, reportRepository, outboxStore, time.Now, ids)
	caseService := application.NewCaseService(caseRepository, time.Now, ids)

	// File two reports: fraud (Tier A → 8h SLA) and spam (Tier C → 72h).
	fraudID, _, err := safetyService.File(ctx, "m-1", "m-2", domain.CategoryFraud, domain.SurfaceRoom, "room_1", "money story")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := safetyService.File(ctx, "m-3", "m-2", domain.CategorySpam, domain.SurfaceCircle, "circle_1", "spam"); err != nil {
		t.Fatal(err)
	}

	builder := safetyjob.NewCaseBuilder(outboxStore, reportRepository, caseService, inbox.NewStore(database, time.Now))
	if err := builder.RunOnce(ctx, 50); err != nil {
		t.Fatalf("builder: %v", err)
	}

	// Both cases exist with tiered SLAs; replay is a no-op.
	if err := builder.RunOnce(ctx, 50); err != nil {
		t.Fatal(err)
	}
	caseCount, err := database.Collection("safety_cases").CountDocuments(ctx, bson.M{})
	if err != nil || caseCount != 2 {
		t.Fatalf("cases = %d, want 2 (replay-safe)", caseCount)
	}

	queued, err := caseService.NextQueued(ctx, domain.QueueTriage, 10)
	if err != nil || len(queued) != 2 {
		t.Fatalf("queue = %#v, %v", queued, err)
	}
	// Oldest SLA first: the Tier A case leads.
	if queued[0].Tier() != domain.TierA || queued[0].ReportID() != fraudID {
		t.Fatalf("queue order = %#v", queued)
	}
	if queued[0].SLADueAt().Sub(queued[0].CreatedAt()) != 8*time.Hour {
		t.Fatalf("tier A SLA = %v", queued[0].SLADueAt())
	}

	// Assign and resolve the Tier A case; breach count drops accordingly.
	if _, err := caseService.Assign(ctx, queued[0].ID(), "agent-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := caseService.Resolve(ctx, queued[0].ID(), "account banned", "agent-1"); err != nil {
		t.Fatal(err)
	}
	resolved, err := caseService.NextQueued(ctx, domain.QueueTriage, 10)
	if err != nil || len(resolved) != 1 {
		t.Fatalf("after resolve queue = %#v", resolved)
	}

	// Force an SLA breach on the remaining case and observe it.
	if _, err := database.Collection("safety_cases").UpdateOne(ctx,
		bson.M{"_id": resolved[0].ID()},
		bson.M{"$set": bson.M{"slaDueAt": time.Now().Add(-time.Hour)}}); err != nil {
		t.Fatal(err)
	}
	breached, err := caseService.BreachCount(ctx)
	if err != nil || breached != 1 {
		t.Fatalf("breached = %d, want 1", breached)
	}
}
