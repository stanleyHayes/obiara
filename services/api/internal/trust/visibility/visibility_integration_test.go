//go:build integration

package visibility

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	trustmongodb "github.com/stanleyHayes/obiara/services/api/internal/trust/adapters/outbound/mongodb"
	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestMongoProjectionRevalidatesWithdrawalBeforeDisclosure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
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
	if strings.Contains(uri, "?") {
		uri += "&directConnection=true"
	} else {
		uri += "?directConnection=true"
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	repository := trustmongodb.NewRepository(client.Database("obiara_trust_visibility_test"))
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	edge, err := domain.Create(domain.Params{
		ID: "edge-ab", SourceID: "member-a", TargetID: "member-b", Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-ab", ConsentRef: "consent-ab",
		Visibility: domain.VisibilityConsentedPath, CreatedAt: now,
	}, domain.Command{
		ID: "create-ab", ActorRef: "actor-1", Kind: "edge.create",
		Payload: "edge-ab", RecordedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, edge, 0, "create-ab"); err != nil {
		t.Fatal(err)
	}

	// The kernel sees active consent, then consent is withdrawn before the
	// explanation policy's mandatory second evaluation.
	consent := &withdrawAfterFirst{callsBeforeWithdrawal: 1}
	projector := trustapplication.NewService(
		repository, integrationProjectionAuthorizer{}, consent,
		func() time.Time { return now.Add(time.Hour) },
	)
	service := NewService(
		projector,
		NewDisclosurePolicy(repository, consent, integrationEndpoints{}, func() time.Time {
			return now.Add(time.Hour)
		}),
	)
	paths, err := service.Explain(ctx, Request{
		RequesterID: "member-a", RootID: "member-a", MaxDepth: 2, MaxNodes: 10,
	})
	if err != nil || len(paths) != 0 {
		t.Fatalf("withdrawn paths = %#v, error = %v", paths, err)
	}

	consent = &withdrawAfterFirst{callsBeforeWithdrawal: 100}
	projector = trustapplication.NewService(
		repository, integrationProjectionAuthorizer{}, consent,
		func() time.Time { return now.Add(time.Hour) },
	)
	service = NewService(
		projector,
		NewDisclosurePolicy(repository, consent, integrationEndpoints{}, func() time.Time {
			return now.Add(time.Hour)
		}),
	)
	paths, err = service.Explain(ctx, Request{
		RequesterID: "member-a", RootID: "member-a", MaxDepth: 2, MaxNodes: 10,
	})
	if err != nil || len(paths) != 1 || paths[0].TargetID != "member-b" {
		t.Fatalf("visible paths = %#v, error = %v", paths, err)
	}
}

type withdrawAfterFirst struct {
	mu                    sync.Mutex
	calls                 int
	callsBeforeWithdrawal int
}

func (consent *withdrawAfterFirst) Allows(context.Context, string, string) (bool, error) {
	consent.mu.Lock()
	defer consent.mu.Unlock()
	consent.calls++
	return consent.calls <= consent.callsBeforeWithdrawal, nil
}

type integrationProjectionAuthorizer struct{}

func (integrationProjectionAuthorizer) CanProject(context.Context, string, string) (bool, error) {
	return true, nil
}

type integrationEndpoints struct{}

func (integrationEndpoints) CanReveal(context.Context, string, string) (bool, error) {
	return true, nil
}
