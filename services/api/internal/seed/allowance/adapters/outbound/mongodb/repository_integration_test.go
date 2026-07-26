//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	allowancemongo "github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestConcurrentRenewalSpendIsAtomicAndPrivate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
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
	defer client.Disconnect(context.Background())
	database := client.Database("seed_allowance_test")
	repository := allowancemongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	policy, _ := domain.NewWeekPolicy("Africa/Accra")
	sunday := time.Date(2026, 1, 11, 23, 59, 59, 0, time.UTC)
	initial, _ := domain.Issue("opaque-subject-key", 5, policy.Start(sunday), sunday, "issue", "issue-fp")
	if err = repository.Create(ctx, initial); err != nil {
		t.Fatal(err)
	}
	monday := sunday.Add(time.Second)
	first, _, _ := initial.Spend(4, policy.Start(monday), monday, "spend-a", "fp-a")
	second, _, _ := initial.Spend(4, policy.Start(monday), monday, "spend-b", "fp-b")

	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, candidate := range []domain.Ledger{first, second} {
		wg.Add(1)
		go func(next domain.Ledger) { defer wg.Done(); results <- repository.Save(ctx, next, initial.Version()) }(candidate)
	}
	wg.Wait()
	close(results)
	var successes, conflicts int
	for saveErr := range results {
		if saveErr == nil {
			successes++
		} else if errors.Is(saveErr, application.ErrConcurrentChange) {
			conflicts++
		} else {
			t.Fatal(saveErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	stored, err := repository.Find(ctx, "opaque-subject-key")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Balance() != 1 {
		t.Fatalf("atomic balance=%d", stored.Balance())
	}
	entries := stored.Entries()
	if len(entries) != 3 || entries[1].Kind != domain.EntryRenewal || entries[2].Kind != domain.EntrySpend {
		t.Fatalf("renewal/spend audit=%#v", entries)
	}
	var document bson.M
	if err = database.Collection("seed_allowance_entries").FindOne(ctx, bson.M{"kind": "spend"}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"raw-member-id", "money", "currency", "payment", "price", "purchase"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("forbidden field/value %q in %s", forbidden, raw)
		}
	}
}
