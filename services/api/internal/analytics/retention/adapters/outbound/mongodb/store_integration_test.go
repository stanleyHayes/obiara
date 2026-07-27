//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

type catalog struct{ p domain.Policy }

func (c catalog) Current(context.Context) (domain.Policy, error) { return c.p, nil }

type clock struct{ at time.Time }

func (c clock) Now() time.Time { return c.at }

type ids struct{ n atomic.Uint64 }

func (i *ids) NewID() string { return key(int(i.n.Add(1) + 10)) }
func TestLiveMongoLeaseTransactionsIdempotencyPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, e := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("analytics_retention_test")
	store := adapter.New(db)
	if e = store.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 2, 0, 0, 0, time.UTC)
	original100 := key(1)
	original13 := key(2)
	docs := []any{bson.M{"name": "epono.pod_heard", "subjectRef": key(3), "occurredAt": now.Add(-10 * 24 * time.Hour)}, bson.M{"name": "epono.pod_heard", "subjectRef": original100, "occurredAt": now.Add(-100 * 24 * time.Hour)}, bson.M{"name": "epono.seed_sown", "subjectRef": original13, "occurredAt": now.AddDate(0, -13, 0)}}
	if _, e = db.Collection("analytics_events").InsertMany(ctx, docs); e != nil {
		t.Fatal(e)
	}
	policy, _ := domain.NewPolicy(domain.PolicySpec{ID: "analytics.retention", ReviewID: "privacy.review", ReviewerKey: key(9), Version: 1, PseudonymKeyVersion: 1, ReviewedAt: now, BatchSize: 10})
	idSource := &ids{}
	pseudonymizer := privacy.New(map[uint64][]byte{1: []byte("01234567890123456789012345678901")})
	job := application.New(catalog{policy}, store, pseudonymizer, idSource, clock{now})
	ch := make(chan error, 2)
	go func() { _, x := job.Run(ctx); ch <- x }()
	go func() { _, x := job.Run(ctx); ch <- x }()
	for range 2 {
		if e = <-ch; e != nil {
			t.Fatal(e)
		}
	}
	if count, _ := db.Collection("analytics_events").CountDocuments(ctx, bson.M{}); count != 2 {
		t.Fatalf("events=%d", count)
	}
	if count, _ := db.Collection("analytics_monthly_aggregates").CountDocuments(ctx, bson.M{"name": "epono.seed_sown", "count": 1}); count != 1 {
		t.Fatalf("aggregate=%d", count)
	}
	if count, _ := db.Collection("analytics_retention_receipts").CountDocuments(ctx, bson.M{}); count != 2 {
		t.Fatalf("receipts=%d", count)
	}
	var raw []bson.M
	cursor, e := db.Collection("analytics_events").Find(ctx, bson.M{})
	if e != nil {
		t.Fatal(e)
	}
	if e = cursor.All(ctx, &raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	text := string(encoded)
	for _, bad := range []string{original100, original13, "memberId", "content", "message", "voice", "email", "phone", "rank", "score"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(bad)) {
			t.Fatalf("privacy leak %q: %s", bad, text)
		}
	}
	if _, e = job.Run(ctx); e != nil {
		t.Fatal(e)
	}
	if count, _ := db.Collection("analytics_monthly_aggregates").CountDocuments(ctx, bson.M{"name": "epono.seed_sown", "count": 1}); count != 1 {
		t.Fatal("retry double counted")
	}
}
