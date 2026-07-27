//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrentIdempotencyCASAndPrivacy(t *testing.T) {
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
	db := client.Database("diaspora_test")
	repo := adapter.New(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	checkout, _ := domain.Create(key(1), key(2), domain.Quote{SKUKey: key(3), Version: 1, Currency: domain.CurrencyUSD, AmountMinor: 1499, ValidUntil: now.Add(time.Hour)}, key(4), "prepare-1", now)
	ch := make(chan error, 2)
	go func() { ch <- repo.Create(ctx, checkout) }()
	go func() { ch <- repo.Create(ctx, checkout) }()
	saved, replayed := 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replayed++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replayed != 1 {
		t.Fatalf("%d %d", saved, replayed)
	}
	next, _ := checkout.Confirm(key(5), "confirm-1", true, now)
	ch = make(chan error, 2)
	go func() { ch <- repo.Save(ctx, next, 1, "confirm-1") }()
	go func() { ch <- repo.Save(ctx, next, 1, "confirm-1") }()
	saved, replayed = 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replayed++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replayed != 1 {
		t.Fatalf("%d %d", saved, replayed)
	}
	var raw bson.M
	if e = db.Collection("diaspora_checkouts").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"card", "paypal", "email", "phone", "address", "pan", "cvv", "momo", "ghs", "accountnumber", "membertransfer"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("privacy/isolation leak %q: %s", bad, encoded)
		}
	}
}
