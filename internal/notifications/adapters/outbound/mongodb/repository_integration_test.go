//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/application"
	"github.com/stanleyHayes/obiara/internal/notifications/domain"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

const notificationsIntegrationTimeout = 3 * time.Minute

func TestNotificationCapsEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), notificationsIntegrationTimeout)
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

	database := client.Database("obiara_notifications_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	// Pin the decision clock outside default quiet hours (21:00-07:00) so
	// the suite is wall-clock independent.
	decisionClock := func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	}
	service := application.NewNotificationService(repository, repository, decisionClock)

	// Defaults are created on first read.
	preferences, err := service.Get(ctx, "m-1")
	if err != nil {
		t.Fatal(err)
	}
	if preferences.Timezone() != "Africa/Accra" {
		t.Fatalf("defaults = %#v", preferences)
	}

	// Configure: mute pods, custom quiet window.
	if _, err := service.Configure(ctx, "m-1", map[domain.Category]bool{domain.CategoryPods: true}, 22*60, 6*60, "Africa/Accra"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Configure(ctx, "m-1", map[domain.Category]bool{domain.CategorySafety: true}, 0, 0, "Africa/Accra"); err != domain.ErrSafetyCannotBeMuted {
		t.Fatalf("muting safety = %v, want rejected", err)
	}

	// Cap race: 20 concurrent sends, exactly DailyCap slots claim.
	const senders = 20
	var wg sync.WaitGroup
	allowed := make([]bool, senders)
	for i := 0; i < senders; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			decision, err := service.Decide(ctx, "m-1", domain.CategoryRitual)
			if err != nil {
				t.Errorf("decide: %v", err)
				return
			}
			allowed[index] = decision.Allowed
		}(i)
	}
	wg.Wait()

	claimed := 0
	for _, ok := range allowed {
		if ok {
			claimed++
		}
	}
	if claimed != domain.DailyCap {
		t.Fatalf("claimed = %d, want exactly %d after cap race", claimed, domain.DailyCap)
	}

	// Muted category is suppressed without touching the cap.
	decision, err := service.Decide(ctx, "m-1", domain.CategoryPods)
	if err != nil || decision.Allowed || decision.Reason != "muted" {
		t.Fatalf("muted = %v, want suppressed", decision)
	}

	// Safety still delivers at the cap.
	decision, err = service.Decide(ctx, "m-1", domain.CategorySafety)
	if err != nil || !decision.Allowed {
		t.Fatalf("safety at cap = %v, want allowed", decision)
	}
}
