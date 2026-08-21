package liveness

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

type Module struct {
	Liveness  application.Service
	Artifacts application.ArtifactService
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "live_" + base64.RawURLEncoding.EncodeToString(value)
}

// ErrProviderRequired reports a module built with no liveness provider.
var ErrProviderRequired = errors.New("liveness module requires a provider")

// NewModule builds the liveness context. provider assesses attempts and must
// not be nil.
func NewModule(ctx context.Context, database *mongo.Database, provider application.Provider, hmacSecret string) (Module, error) {
	if provider == nil {
		return Module{}, ErrProviderRequired
	}
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewHMACKeyer([]byte(hmacSecret))
	if err != nil {
		return Module{}, err
	}
	sealer, err := privacy.NewAESGCMSealer([]byte(hmacSecret))
	if err != nil {
		return Module{}, err
	}
	ids := idSource{}
	return Module{
		Liveness: application.NewService(
			store, provider, store, keyer, ids, time.Now,
		),
		Artifacts: application.NewArtifactService(store, sealer, keyer, ids, time.Now),
	}, nil
}
