//go:build integration

package mongodb

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

func TestRepositoryConcurrencyPrivacyAndGlobalReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
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
		t.Fatal(err)
	}
	if strings.Contains(uri, "?") {
		uri += "&directConnection=true"
	} else {
		uri += "?directConnection=true"
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	repository := NewRepository(client.Database("obiara_circle_test"))
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	circle := integrationCircle(t, "circle-1", "create-global", base)
	if err := repository.Save(ctx, circle, 0, "create-global"); err != nil {
		t.Fatalf("create: %v", err)
	}
	loaded, err := repository.Find(ctx, "circle-1")
	if err != nil || loaded.Visibility() != domain.VisibilityPrivate ||
		loaded.Allows("stranger", domain.CapabilityDiscover) {
		t.Fatalf("privacy default lost: circle = %#v, error = %v", loaded, err)
	}

	first, err := loaded.Request("member-1", domain.Command{
		ID: "request-1", ExpectedRevision: 1, ActorID: "member-1",
		Kind: "membership.request", Payload: "member-1", RecordedAt: base.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := loaded.Request("member-2", domain.Command{
		ID: "request-2", ExpectedRevision: 1, ActorID: "member-2",
		Kind: "membership.request", Payload: "member-2", RecordedAt: base.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, candidate := range []domain.Circle{first, second} {
		candidate := candidate
		wait.Add(1)
		go func() {
			defer wait.Done()
			commandID := candidate.History()[candidate.Revision()-1].CommandID()
			results <- repository.Save(ctx, candidate, 1, commandID)
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for saveErr := range results {
		switch {
		case saveErr == nil:
			successes++
		case errors.Is(saveErr, application.ErrOptimisticConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent save error: %v", saveErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent saves: successes=%d conflicts=%d", successes, conflicts)
	}
	after, err := repository.Find(ctx, "circle-1")
	if err != nil || after.Revision() != 2 {
		t.Fatalf("after race = %#v, error = %v", after, err)
	}
	if after.Allows("member-1", domain.CapabilityView) || after.Allows("member-2", domain.CapabilityView) {
		t.Fatal("requested membership gained private-circle access")
	}

	other := integrationCircle(t, "circle-2", "create-global", base)
	if err := repository.Save(ctx, other, 0, "create-global"); !errors.Is(err, application.ErrCommandAlreadyApplied) {
		t.Fatalf("global replay error = %v, want %v", err, application.ErrCommandAlreadyApplied)
	}
}

func integrationCircle(t *testing.T, circleID, commandID string, now time.Time) domain.Circle {
	t.Helper()
	circle, err := domain.Create(circleID, domain.TypeCommunity, "owner-1", domain.Command{
		ID: commandID, ActorID: "owner-1", Kind: "circle.create",
		Payload: string(domain.TypeCommunity), RecordedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return circle
}
