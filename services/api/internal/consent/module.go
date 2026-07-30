package consent

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

type Module struct {
	Onboarding application.OnboardingService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	effectiveSince := time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC)
	purposes := make([]domain.Purpose, 0, 3)
	for _, params := range []domain.NewPurposeParams{
		{ID: application.CommunityPromiseID, Kind: domain.PurposePromise, Version: application.CurrentVersion, ContentRef: "content.promise.v1", Status: domain.PurposeActive, EffectiveSince: effectiveSince},
		{ID: application.ServiceTermsID, Kind: domain.PurposeTerms, Version: application.CurrentVersion, ContentRef: "content.terms.v1", Status: domain.PurposeActive, EffectiveSince: effectiveSince},
		{ID: application.AdultAgeID, Kind: domain.PurposeAge, Version: application.CurrentVersion, ContentRef: "content.age.v1", Status: domain.PurposeActive, EffectiveSince: effectiveSince},
	} {
		purpose, err := domain.NewPurpose(params)
		if err != nil {
			return Module{}, err
		}
		purposes = append(purposes, purpose)
	}
	service := application.NewService(store, mongodb.NewCatalog(purposes...), time.Now)
	return Module{Onboarding: application.NewOnboardingService(service)}, nil
}
