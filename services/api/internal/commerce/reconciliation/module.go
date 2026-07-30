package reconciliation

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/application"
)

type Module struct {
	Queries application.QueryService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Queries: application.NewQueryService(repository)}, nil
}
