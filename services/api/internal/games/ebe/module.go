package ebe

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/application"
)

type Module struct {
	Catalog application.CatalogService
	Duels   application.DuelService
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	secret string,
	pairs application.PairAuthority,
	reviewers application.ReviewerAuthority,
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
		Catalog: application.NewCatalogService(repository, reviewers, keyer, idSource{"review_"}, time.Now),
		Duels:   application.NewDuelService(repository, repository, pairs, keyer, idSource{"ebe_"}),
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
