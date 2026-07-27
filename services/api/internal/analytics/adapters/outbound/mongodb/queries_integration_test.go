//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/application"
)

const metricsIntegrationTimeout = 3 * time.Minute

func TestFunnelMetricsEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), metricsIntegrationTimeout)
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

	database := client.Database("obiara_metrics_test")
	now := time.Now()
	recent := now.Add(-time.Hour)

	// Funnel: 4 sown → 3 heard, 2 sprouted, 1 room.
	insertEvents := func(name string, subjects []string, at time.Time) {
		for _, subject := range subjects {
			if _, err := database.Collection("analytics_events").InsertOne(ctx, bson.M{
				"name": name, "subjectRef": subject, "occurredAt": at,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	insertEvents("epono.seed_sown", []string{"a", "b", "c", "d"}, recent)
	insertEvents("epono.pod_heard", []string{"a", "b", "c"}, recent)
	insertEvents("epono.sprout_opened", []string{"a", "b"}, recent)
	insertEvents("epono.room_opened", []string{"a"}, recent)
	insertEvents("gyaase.fire_attended", []string{"a", "b"}, recent)
	insertEvents("wellbeing.regret_reported", []string{"a"}, recent)
	insertEvents("wellbeing.regret_reported", []string{"x", "y", "z"}, now.Add(-45*24*time.Hour))

	// Cohort: 4 active accounts.
	for _, id := range []string{"a", "b", "c", "d"} {
		if _, err := database.Collection("accounts").InsertOne(ctx, bson.M{
			"_id": id, "phone": "+2335500001" + id, "status": "active", "tier": 1, "version": 1, "createdAt": now,
		}); err != nil {
			t.Fatal(err)
		}
	}

	service := application.NewMetricsService(mongodb.NewQueryStore(database), mongodb.NewCohortStore(database), time.Now)
	report, err := service.Funnel(ctx, 30)
	if err != nil {
		t.Fatal(err)
	}
	if report.PodsHeardRate != 0.75 {
		t.Fatalf("podsHeard = %v, want 0.75", report.PodsHeardRate)
	}
	if report.SeedToSproutRate != 0.5 {
		t.Fatalf("seedToSprout = %v, want 0.5", report.SeedToSproutRate)
	}
	if report.SproutToRoomRate != 0.5 {
		t.Fatalf("sproutToRoom = %v, want 0.5", report.SproutToRoomRate)
	}
	if report.FireAttendeeCount != 2 || report.FireAttendanceRate != 0.5 {
		t.Fatalf("fire = %d/%v, want 2/0.5", report.FireAttendeeCount, report.FireAttendanceRate)
	}
	// Current 1 regret; prior window had 3 → down.
	if report.RegretTrend != "down" {
		t.Fatalf("trend = %q, want down", report.RegretTrend)
	}
}
