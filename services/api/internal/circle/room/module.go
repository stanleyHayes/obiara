package room

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
)

type Module struct {
	Rooms application.Service
}

func NewModule(ctx context.Context, database *mongo.Database, secret string, authorizer application.Authorizer) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.New([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{
		Rooms: application.NewService(authorizer, repository, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "entry_" + base64.RawURLEncoding.EncodeToString(value)
}
