// Package listening is the composition root of the seed context's
// listening-eligibility slice (E06-S03). The sow boundary (E06-S04) reads
// eligibility through the application service only.
package listening

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/application"
)

type Module struct {
	Listening application.ListeningService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Listening: application.NewListeningService(repository, time.Now)}, nil
}
