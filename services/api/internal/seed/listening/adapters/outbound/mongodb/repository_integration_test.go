//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/application"
)

const listeningIntegrationTimeout = 3 * time.Minute

func TestListeningEligibilityEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), listeningIntegrationTimeout)
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

	database := client.Database("obiara_listening_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	service := application.NewListeningService(repository, time.Now)

	// Below threshold after the first batch.
	record, err := service.RecordHeartbeats(ctx, "m-1", "asset-1", 60, []application.HeartbeatRange{{Start: 0, End: 12}})
	if err != nil {
		t.Fatalf("first batch: %v", err)
	}
	if record.Eligible() || record.TotalSeconds() != 12 {
		t.Fatalf("record = %v/%v", record.TotalSeconds(), record.Eligible())
	}

	// Out-of-order, overlapping and replayed ranges merge without
	// double-counting: 12 + (18-12) + (30-25) = 23 unique seconds.
	record, err = service.RecordHeartbeats(ctx, "m-1", "asset-1", 60, []application.HeartbeatRange{
		{Start: 25, End: 30},
		{Start: 0, End: 12}, // replay of the first batch
		{Start: 10, End: 18},
	})
	if err != nil {
		t.Fatalf("second batch: %v", err)
	}
	if record.TotalSeconds() != 23 {
		t.Fatalf("total = %v, want 23", record.TotalSeconds())
	}

	// Eligibility is server-derived and now granted.
	eligible, total, err := service.Eligibility(ctx, "m-1", "asset-1")
	if err != nil || !eligible || total != 23 {
		t.Fatalf("eligibility = %v/%v, %v", eligible, total, err)
	}

	// A different listener starts from zero.
	eligible, total, err = service.Eligibility(ctx, "m-2", "asset-1")
	if err != nil || eligible || total != 0 {
		t.Fatalf("other listener = %v/%v, %v", eligible, total, err)
	}

	// Clamping: telemetry beyond the asset duration is cut to duration.
	record, err = service.RecordHeartbeats(ctx, "m-2", "asset-1", 60, []application.HeartbeatRange{{Start: 50, End: 95}})
	if err != nil {
		t.Fatalf("clamped batch: %v", err)
	}
	if record.TotalSeconds() != 10 {
		t.Fatalf("clamped total = %v, want 10", record.TotalSeconds())
	}
}
