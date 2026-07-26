//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	notificationmongodb "github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	notificationapplication "github.com/stanleyHayes/obiara/internal/notifications/application"
	ritualapplication "github.com/stanleyHayes/obiara/internal/notifications/ritual/application"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/ritual/adapters/outbound/mongodb"
)

const ritualIntegrationTimeout = 3 * time.Minute

func TestRitualDispatchEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), ritualIntegrationTimeout)
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

	database := client.Database("obiara_ritual_test")

	// Seed: one active member with defaults, one with pods muted (irrelevant
	// here), one blocked account that must be skipped.
	if _, err := database.Collection("accounts").InsertMany(ctx, []any{
		bson.M{"_id": "m-1", "phone": "+233550000101", "status": "active", "createdAt": time.Now()},
		bson.M{"_id": "m-2", "phone": "+233550000102", "status": "blocked", "createdAt": time.Now()},
	}); err != nil {
		t.Fatal(err)
	}

	// A fire starting in 30 minutes with m-1 going.
	startsAt := time.Now().Add(30 * time.Minute)
	if _, err := database.Collection("fires").InsertOne(ctx, bson.M{
		"_id": "fire_1", "hostId": "host-1", "title": "Friday Fire", "startsAt": startsAt,
		"capacity": 50, "goingCount": 1, "status": "scheduled", "version": 1, "createdAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Collection("fire_attendance").InsertOne(ctx, bson.M{
		"_id": "fire_1|m-1", "fireId": "fire_1", "memberId": "m-1", "status": "going",
		"position": 0, "version": 1, "createdAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	outboxStore := outbox.NewStore(database, time.Now)
	if err := outboxStore.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	notificationRepository := notificationmongodb.NewRepository(database)
	// Pin the decision clock outside quiet hours (21:00-07:00 default) so
	// preference evaluation is deterministic regardless of wall time.
	decisionClock := func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	}
	decider := notificationapplication.NewNotificationService(notificationRepository, notificationRepository, decisionClock)
	sources := mongodb.NewSources(database)
	dispatcher := ritualapplication.NewDispatcher(
		sources, sources, decider, sources, outboxStore, inbox.NewStore(database, time.Now), time.Now,
	)

	// Herald: m-1 gets exactly one herald event; the blocked account and
	// non-attendees get nothing.
	if err := dispatcher.DispatchHeralds(ctx); err != nil {
		t.Fatalf("herald dispatch: %v", err)
	}
	heraldCount, err := database.Collection("outbox").CountDocuments(ctx, bson.M{"eventType": "notification.ritual.fire_herald"})
	if err != nil || heraldCount != 1 {
		t.Fatalf("herald events = %d, want 1", heraldCount)
	}

	// Re-running is a no-op (inbox dedup).
	if err := dispatcher.DispatchHeralds(ctx); err != nil {
		t.Fatal(err)
	}
	heraldCount, _ = database.Collection("outbox").CountDocuments(ctx, bson.M{"eventType": "notification.ritual.fire_herald"})
	if heraldCount != 1 {
		t.Fatalf("herald redispatched: %d events", heraldCount)
	}

	// Calendar: dawn is not due at 03:00 local; the blocked account is skipped.
	night := time.Now().Add(-3 * time.Hour)
	dispatcherNight := ritualapplication.NewDispatcher(
		sources, sources, decider, sources, outboxStore, inbox.NewStore(database, time.Now),
		func() time.Time { return time.Date(night.Year(), night.Month(), night.Day(), 3, 0, 0, 0, time.UTC) },
	)
	if err := dispatcherNight.DispatchCalendar(ctx); err != nil {
		t.Fatal(err)
	}
	dawnCount, _ := database.Collection("outbox").CountDocuments(ctx, bson.M{"eventType": "notification.ritual.dawn"})
	if dawnCount != 0 {
		t.Fatalf("dawn dispatched before 06:00: %d", dawnCount)
	}

	// At 06:30 local, dawn dispatches exactly once for m-1.
	morning := time.Now().Add(3*time.Hour + 30*time.Minute)
	dispatcherMorning := ritualapplication.NewDispatcher(
		sources, sources, decider, sources, outboxStore, inbox.NewStore(database, time.Now),
		func() time.Time { return morning },
	)
	if morning.Hour() < 6 {
		t.Skipf("test clock before 06:00 local, adjust fixture")
	}
	if err := dispatcherMorning.DispatchCalendar(ctx); err != nil {
		t.Fatal(err)
	}
	if err := dispatcherMorning.DispatchCalendar(ctx); err != nil {
		t.Fatal(err)
	}
	dawnCount, _ = database.Collection("outbox").CountDocuments(ctx, bson.M{"eventType": "notification.ritual.dawn"})
	if dawnCount > 1 {
		t.Fatalf("dawn events = %d, want at most 1 (deduped)", dawnCount)
	}
}
