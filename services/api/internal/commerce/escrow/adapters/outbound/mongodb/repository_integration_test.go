//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrencyIdempotencyPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, e := mongo.Connect(options.Client().ApplyURI(uri).SetDirect(true))
	if e != nil {
		t.Fatal(e)
	}
	if e = client.Ping(ctx, nil); e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("escrow_test")
	r, e := adapter.NewWithSettlementSecret(db, []byte("escrow-settlement-test-secret-32-bytes"))
	if e != nil {
		t.Fatal(e)
	}
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	x, _ := domain.Fund(key(1), key(2), key(3), key(4), 1000, domain.Terms{ID: "terms.1", Version: 1, Milestones: []domain.MilestoneTerm{{ID: "one", GrossPesewas: 1000, FeePesewas: 100}}}, "fund-1", now)
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

	delivered, e := r.Find(ctx, key(1))
	if e != nil {
		t.Fatal(e)
	}
	accepted, e := delivered.AddEvidence("one", domain.AcceptanceEvidence, key(6), "accept-1", now)
	if e != nil || r.Save(ctx, accepted, delivered.Revision(), "accept-1") != nil {
		t.Fatal("retain acceptance", e)
	}
	settleCurrent, e := r.Find(ctx, key(1))
	if e != nil {
		t.Fatal(e)
	}
	settled, statement, e := settleCurrent.Settle("one", key(7), "settle-1", now)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.SettleAudited(ctx, settled, settleCurrent.Revision(), "settle-1", "finance-1", statement); e != nil {
		t.Fatal(e)
	}
	if count, countErr := db.Collection("commerce_ledger_postings").CountDocuments(ctx, bson.M{"commandId": "settle-1"}); countErr != nil || count != 1 {
		t.Fatalf("ledger posting count=%d err=%v", count, countErr)
	}
	if count, countErr := db.Collection("admin_access").CountDocuments(ctx, bson.M{"action": "admin.escrow.settlement", "target": key(1)}); countErr != nil || count != 1 {
		t.Fatalf("settlement audit count=%d err=%v", count, countErr)
	}
	var posting struct {
		Lines []struct {
			Side  string `bson:"side"`
			Minor int64  `bson:"minor"`
		} `bson:"lines"`
	}
	if e = db.Collection("commerce_ledger_postings").FindOne(ctx, bson.M{"commandId": "settle-1"}).Decode(&posting); e != nil {
		t.Fatal(e)
	}
	var debits, credits int64
	for _, line := range posting.Lines {
		if line.Side == "debit" {
			debits += line.Minor
		} else if line.Side == "credit" {
			credits += line.Minor
		}
	}
	if debits != 1000 || credits != 1000 {
		t.Fatalf("unbalanced settlement debit=%d credit=%d", debits, credits)
	}
}
