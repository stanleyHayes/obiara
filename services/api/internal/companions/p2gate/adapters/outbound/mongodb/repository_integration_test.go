package mongodb

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/domain"
	"github.com/testcontainers/testcontainers-go"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryConcurrentIdempotencyExpiryAndPrivacy(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
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
	repo := New(client.Database("p2_gate_test"))
	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	proposal, err := domain.Propose(
		"proposal-001", "command-0001", "courtship-001", "reviewer-001",
		"tokenref-001", "watermark-001", 1, []domain.PackItem{domain.IdentityCard},
		domain.GateConsent{
			CourtshipRef: "courtship-001", PackVersion: 1,
			ConsentedItems: []domain.PackItem{domain.IdentityCard},
			PartyAApproved: true, PartyBApproved: true, Current: true,
		}, now,
	)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 12
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- repo.Create(ctx, proposal)
		}()
	}
	wg.Wait()
	close(results)
	var created, applied int
	for result := range results {
		switch {
		case result == nil:
			created++
		case errors.Is(result, application.ErrApplied):
			applied++
		default:
			t.Fatalf("unexpected concurrent result: %v", result)
		}
	}
	if created != 1 || applied != workers-1 {
		t.Fatalf("created=%d applied=%d", created, applied)
	}

	var raw bson.Raw
	if err := repo.collection.FindOne(ctx, bson.M{"commandId": proposal.CommandID}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	stored := strings.ToLower(raw.String())
	for _, forbidden := range []string{"phone", "contact", "content", "question", "+233", "whatsapp"} {
		if strings.Contains(stored, forbidden) {
			t.Fatalf("forbidden raw BSON material %q: %s", forbidden, stored)
		}
	}
	var indexes []bson.M
	cursor, err := repo.collection.Indexes().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Fatal(err)
	}
	foundTTL := false
	for _, index := range indexes {
		if index["name"] == "p2_gate_expiry" {
			foundTTL = true
			if seconds, ok := index["expireAfterSeconds"].(int32); !ok || seconds != 0 {
				t.Fatalf("unexpected TTL index: %+v", index)
			}
		}
	}
	if !foundTTL {
		t.Fatal("missing expiry index")
	}
}
