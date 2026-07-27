//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	reconmongo "github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const (
	providerKey  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	eventKey     = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	referenceKey = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestAppendOnlyConcurrencyCheckpointValidationAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, err := box.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("reconciliation_test")
	repo := reconmongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 27, 8, 0, 0, 0, time.UTC)
	a, _ := domain.NewFact("fact:a", providerKey, eventKey, referenceKey, "ledger:command", domain.CurrencyGHS, domain.StatusSettled, 500, now, now.Add(time.Second))
	b, _ := domain.NewFact("fact:b", providerKey, eventKey, referenceKey, "ledger:command", domain.CurrencyGHS, domain.StatusSettled, 500, now, now.Add(2*time.Second))
	ch := make(chan error, 2)
	go func() { ch <- repo.AppendFact(ctx, a) }()
	go func() { ch <- repo.AppendFact(ctx, b) }()
	ok, applied := 0, 0
	for range 2 {
		switch e := <-ch; {
		case e == nil:
			ok++
		case errors.Is(e, application.ErrApplied):
			applied++
		default:
			t.Fatal(e)
		}
	}
	if ok != 1 || applied != 1 {
		t.Fatalf("ok=%d applied=%d", ok, applied)
	}
	stored, err := repo.FindFactByEvent(ctx, eventKey)
	if err != nil || stored.Fingerprint() != a.Fingerprint() {
		t.Fatalf("fact=%+v err=%v", stored, err)
	}
	facts, err := repo.ListFactsForDay(ctx, "2026-07-27")
	if err != nil || len(facts) != 1 {
		t.Fatalf("facts=%d err=%v", len(facts), err)
	}

	decision := domain.Compare(stored, domain.LedgerProof{CommandID: "ledger:command", ReferenceKey: referenceKey, Currency: domain.CurrencyGHS, Minor: 500, Balanced: true}, true)
	audit, _ := domain.NewAudit("audit:a", stored, decision, now.Add(3*time.Second))
	if err = repo.AppendAudit(ctx, audit); err != nil {
		t.Fatal(err)
	}
	replay, _ := domain.NewAudit("audit:b", stored, decision, now.Add(4*time.Second))
	if err = repo.AppendAudit(ctx, replay); !errors.Is(err, application.ErrApplied) {
		t.Fatalf("audit replay=%v", err)
	}
	cp, _ := domain.NewCheckpoint("run:a", "2026-07-27", 1, 1, 0, now.Add(5*time.Second))
	if err = repo.AppendCheckpoint(ctx, cp); err != nil {
		t.Fatal(err)
	}
	cp2, _ := domain.NewCheckpoint("run:b", "2026-07-27", 1, 1, 0, now.Add(6*time.Second))
	if err = repo.AppendCheckpoint(ctx, cp2); !errors.Is(err, application.ErrApplied) {
		t.Fatalf("checkpoint replay=%v", err)
	}
	read, err := repo.FindCheckpoint(ctx, "2026-07-27")
	if err != nil || read.Fingerprint() != cp.Fingerprint() {
		t.Fatalf("checkpoint=%+v err=%v", read, err)
	}

	badEvent := strings.Repeat("d", 64)
	_, err = db.Collection("commerce_reconciliation_facts").InsertOne(ctx, bson.M{"_id": "bad", "providerKey": providerKey, "eventKey": badEvent, "referenceKey": referenceKey, "ledgerCommand": "ledger:bad", "fingerprint": referenceKey, "currency": "GHS", "status": "settled", "minor": -1, "occurredAt": now, "receivedAt": now, "occurredDay": "2026-07-27"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.FindFactByEvent(ctx, badEvent); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("malformed read=%v", err)
	}

	for _, collection := range []string{"commerce_reconciliation_facts", "commerce_reconciliation_audits", "commerce_reconciliation_checkpoints"} {
		cur, e := db.Collection(collection).Find(ctx, bson.M{})
		if e != nil {
			t.Fatal(e)
		}
		var docs []bson.M
		if e = cur.All(ctx, &docs); e != nil {
			t.Fatal(e)
		}
		raw, _ := bson.MarshalExtJSON(docs, false, false)
		value := strings.ToLower(string(raw))
		for _, bad := range []string{"raw-payment-reference", "authorization", "credential", "signature", "phone", "member@example", "tolerance", "balanceedit", "providernetwork"} {
			if strings.Contains(value, bad) {
				t.Fatalf("%s leaked %q: %s", collection, bad, value)
			}
		}
	}
}
