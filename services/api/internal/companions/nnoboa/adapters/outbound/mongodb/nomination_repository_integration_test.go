//go:build integration

package mongodb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

func TestNominationRepositoryRoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Skipf("Docker unavailable: %v", err)
	}
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri + separator + "directConnection=true"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	repo, err := NewNominationRepository(ctx, client.Database("nnoboa_test"))
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	n, err := domain.NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != domain.StatusPending || got.KinName != "Auntie Efua" || got.Version != 1 {
		t.Fatalf("unexpected readback: %+v", got)
	}

	pending, err := repo.HasPendingForKin(ctx, n.MemberID, n.KinPhone)
	if err != nil || !pending {
		t.Fatalf("HasPendingForKin = %v, %v", pending, err)
	}

	if err := got.Consent(now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	after, err := repo.FindByID(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != domain.StatusConsented || after.RespondedAt == nil || after.Version != 2 {
		t.Fatalf("unexpected after consent: %+v", after)
	}

	// Optimistic concurrency: replaying the stale version must conflict.
	stale := got
	stale.Status = domain.StatusDeclined
	if err := repo.Update(ctx, stale); err == nil {
		t.Fatal("expected version conflict")
	}

	pending, err = repo.HasPendingForKin(ctx, n.MemberID, n.KinPhone)
	if err != nil || pending {
		t.Fatalf("HasPendingForKin after consent = %v, %v", pending, err)
	}

	// Listing orders latest first.
	second, err := domain.NewNomination(n.MemberID, "Uncle Kofi", "+233550000102", "uncle", now.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListByMember(ctx, n.MemberID)
	if err != nil {
		t.Fatalf("ListByMember: %v", err)
	}
	if len(list) != 2 || list[0].ID != second.ID {
		t.Fatalf("unexpected list: %+v", list)
	}

	if _, err := repo.FindByID(ctx, "nom_missing"); err != application.ErrNominationNotFound {
		t.Fatalf("want ErrNominationNotFound, got %v", err)
	}
}
