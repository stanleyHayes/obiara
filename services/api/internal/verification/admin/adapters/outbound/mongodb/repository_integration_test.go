//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adminmongo "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

func TestAdminReviewIsRedactedAuditedAndIdempotent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_verification_admin_test")
	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("v", 32)))
	if err != nil {
		t.Fatal(err)
	}
	repository := adminmongo.NewRepository(database, keyer)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	rawAccount := "account:private"
	rawCard := "GHA-123456789-4"
	birth := time.Date(1995, 2, 3, 0, 0, 0, 0, time.UTC)
	submitted := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	if _, err := database.Collection("identity_verifications").InsertOne(ctx, bson.M{
		"_id": "IDV-1", "accountId": rawAccount, "cardNumber": rawCard,
		"status": "queued_manual", "reason": "provider uncertain",
		"dateOfBirth": birth, "version": int64(2), "createdAt": submitted,
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.NewService(repository, func() time.Time { return now })
	principal := application.Principal{
		ActorID: "verifier:private", MFAVerified: true,
		Scopes: []application.Scope{
			application.ScopeQueueRead, application.ScopeEvidenceRead, application.ScopeReview,
		},
	}

	queue, err := service.ListQueue(ctx, principal, 10)
	if err != nil || len(queue) != 1 || queue[0].SubjectRef == rawAccount {
		t.Fatalf("queue=%+v err=%v", queue, err)
	}
	evidence, err := service.OpenEvidence(ctx, principal, "IDV-1", "identity_review", "Provider result needs comparison", "corr-1")
	if err != nil || evidence.MaskedCard != "•••• 89-4" || evidence.AgeBand != "25_34" {
		t.Fatalf("evidence=%+v err=%v", evidence, err)
	}
	decision, err := service.Decide(ctx, principal, "IDV-1", application.OutcomeApprove, "Document and provider result are consistent", "command-1", "corr-1", 2)
	if err != nil || decision.Replayed || decision.Case.Status != "approved" || decision.Case.Version != 3 {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	replay, err := service.Decide(ctx, principal, "IDV-1", application.OutcomeApprove, "Document and provider result are consistent", "command-1", "corr-1", 2)
	if err != nil || !replay.Replayed {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	if _, err := service.Decide(ctx, principal, "IDV-1", application.OutcomeReject, "Document does not match provider result", "command-1", "corr-1", 2); !errors.Is(err, application.ErrIdempotencyConflict) {
		t.Fatalf("idempotency conflict=%v", err)
	}
	if count, _ := database.Collection("verification_admin_audit").CountDocuments(ctx, bson.M{"caseId": "IDV-1"}); count != 2 {
		t.Fatalf("audit count=%d, want evidence+decision", count)
	}
	var evidenceAudit bson.M
	if err := database.Collection("verification_admin_audit").FindOne(ctx, bson.M{"eventType": "evidence_access"}).Decode(&evidenceAudit); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(evidenceAudit, false, false)
	for _, forbidden := range []string{rawAccount, rawCard, birth.Format("2006-01-02")} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("audit leaked evidence %q: %s", forbidden, encoded)
		}
	}
}
