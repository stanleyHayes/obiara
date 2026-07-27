//go:build integration

package retention_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/retention"
)

const integrationTimeout = 3 * time.Minute

func TestRetentionRunnerEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_retention_test")
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)         // past 90d
	ancient := now.Add(-14 * 30 * 24 * time.Hour) // past 13mo

	// Analytics: one ancient (aggregate+delete), two old (strip subjectRef),
	// one fresh (untouched).
	for _, event := range []bson.M{
		{"name": "epono.pod_heard", "subjectRef": "ancient-ref", "occurredAt": ancient},
		{"name": "epono.pod_heard", "subjectRef": "old-ref-1", "occurredAt": old},
		{"name": "gyaase.fire_attended", "subjectRef": "old-ref-2", "occurredAt": old},
		{"name": "epono.pod_heard", "subjectRef": "fresh-ref", "occurredAt": now},
	} {
		if _, err := database.Collection("analytics_events").InsertOne(ctx, event); err != nil {
			t.Fatal(err)
		}
	}
	// Privacy requests: one completed long ago (delete), one open (keep).
	if _, err := database.Collection("privacy_requests").InsertMany(ctx, []any{
		bson.M{"_id": "pr_old", "accountId": "id_1", "status": "completed", "completedAt": old, "dueAt": old, "version": 3, "createdAt": old},
		bson.M{"_id": "pr_open", "accountId": "id_1", "status": "requested", "dueAt": now, "version": 1, "createdAt": now},
	}); err != nil {
		t.Fatal(err)
	}

	runner := retention.NewRunner(database, retention.BindingPolicies, time.Now)
	reports, err := runner.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	byPolicy := map[string]int{}
	for _, report := range reports {
		byPolicy[report.Policy] = report.Matched
	}
	// Strip applies to every row older than 90 days (the ancient row too).
	if byPolicy["analytics_pseudonymize_90d"] != 3 {
		t.Fatalf("pseudonymized = %d, want 3", byPolicy["analytics_pseudonymize_90d"])
	}
	if byPolicy["analytics_aggregate_13mo"] != 1 {
		t.Fatalf("aggregated = %d, want 1", byPolicy["analytics_aggregate_13mo"])
	}
	if byPolicy["privacy_requests_completed_90d"] != 1 {
		t.Fatalf("deleted requests = %d, want 1", byPolicy["privacy_requests_completed_90d"])
	}

	// Pseudonymized: subjectRef stripped on old rows, fresh untouched.
	count, err := database.Collection("analytics_events").CountDocuments(ctx, bson.M{"subjectRef": bson.M{"$exists": true}})
	if err != nil || count != 1 {
		t.Fatalf("rows with subjectRef = %d, want 1 (fresh only)", count)
	}
	// Aggregated: ancient event rolled into per-day counts and deleted.
	var aggregate struct {
		Count int `bson:"count"`
	}
	if err := database.Collection("analytics_aggregates").FindOne(ctx, bson.M{"_id": "epono.pod_heard|" + ancient.UTC().Format("2006-01-02")}).Decode(&aggregate); err != nil {
		t.Fatalf("aggregate missing: %v", err)
	}
	if aggregate.Count != 1 {
		t.Fatalf("aggregate count = %d", aggregate.Count)
	}
	count, _ = database.Collection("analytics_events").CountDocuments(ctx, bson.M{})
	if count != 3 {
		t.Fatalf("events after aggregation = %d, want 3", count)
	}
	// Privacy: old completed request gone, open kept.
	count, _ = database.Collection("privacy_requests").CountDocuments(ctx, bson.M{})
	if count != 1 {
		t.Fatalf("privacy requests = %d, want 1", count)
	}

	// Proof-of-retention records exist per policy.
	count, _ = database.Collection("retention_audit").CountDocuments(ctx, bson.M{})
	if count != int64(len(retention.BindingPolicies)) {
		t.Fatalf("proof records = %d, want %d", count, len(retention.BindingPolicies))
	}

	// Rerun is a no-op (idempotent).
	reports, err = runner.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, report := range reports {
		if report.Matched != 0 {
			t.Fatalf("rerun matched %d on %s, want 0", report.Matched, report.Policy)
		}
	}
}
