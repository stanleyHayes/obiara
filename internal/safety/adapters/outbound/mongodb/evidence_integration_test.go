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
	"github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

const evidenceIntegrationTimeout = 3 * time.Minute

func TestEvidenceAccessEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), evidenceIntegrationTimeout)
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

	database := client.Database("obiara_evidence_test")
	reportRepository := mongodb.NewRepository(database)
	if err := reportRepository.EnsureIndexes(ctx); err != nil {
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
			return "id_" + strings.Repeat("e", counter)
		}
	}()
	safetyService := application.NewSafetyService(reportRepository, reportRepository, outboxStore, time.Now, ids)
	caseRepository := mongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureCaseIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	// A report whose reason carries third-party identifiers.
	reportID, _, err := safetyService.File(ctx, "m-1", "m-2", domain.CategoryHarassment, domain.SurfaceRoom, "room_1",
		"told me to call his friend on +233 55 000 0101 or mail fixer@example.com")
	if err != nil {
		t.Fatalf("file: %v", err)
	}

	report, err := reportRepository.FindByID(ctx, reportID)
	if err != nil {
		t.Fatalf("find report: %v", err)
	}
	caseService := application.NewCaseService(caseRepository, time.Now, ids)
	safetyCase, err := caseService.Open(ctx, report)
	if err != nil {
		t.Fatalf("open case: %v", err)
	}
	evidence := application.NewEvidenceService(reportRepository, caseRepository, mongodb.NewAccessAuditStore(database), time.Now, ids)
	bundle, err := evidence.View(ctx, safetyCase.ID(), "agent-1", domain.PurposeTriage)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if strings.Contains(bundle.Description, "0101") || strings.Contains(bundle.Description, "fixer@example.com") {
		t.Fatalf("bundle leaked identifiers: %q", bundle.Description)
	}
	if bundle.SubjectID != "m-2" {
		t.Fatalf("bundle = %#v", bundle)
	}

	// Every access is audited — twice viewed, two records.
	if _, err := evidence.View(ctx, safetyCase.ID(), "agent-1", domain.PurposeAppeal); err != nil {
		t.Fatal(err)
	}
	auditCount, err := database.Collection("evidence_access_log").CountDocuments(ctx, bson.M{"caseId": safetyCase.ID()})
	if err != nil || auditCount != 2 {
		t.Fatalf("audit records = %d, want 2", auditCount)
	}
	var record struct {
		AgentID string `bson:"agentId"`
		Purpose string `bson:"purpose"`
	}
	if err := database.Collection("evidence_access_log").FindOne(ctx, bson.M{"caseId": safetyCase.ID(), "purpose": "appeal"}).Decode(&record); err != nil {
		t.Fatal(err)
	}
	if record.AgentID != "agent-1" {
		t.Fatalf("audit record = %#v", record)
	}

	// Curiosity is not a purpose: no view, no audit.
	if _, err := evidence.View(ctx, safetyCase.ID(), "agent-2", domain.Purpose("curiosity")); err != domain.ErrInvalidPurpose {
		t.Fatalf("curiosity = %v, want rejected", err)
	}
	auditCount, _ = database.Collection("evidence_access_log").CountDocuments(ctx, bson.M{"caseId": safetyCase.ID()})
	if auditCount != 2 {
		t.Fatal("rejected access was audited anyway (or accepted)")
	}
}
