//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	platformmongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

func TestOnboardingConsentPersistenceAndReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := platformmongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	database := client.Database("obiara_onboarding_consent_test")
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	purposes := make([]domain.Purpose, 0, 3)
	for _, params := range []domain.NewPurposeParams{
		{ID: application.CommunityPromiseID, Kind: domain.PurposePromise, Version: 1, ContentRef: "content.promise.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
		{ID: application.ServiceTermsID, Kind: domain.PurposeTerms, Version: 1, ContentRef: "content.terms.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
		{ID: application.AdultAgeID, Kind: domain.PurposeAge, Version: 1, ContentRef: "content.age.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
	} {
		purpose, err := domain.NewPurpose(params)
		if err != nil {
			t.Fatal(err)
		}
		purposes = append(purposes, purpose)
	}
	service := application.NewOnboardingService(
		application.NewService(store, mongodb.NewCatalog(purposes...), func() time.Time { return now }),
	)
	command := application.OnboardingCommand{
		CommandID: "onboarding-command-1", SubjectID: "member-1", Source: domain.SourceWeb,
	}
	if _, err := service.Accept(ctx, command); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Accept(ctx, command); err != nil {
		t.Fatalf("idempotent replay failed: %v", err)
	}
	count, err := database.Collection("consent_records").CountDocuments(ctx, map[string]any{"subjectId": "member-1"})
	if err != nil || count != 3 {
		t.Fatalf("record count=%d err=%v", count, err)
	}
	for _, purposeID := range []string{
		application.CommunityPromiseID, application.ServiceTermsID, application.AdultAgeID,
	} {
		record, err := store.Find(ctx, application.Key{SubjectID: "member-1", PurposeID: purposeID})
		if err != nil || record.Revision() != 1 || len(record.History()) != 1 {
			t.Fatalf("%s record=%+v err=%v", purposeID, record, err)
		}
	}
}
