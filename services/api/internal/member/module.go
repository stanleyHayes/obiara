package member

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/member/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/member/application"
)

type Module struct {
	Register application.RegisterMember
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Register: application.NewRegisterMember(repository, time.Now),
	}, nil
}
