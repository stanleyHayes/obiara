//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrencyIdempotencyPrivacy(t *testing.T) {
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
	db := client.Database("escrow_test")
	r := adapter.New(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	x, _ := domain.Fund(key(1), key(2), 1000, domain.Terms{ID: "terms.1", Version: 1, Milestones: []domain.MilestoneTerm{{ID: "one", GrossPesewas: 1000, FeePesewas: 100}}}, "fund-1", now)
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, x) }()
	go func() { ch <- r.Create(ctx, x) }()
	saved, replay := 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("%d %d", saved, replay)
	}
	loaded, e := r.Find(ctx, key(1))
	if e != nil {
		t.Fatal(e)
	}
	next, _ := loaded.AddEvidence("one", domain.DeliveryEvidence, key(3), "delivery-1", now)
	ch = make(chan error, 2)
	go func() { ch <- r.Save(ctx, next, loaded.Revision(), "delivery-1") }()
	go func() { ch <- r.Save(ctx, next, loaded.Revision(), "delivery-1") }()
	saved, replay = 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("save %d %d", saved, replay)
	}
	var raw bson.M
	if e = db.Collection("commerce_escrows").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "card", "accountnumber", "memberid", "provider", "rawpayment", "hiddenfee"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
