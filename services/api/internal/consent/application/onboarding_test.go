package application

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

type memoryConsentRepository struct {
	records map[Key]domain.Record
}

func (repository *memoryConsentRepository) Find(_ context.Context, key Key) (domain.Record, error) {
	record, ok := repository.records[key]
	if !ok {
		return domain.Record{}, ErrNotFound
	}
	return record, nil
}

func (repository *memoryConsentRepository) Save(_ context.Context, record domain.Record, expected uint64, _ string) error {
	key := Key{SubjectID: record.SubjectID(), PurposeID: record.PurposeID()}
	current, exists := repository.records[key]
	if (!exists && expected != 0) || (exists && current.Revision() != expected) {
		return ErrOptimisticConflict
	}
	repository.records[key] = record
	return nil
}

type memoryPurposeCatalog map[string]domain.Purpose

func (catalog memoryPurposeCatalog) FindVersion(_ context.Context, id string, version uint64) (domain.Purpose, error) {
	purpose, ok := catalog[id]
	if !ok || purpose.Version() != version {
		return domain.Purpose{}, domain.ErrInvalidPurpose
	}
	return purpose, nil
}

func TestOnboardingAcceptRecordsThreeVersionedReceiptsAndReplays(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	catalog := memoryPurposeCatalog{}
	for _, params := range []domain.NewPurposeParams{
		{ID: CommunityPromiseID, Kind: domain.PurposePromise, Version: 1, ContentRef: "content.promise.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
		{ID: ServiceTermsID, Kind: domain.PurposeTerms, Version: 1, ContentRef: "content.terms.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
		{ID: AdultAgeID, Kind: domain.PurposeAge, Version: 1, ContentRef: "content.age.v1", Status: domain.PurposeActive, EffectiveSince: now.Add(-time.Hour)},
	} {
		purpose, err := domain.NewPurpose(params)
		if err != nil {
			t.Fatal(err)
		}
		catalog[purpose.ID()] = purpose
	}
	repository := &memoryConsentRepository{records: map[Key]domain.Record{}}
	service := NewOnboardingService(NewService(repository, catalog, func() time.Time { return now }))
	command := OnboardingCommand{
		CommandID: "onboarding-command-1", SubjectID: "member-1", Source: domain.SourceWeb,
	}

	first, err := service.Accept(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := service.Accept(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if first != (OnboardingResult{PromiseRevision: 1, TermsRevision: 1, AgeRevision: 1}) ||
		replayed != first || len(repository.records) != 3 {
		t.Fatalf("first=%+v replayed=%+v records=%d", first, replayed, len(repository.records))
	}
	for _, record := range repository.records {
		if len(record.History()) != 1 || record.History()[0].ActorID() != "member-1" {
			t.Fatalf("unexpected consent history: %+v", record.History())
		}
	}
}
