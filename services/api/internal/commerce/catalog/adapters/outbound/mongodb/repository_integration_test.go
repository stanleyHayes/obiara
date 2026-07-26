//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	catalogmongo "github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const title = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestVersionPublishedPriceAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13")
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, e := apimongo.Connect(ctx, uri)
	if e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("catalog_test")
	repo := catalogmongo.NewRepository(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 21, 0, 0, 0, time.UTC)
	p1, _ := domain.NewPrice(domain.CurrencyGHS, 2500)
	v1, e := domain.Create("sku:1", "event.entry", title, 1, domain.KindEventTicket, p1, domain.Command{ID: "create:1", At: now})
	if e != nil || repo.Create(ctx, v1) != nil {
		t.Fatal(e)
	}
	v1, e = v1.Publish(domain.Command{ID: "publish:1", ExpectedRevision: 1, At: now})
	if e != nil || repo.Append(ctx, v1, 1, "publish:1") != nil {
		t.Fatal(e)
	}
	stored, e := repo.Find(ctx, "event.entry", 1)
	if e != nil || stored.Price() != p1 || stored.Status() != domain.StatusPublished {
		t.Fatalf("stored=%+v err=%v", stored, e)
	}
	p2, _ := domain.NewPrice(domain.CurrencyGHS, 3000)
	v2, e := domain.NextVersion(stored, "sku:2", p2, domain.Command{ID: "create:2", At: now})
	if e != nil || repo.Create(ctx, v2) != nil {
		t.Fatal(e)
	}
	latest, e := repo.FindLatest(ctx, "event.entry")
	if e != nil || latest.Version() != 2 || latest.Price() != p2 {
		t.Fatalf("latest=%+v err=%v", latest, e)
	}
	duplicate, _ := domain.Create("sku:duplicate", "event.entry", title, 2, domain.KindEventTicket, p2, domain.Command{ID: "duplicate", At: now})
	if e = repo.Create(ctx, duplicate); !errors.Is(e, application.ErrConflict) {
		t.Fatalf("version uniqueness=%v", e)
	}
	_, e = db.Collection("commerce_catalog_skus").InsertOne(ctx, bson.M{"_id": "bad:1", "skuKey": "bad.item", "titleRef": title, "version": uint64(1), "kind": "rank", "price": bson.M{"currency": "GHS", "minor": 100}, "status": "draft", "revision": uint64(1), "events": bson.A{bson.M{"sequence": uint64(1), "commandId": "bad", "action": "create", "at": now}}, "commands": bson.A{bson.M{"id": "bad", "fingerprint": title, "revision": uint64(1)}}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = repo.Find(ctx, "bad.item", 1); !errors.Is(e, domain.ErrInvalid) {
		t.Fatalf("forbidden read=%v", e)
	}
	var docs []bson.M
	cur, e := db.Collection("commerce_catalog_skus").Find(ctx, bson.M{"skuKey": "event.entry"})
	if e != nil {
		t.Fatal(e)
	}
	if e = cur.All(ctx, &docs); e != nil {
		t.Fatal(e)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, bad := range []string{"seed", "approach", "visibility", "rank", "matchingadvantage", "matching_advantage", "suban", "trust", "urgency", "membertransfer", "member_transfer", "hiddenfee", "checkout", "provider", "profile"} {
		if strings.Contains(value, bad) {
			t.Fatalf("leaked %q: %s", bad, value)
		}
	}
}
