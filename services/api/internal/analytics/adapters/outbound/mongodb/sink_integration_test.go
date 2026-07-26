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

const integrationTimeout = 3 * time.Minute

func TestAnalyticsPipelineEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_analytics_test")
	sink := mongodb.NewSink(database)
	if err := sink.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	service := application.NewAnalyticsService(sink, nil, time.Now)

	// Valid event persists pseudonymized.
	if err := service.Emit(ctx, "m-1", "epono.pod_heard", map[string]any{"durationPct": 92, "trustPathType": "circle"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	count, err := sink.CountByName(ctx, "epono.pod_heard")
	if err != nil || count != 1 {
		t.Fatalf("events = %d, want 1", count)
	}
	var stored struct {
		SubjectRef string `bson:"subjectRef"`
	}
	if err := database.Collection("analytics_events").FindOne(ctx, bson.M{"name": "epono.pod_heard"}).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.SubjectRef == "m-1" || stored.SubjectRef == "" {
		t.Fatalf("subjectRef leaked member id: %q", stored.SubjectRef)
	}

	// Content smuggling fails before persistence.
	if err := service.Emit(ctx, "m-1", "epono.pod_heard", map[string]any{
		"durationPct": 92, "trustPathType": "circle", "transcript": "private words",
	}); err == nil {
		t.Fatal("content prop must fail")
	}
	count, _ = sink.CountByName(ctx, "epono.pod_heard")
	if count != 1 {
		t.Fatal("invalid event persisted")
	}

	// Opted-out member emits nothing (gate wired).
	service = application.NewAnalyticsService(sink, optOutGate{}, time.Now)
	if err := service.Emit(ctx, "m-2", "gyaase.ember_converted", nil); err != application.ErrAnalyticsOptedOut {
		t.Fatalf("opted out emit = %v", err)
	}
	count, _ = sink.CountByName(ctx, "gyaase.ember_converted")
	if count != 0 {
		t.Fatal("opted-out event persisted")
	}
}

type optOutGate struct{}

func (optOutGate) AllowsAnalytics(context.Context, string) (bool, error) { return false, nil }
