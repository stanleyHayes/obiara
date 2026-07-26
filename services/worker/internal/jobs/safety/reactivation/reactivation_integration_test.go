//go:build integration

package reactivation_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/safety/reactivation"
)

const integrationTimeout = 3 * time.Minute

func TestSuspensionReactivationEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_reactivation_test")

	// One expired suspension (lifted), one running suspension (untouched),
	// one active account (untouched).
	expired := time.Now().Add(-time.Hour)
	running := time.Now().Add(48 * time.Hour)
	for _, account := range []bson.M{
		{"_id": "m-expired", "phone": "+233550000191", "status": "suspended", "suspendedUntil": expired, "tier": 1, "version": 2, "createdAt": time.Now()},
		{"_id": "m-running", "phone": "+233550000192", "status": "suspended", "suspendedUntil": running, "tier": 1, "version": 2, "createdAt": time.Now()},
		{"_id": "m-active", "phone": "+233550000193", "status": "active", "tier": 1, "version": 1, "createdAt": time.Now()},
	} {
		if _, err := database.Collection("accounts").InsertOne(ctx, account); err != nil {
			t.Fatal(err)
		}
	}

	store := reactivation.NewStore(database, time.Now)
	count, err := store.ReactivateExpired(ctx)
	if err != nil || count != 1 {
		t.Fatalf("reactivated = %d, %v; want 1", count, err)
	}

	statusOf := func(id string) string {
		var document struct {
			Status string `bson:"status"`
		}
		if err := database.Collection("accounts").FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
			t.Fatal(err)
		}
		return document.Status
	}
	if got := statusOf("m-expired"); got != "active" {
		t.Fatalf("expired suspension = %q, want active", got)
	}
	if got := statusOf("m-running"); got != "suspended" {
		t.Fatalf("running suspension = %q, must stay suspended", got)
	}
	if got := statusOf("m-active"); got != "active" {
		t.Fatalf("active account = %q", got)
	}
}
