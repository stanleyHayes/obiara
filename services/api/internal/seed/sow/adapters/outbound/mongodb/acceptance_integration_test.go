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
	allowanceprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	sowmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestConcurrentAcceptanceSpendsOnceAndReplayIsFree(t *testing.T) {
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
	sep := "?"
	if strings.Contains(uri, "?") {
		sep = "&"
	}
	client, err := apimongo.Connect(ctx, uri+sep+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("seed_sow_test")
	allowances := allowancemongo.NewRepository(database)
	if err = allowances.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	acceptance := sowmongo.NewAcceptance(database)
	if err = acceptance.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	secret := []byte(strings.Repeat("k", 32))
	keyer, _ := allowanceprivacy.NewKeyer(secret)
	actorKey, _ := keyer.Key("raw-actor")
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	ledger, _ := domain.Issue(actorKey, 5, now, now, "issue", "issue-fingerprint")
	if err = allowances.Create(ctx, ledger); err != nil {
		t.Fatal(err)
	}
	first, _ := sowdomain.Accept("sow-a", actorKey, "body a", []sowdomain.Media{{Key: "media-a", ScreeningKey: "screen-a"}}, "command-a", "fingerprint-a", 4, now)
	second, _ := sowdomain.Accept("sow-b", actorKey, "body b", []sowdomain.Media{{Key: "media-b", ScreeningKey: "screen-b"}}, "command-b", "fingerprint-b", 4, now)
	type result struct {
		sow      sowdomain.Sow
		replayed bool
		err      error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for _, candidate := range []sowdomain.Sow{first, second} {
		wg.Add(1)
		go func(s sowdomain.Sow) {
			defer wg.Done()
			got, replayed, e := acceptance.Accept(ctx, s)
			results <- result{got, replayed, e}
		}(candidate)
	}
	wg.Wait()
	close(results)
	var winner sowdomain.Sow
	successes, insufficient := 0, 0
	for result := range results {
		if result.err == nil {
			successes++
			winner = result.sow
		} else if errors.Is(result.err, application.ErrInsufficientAllowance) {
			insufficient++
		} else {
			t.Fatal(result.err)
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Fatalf("success=%d insufficient=%d", successes, insufficient)
	}
	updated, err := allowances.Find(ctx, actorKey)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Balance() != 1 {
		t.Fatalf("balance=%d", updated.Balance())
	}
	if _, replayed, err := acceptance.Accept(ctx, winner); err != nil || !replayed {
		t.Fatalf("replay=%v err=%v", replayed, err)
	}
	updated, err = allowances.Find(ctx, actorKey)
	if err != nil || updated.Balance() != 1 {
		t.Fatalf("replay charged: balance=%d err=%v", updated.Balance(), err)
	}
	conflict := winner
	conflict.Fingerprint = "changed"
	if _, _, err = acceptance.Accept(ctx, conflict); !errors.Is(err, sowdomain.ErrCommandMismatch) {
		t.Fatalf("mismatch=%v", err)
	}
	var stored bson.M
	if err = database.Collection("seed_sow_events").FindOne(ctx, bson.M{"sowId": winner.ID}).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(stored, false, false)
	for _, forbidden := range []string{"raw-actor", "payment", "price", "boost", "purchase", "currency"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy/bypass leak %q in %s", forbidden, raw)
		}
	}
}
