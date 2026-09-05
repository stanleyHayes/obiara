package circle

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/application"
)

type Module struct {
	Circles application.Service
	// Circle reading, exposed for contexts that need a roster rather than a
	// transition — introduction sources resolve candidates from one.
	Repository *mongodb.Repository
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Circles:    application.NewService(repository, time.Now),
		Repository: repository,
	}, nil
}
