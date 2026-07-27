//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrentIdempotencyAndPrivacy(t *testing.T) {
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
	db := client.Database("p0_gate_test")
	repo := adapter.New(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	definition, _ := domain.NewDefinition(domain.DefinitionSpec{ID: "p0.gates", Version: 1, ReviewID: "founder.gates", ReviewerKey: key(1), ReviewedAt: now, PodsHeardPermille: 650, SeedToSproutPermille: 250, SproutToRoomPermille: 350, WeeklyFirePermille: 400, Day30RetentionPermille: 450})
	snapshot := domain.Snapshot{ID: key(2), WindowKey: key(3), SourceWatermark: key(4), Version: 1, WindowStartedAt: now.Add(-time.Hour), WindowEndedAt: now, CohortSize: 100, PodEligible: 100, PodsHeard: 65, SeedsSown: 100, SproutsOpened: 25, SproutEligible: 100, RoomsOpened: 35, WeeklyFireAttendees: 40, Day30Eligible: 100, Day30Retained: 45, PreviousRegretReports: 2, CurrentRegretReports: 1, CompleteMetrics: []domain.Metric{domain.MetricPodsHeard, domain.MetricSeedToSprout, domain.MetricSproutToRoom, domain.MetricWeeklyFire, domain.MetricDay30Retention, domain.MetricRegretTrend, domain.MetricTierAResolved}}
	report, _ := domain.Evaluate(key(5), definition, snapshot, now)
	ch := make(chan error, 2)
	go func() { ch <- repo.Insert(ctx, report) }()
	go func() { ch <- repo.Insert(ctx, report) }()
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
	var raw bson.M
	if e = db.Collection("analytics_p0_gate_reports").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "memberid", "content", "voice", "conversation", "profile", "rank", "score", "freeText"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
}
