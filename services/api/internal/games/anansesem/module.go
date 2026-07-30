package anansesem

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/application"
)

type Module struct {
	Stories application.Service
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	secret string,
	authority application.Authority,
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
		Stories: application.NewService(
			repository, authority, keyer, idSource{"story_"},
			idSource{"passage_"}, time.Now,
		),
	}, nil
}

type idSource struct{ prefix string }

func (source idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return source.prefix + base64.RawURLEncoding.EncodeToString(value)
}
