package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
)

type Module struct {
	Sessions application.Service
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	secret string,
	rooms application.RoomEmbedding,
	authorizer application.Authorizer,
) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{
		Sessions: application.NewService(
			repository,
			rooms,
			authorizer,
			keyer,
			idSource{},
			time.Now,
		),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "oware_" + base64.RawURLEncoding.EncodeToString(value)
}
