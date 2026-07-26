//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	retentionmongo "github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/application"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const subject = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const hold = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const proof = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func TestHoldErasureCASIdempotencyTTLAndPrivacy(t *testing.T) {
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
	db := client.Database("retention_test")
	repo := retentionmongo.NewRepository(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 19, 0, 0, 0, time.UTC)
	p, _ := domain.NewPolicy("messages.metadata", "safety.audit", 1, 30*24*time.Hour, now.Add(-time.Hour))
	r, _ := domain.Create("retention:1", subject, p, domain.Command{ID: "create", At: now})
	if e = repo.Create(ctx, r); e != nil {
		t.Fatal(e)
	}
	held, _ := r.PlaceHold(hold, domain.Command{ID: "hold", ExpectedRevision: 1, At: now})
	requested, _ := r.RequestErasure(domain.Command{ID: "request", ExpectedRevision: 1, At: now})
	ch := make(chan error, 2)
	go func() { ch <- repo.Append(ctx, held, 1, "hold") }()
	go func() { ch <- repo.Append(ctx, requested, 1, "request") }()
	success, conflict := 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			success++
		} else if errors.Is(x, application.ErrConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	r, e = repo.Find(ctx, "retention:1")
	if e != nil {
		t.Fatal(e)
	}
	if !r.Hold().Active {
		r, e = r.PlaceHold(hold, domain.Command{ID: "hold:retry", ExpectedRevision: r.Revision(), At: now.Add(time.Second)})
		if e != nil {
			t.Fatal(e)
		}
		if e = repo.Append(ctx, r, r.Revision()-1, "hold:retry"); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = r.CompleteErasure(proof, domain.Command{ID: "complete:blocked", ExpectedRevision: r.Revision(), At: now.Add(2 * time.Second)}); !errors.Is(e, domain.ErrHeld) {
		t.Fatalf("hold bypassed: %v", e)
	}
	r, e = r.ReleaseHold(hold, domain.Command{ID: "release", ExpectedRevision: r.Revision(), At: now.Add(3 * time.Second)})
	if e != nil {
		t.Fatal(e)
	}
	if e = repo.Append(ctx, r, r.Revision()-1, "release"); e != nil {
		t.Fatal(e)
	}
	if r.Status() == domain.StatusRetained {
		r, e = r.RequestErasure(domain.Command{ID: "request:retry", ExpectedRevision: r.Revision(), At: now.Add(4 * time.Second)})
		if e != nil {
			t.Fatal(e)
		}
		if e = repo.Append(ctx, r, r.Revision()-1, "request:retry"); e != nil {
			t.Fatal(e)
		}
	}
	expected := r.Revision()
	r, e = r.CompleteErasure(proof, domain.Command{ID: "complete", ExpectedRevision: expected, At: now.Add(5 * time.Second)})
	if e != nil {
		t.Fatal(e)
	}
	if e = repo.Append(ctx, r, expected, "complete"); e != nil {
		t.Fatal(e)
	}
	if e = repo.Append(ctx, r, expected, "complete"); !errors.Is(e, application.ErrApplied) {
		t.Fatalf("idempotency=%v", e)
	}
	cur, e := db.Collection("compliance_retention_leases").Indexes().List(ctx)
	if e != nil {
		t.Fatal(e)
	}
	var indexes []bson.M
	if e = cur.All(ctx, &indexes); e != nil {
		t.Fatal(e)
	}
	ttl := false
	for _, idx := range indexes {
		if idx["name"] == "retention_coordination_lease_ttl" && idx["expireAfterSeconds"] != nil {
			ttl = true
		}
	}
	if !ttl {
		t.Fatalf("ttl index missing: %+v", indexes)
	}
	var docs []bson.M
	data, e := db.Collection("compliance_retention_records").Find(ctx, bson.M{})
	if e != nil {
		t.Fatal(e)
	}
	if e = data.All(ctx, &docs); e != nil {
		t.Fatal(e)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, bad := range []string{"member@example.invalid", "productcontent", "rawmember", "reversible", "crosspurpose", "silentdelete"} {
		if strings.Contains(value, bad) {
			t.Fatalf("leaked %q: %s", bad, value)
		}
	}
}
