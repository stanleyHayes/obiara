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
	hostmongodb "github.com/stanleyHayes/obiara/services/api/internal/host/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/host/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/host/application"
	"github.com/stanleyHayes/obiara/services/api/internal/host/domain"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "host:integration" }

type uncertainProvider struct{ calls int }

func (provider *uncertainProvider) Verify(context.Context, application.ProviderRequest) (application.ProviderResult, error) {
	provider.calls++
	return application.ProviderResult{Outcome: application.OutcomeUncertain, ProviderRef: "provider:uncertain"}, nil
}

type reviewQueue struct {
	tasks map[string]application.ReviewTask
}

func (queue *reviewQueue) Enqueue(_ context.Context, task application.ReviewTask) error {
	queue.tasks[task.ApplicationID] = task
	return nil
}
func (queue *reviewQueue) Complete(_ context.Context, id string) error {
	delete(queue.tasks, id)
	return nil
}

func TestHostApplicationManualVerificationAndIdempotency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker required): %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
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

	database := client.Database("obiara_host_test")
	repository := hostmongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	keyer, _ := privacy.NewHMACKeyer([]byte(strings.Repeat("k", 32)))
	provider := &uncertainProvider{}
	reviews := &reviewQueue{tasks: make(map[string]application.ReviewTask)}
	now := time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)
	service := application.NewService(repository, provider, reviews, keyer, fixedIDs{}, func() time.Time { return now })
	request := application.SubmitRequest{
		CommandID: "submission:integration", ApplicantID: "member:raw",
		ProofReference: "evidence:institution", InstitutionKind: domain.InstitutionUniversity,
		IssuerID: "university:raw", IssuedAt: now.Add(-24 * time.Hour),
		ExpiresAt: now.Add(180 * 24 * time.Hour),
	}
	submitted, err := service.Submit(ctx, request)
	if !errors.Is(err, application.ErrManualReviewRequired) ||
		submitted.Application.Status() != domain.StatusQueuedManual ||
		submitted.Application.Eligible(now) {
		t.Fatalf("submitted=%+v, err=%v", submitted, err)
	}
	if _, exists := reviews.tasks[submitted.Application.ID()]; !exists {
		t.Fatal("manual review task missing")
	}

	now = now.Add(time.Minute)
	approved, err := service.ManualDecision(
		ctx, submitted.Application.ID(), "decision:1", "reviewer:raw", true,
	)
	if err != nil || !approved.Application.Eligible(now) {
		t.Fatalf("approved=%+v, err=%v", approved, err)
	}
	if _, exists := reviews.tasks[submitted.Application.ID()]; exists {
		t.Fatal("completed manual task retained")
	}

	replayed, err := service.Submit(ctx, request)
	if err != nil || !replayed.Replayed || !replayed.Application.Eligible(now) ||
		provider.calls != 1 {
		t.Fatalf("replay=%+v, err=%v provider calls=%d", replayed, err, provider.calls)
	}

	var stored bson.M
	if err := database.Collection("host_applications").FindOne(
		ctx, bson.M{"_id": submitted.Application.ID()},
	).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(stored, false, false)
	for _, raw := range []string{"member:raw", "university:raw", "reviewer:raw"} {
		if strings.Contains(string(encoded), raw) {
			t.Fatalf("privacy-safe record leaked %q: %s", raw, encoded)
		}
	}
	audit, ok := stored["audit"].(bson.A)
	if !ok || len(audit) != 3 {
		t.Fatalf("audit=%T %#v, want submitted+queued+approved", stored["audit"], stored["audit"])
	}
}
