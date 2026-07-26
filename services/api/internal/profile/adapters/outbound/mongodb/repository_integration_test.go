//go:build integration

package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

func TestRepositoryEnforcesRevisionAndGlobalCommandIdempotency(t *testing.T) {
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
	repository := NewRepository(client.Database("obiara_profile_test"))
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	first := integrationProfile(t, "member-1", "cmd-global")
	if err := repository.Save(ctx, first, 0, "cmd-global"); err != nil {
		t.Fatalf("create: %v", err)
	}
	loaded, err := repository.Find(ctx, "member-1")
	if err != nil || loaded.Revision() != 1 || loaded.DisplayName().Value() != "Ama" {
		t.Fatalf("loaded = %#v, error = %v", loaded, err)
	}

	display, _ := domain.NewField("Akua", domain.VisibilityCircles, "", 80, true)
	intro, _ := domain.NewField("", domain.VisibilityPrivate, "", 280, false)
	updated, err := loaded.Update(domain.Change{
		CommandID: "cmd-update", ExpectedRevision: 1, DisplayName: display, Introduction: intro,
		RecordedAt: time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, updated, 1, "cmd-update"); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := repository.Save(ctx, updated, 1, "cmd-update"); !errors.Is(err, application.ErrCommandAlreadyApplied) {
		t.Fatalf("replay error = %v, want %v", err, application.ErrCommandAlreadyApplied)
	}

	other := integrationProfile(t, "member-2", "cmd-global")
	if err := repository.Save(ctx, other, 0, "cmd-global"); !errors.Is(err, application.ErrCommandAlreadyApplied) {
		t.Fatalf("global duplicate error = %v, want %v", err, application.ErrCommandAlreadyApplied)
	}

	staleDisplay, _ := domain.NewField("Esi", domain.VisibilityPrivate, "", 80, true)
	stale, _ := loaded.Update(domain.Change{
		CommandID: "cmd-stale", ExpectedRevision: 1, DisplayName: staleDisplay, Introduction: intro,
		RecordedAt: time.Date(2026, 7, 26, 19, 0, 0, 0, time.UTC),
	})
	if err := repository.Save(ctx, stale, 1, "cmd-stale"); !errors.Is(err, application.ErrOptimisticConflict) {
		t.Fatalf("stale error = %v, want %v", err, application.ErrOptimisticConflict)
	}
}

func integrationProfile(t *testing.T, memberID, commandID string) domain.Profile {
	t.Helper()
	display, _ := domain.NewField("Ama", domain.VisibilityCircles, "", 80, true)
	intro, _ := domain.NewField("", domain.VisibilityPrivate, "", 280, false)
	profile, err := domain.Create(memberID, domain.Change{
		CommandID: commandID, DisplayName: display, Introduction: intro,
		RecordedAt: time.Date(2026, 7, 26, 17, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	return profile
}
