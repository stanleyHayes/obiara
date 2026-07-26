//go:build integration

package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestRepositorySupportsOnlyBoundedOutgoingAndReplaySafeWrites(t *testing.T) {
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
	repository := NewRepository(client.Database("obiara_trust_test"))
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	first := integrationEdge(t, "edge-ab", "a", "b", "create-global", base)
	second := integrationEdge(t, "edge-bc", "b", "c", "create-bc", base)
	for _, edge := range []domain.Edge{first, second} {
		if err := repository.Save(ctx, edge, 0, edge.Commands()[0].ID()); err != nil {
			t.Fatalf("save %s: %v", edge.ID(), err)
		}
	}
	outgoing, err := repository.Outgoing(ctx, []string{"a"})
	if err != nil || len(outgoing) != 1 || outgoing[0].TargetID() != "b" {
		t.Fatalf("outgoing = %#v, error = %v", outgoing, err)
	}
	if _, err := repository.Outgoing(ctx, make([]string, domain.MaxProjectionNodes+1)); !errors.Is(err, domain.ErrProjectionBounds) {
		t.Fatalf("unbounded read error = %v, want %v", err, domain.ErrProjectionBounds)
	}

	revoke := domain.Command{
		ID: "revoke-ab", ExpectedRevision: 1, ActorRef: "actor-1",
		Kind: "edge.revoke", Payload: "edge-ab", RecordedAt: base.Add(time.Minute),
	}
	revoked, err := first.Revoke(revoke)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, revoked, 1, revoke.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	loaded, err := repository.Find(ctx, "edge-ab")
	if err != nil || loaded.RevokedAt() == nil || loaded.ProvenanceRef() != "provenance-edge-ab" {
		t.Fatalf("loaded = %#v, error = %v", loaded, err)
	}
	if err := repository.Save(ctx, revoked, 1, revoke.ID); !errors.Is(err, application.ErrCommandAlreadyApplied) {
		t.Fatalf("replay error = %v, want %v", err, application.ErrCommandAlreadyApplied)
	}

	duplicate := integrationEdge(t, "edge-other", "x", "y", "create-global", base)
	if err := repository.Save(ctx, duplicate, 0, "create-global"); !errors.Is(err, application.ErrCommandAlreadyApplied) {
		t.Fatalf("global command replay = %v, want %v", err, application.ErrCommandAlreadyApplied)
	}
}

func integrationEdge(t *testing.T, id, source, target, commandID string, now time.Time) domain.Edge {
	t.Helper()
	edge, err := domain.Create(domain.Params{
		ID: id, SourceID: source, TargetID: target, Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-" + id, ConsentRef: "consent-" + id,
		Visibility: domain.VisibilityConsentedPath, CreatedAt: now,
	}, domain.Command{
		ID: commandID, ActorRef: "actor-1", Kind: "edge.create",
		Payload: id, RecordedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return edge
}
