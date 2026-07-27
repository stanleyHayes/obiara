//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	notificationmongodb "github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	notificationapplication "github.com/stanleyHayes/obiara/internal/notifications/application"
	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

const careIntegrationTimeout = 3 * time.Minute

func TestCareQueueAndQuieteningEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), careIntegrationTimeout)
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

	database := client.Database("obiara_care_test")
	careRepository := mongodb.NewCareRepository(database)
	if err := careRepository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "care_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewCareService(careRepository, careRepository, time.Now, ids)

	// Closure flag: care case opens and the quietening window is set.
	careCase, err := service.Flag(ctx, "m-1", domain.SignalClosure)
	if err != nil {
		t.Fatalf("flag: %v", err)
	}
	if careCase.Status() != domain.CareOpen {
		t.Fatalf("case = %#v", careCase)
	}
	quiet, err := careRepository.QuietUntil(ctx, "m-1", time.Now())
	if err != nil || !quiet {
		t.Fatalf("quietening not set: %v, %v", quiet, err)
	}

	// Notifications: quietened for pods, but safety still delivers.
	preferencesRepository := notificationmongodb.NewRepository(database)
	decisionClock := func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	}
	decider := notificationapplication.NewNotificationServiceWithQuietening(preferencesRepository, preferencesRepository, careRepository, decisionClock)
	decision, err := decider.Decide(ctx, "m-1", notificationdomain.CategoryPods)
	if err != nil || decision.Allowed || decision.Reason != "care_quietening" {
		t.Fatalf("decision = %v, want suppressed care_quietening", decision)
	}
	decision, err = decider.Decide(ctx, "m-1", notificationdomain.CategorySafety)
	if err != nil || !decision.Allowed {
		t.Fatalf("safety during quietening = %v, want allowed", decision)
	}
	// Other members are unaffected.
	decision, err = decider.Decide(ctx, "m-2", notificationdomain.CategoryPods)
	if err != nil || !decision.Allowed {
		t.Fatalf("other member = %v, want allowed", decision)
	}

	// Engage and resolve with approved scripts only.
	if _, err := service.Engage(ctx, careCase.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resolve(ctx, careCase.ID(), []domain.ScriptKey{domain.ScriptKey("diagnose_ptsd")}); err != domain.ErrInvalidScript {
		t.Fatalf("diagnostic script = %v, want rejected", err)
	}
	resolved, err := service.Resolve(ctx, careCase.ID(), []domain.ScriptKey{domain.ScriptHelplineDirectory, domain.ScriptClosureQuietening})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status() != domain.CareResolved || len(resolved.Scripts()) != 2 {
		t.Fatalf("resolved = %#v", resolved)
	}

	// Queue shows nothing open.
	open, err := service.NextOpen(ctx, 10)
	if err != nil || len(open) != 0 {
		t.Fatalf("open = %#v", open)
	}
}
