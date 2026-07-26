//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	ledgermongo "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const asset = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const revenue = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const reference = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func TestConcurrentPostingIdempotencyBalanceAndPrivacy(t *testing.T) {
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
	db := client.Database("ledger_test")
	repo := ledgermongo.NewRepository(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 22, 0, 0, 0, time.UTC)
	lines := []domain.Line{{AccountKey: asset, Class: domain.ClassAsset, Side: domain.SideDebit, Minor: 500}, {AccountKey: revenue, Class: domain.ClassRevenue, Side: domain.SideCredit, Minor: 500}}
	a, _ := domain.NewPosting("posting:a", "command:sale", reference, domain.PurposeSaleSettlement, domain.CurrencyGHS, lines, now)
	b, _ := domain.NewPosting("posting:b", "command:sale", reference, domain.PurposeSaleSettlement, domain.CurrencyGHS, lines, now.Add(time.Second))
	ch := make(chan error, 2)
	go func() { ch <- repo.Create(ctx, a) }()
	go func() { ch <- repo.Create(ctx, b) }()
	ok, applied := 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			ok++
		} else if errors.Is(x, application.ErrApplied) {
			applied++
		} else {
			t.Fatal(x)
		}
	}
	if ok != 1 || applied != 1 {
		t.Fatalf("ok=%d applied=%d", ok, applied)
	}
	booked, e := repo.ListLines(ctx, asset, domain.CurrencyGHS)
	if e != nil {
		t.Fatal(e)
	}
	balance, e := domain.RecomputeBalance(asset, domain.ClassAsset, domain.CurrencyGHS, booked)
	if e != nil || balance != 500 {
		t.Fatalf("balance=%d err=%v", balance, e)
	}
	stored, e := repo.FindByCommand(ctx, "command:sale")
	if e != nil || stored.Fingerprint() != a.Fingerprint() {
		t.Fatalf("stored=%+v err=%v", stored, e)
	}
	_, e = db.Collection("commerce_ledger_postings").InsertOne(ctx, bson.M{"_id": "bad", "commandId": "bad:single", "fingerprint": reference, "referenceKey": reference, "purpose": "sale_settlement", "currency": "GHS", "lines": bson.A{bson.M{"accountKey": asset, "class": "asset", "side": "debit", "minor": 100}}, "postedAt": now})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = repo.FindByCommand(ctx, "bad:single"); !errors.Is(e, domain.ErrInvalid) {
		t.Fatalf("single entry read=%v", e)
	}
	var docs []bson.M
	cur, e := db.Collection("commerce_ledger_postings").Find(ctx, bson.M{"commandId": "command:sale"})
	if e != nil {
		t.Fatal(e)
	}
	if e = cur.All(ctx, &docs); e != nil {
		t.Fatal(e)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, bad := range []string{"member@example", "paymenttoken", "cardnumber", "membertransfer", "member_transfer", "hiddenfee", "hidden_fee", "provider", "catalog", "admin", "cachedbalance"} {
		if strings.Contains(value, bad) {
			t.Fatalf("leaked %q: %s", bad, value)
		}
	}
}
