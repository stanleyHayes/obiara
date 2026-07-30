// Package marketpack is the composition root of the market-pack
// governance slice (E16-S06).
package marketpack

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/application"
)

type Module struct {
	Packs application.MarketPackService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Packs: application.NewMarketPackService(repository, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "pack_" + base64.RawURLEncoding.EncodeToString(id)
}
