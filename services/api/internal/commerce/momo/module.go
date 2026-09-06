// Package momo is the composition root of the mobile money collection.
//
// It is composed only when a provider is configured. Without one there is no
// way to take money at all, and the product is honest about that by simply not
// offering a purchase — see agent_plan.md §72 for how long that was the case
// without anybody noticing.
package momo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
)

type Module struct {
	Intents application.Service
}

// ErrDependenciesRequired reports a module built without a provider or a
// callback secret. Both decide whether money moved: a nil provider would let
// an intent be confirmed with nobody prompted, and an empty secret would let
// any caller claim a payment succeeded.
var ErrDependenciesRequired = errors.New(
	"mobile money module requires a provider and a callback secret",
)

func NewModule(
	ctx context.Context, database *mongo.Database,
	provider application.Provider, secret string,
) (Module, error) {
	if provider == nil || len(secret) < 32 {
		return Module{}, ErrDependenciesRequired
	}
	repository := mongodb.New(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Intents: application.New(repository, provider, idSource{}, systemClock{}, []byte(secret)),
	}, nil
}

type idSource struct{}

// NewID is 64 hex characters because every identifier the payment context
// stores is opaque: an intent id ends up in a provider's logs and in a
// callback, and one that carried anything readable would carry it there.
func (idSource) NewID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
