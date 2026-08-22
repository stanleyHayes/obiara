// Package ledger is the composition root of the double-entry commerce
// ledger: the record of what the platform owes and is owed.
package ledger

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("ledger module requires a privacy secret")

type Module struct {
	Ledger application.Service
}

// NewModule builds the slice. authority decides who may post and who may
// read balances; the composition root bridges it to the admin context.
func NewModule(ctx context.Context, database *mongo.Database, authority application.Authority, secret string) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
	}
	if authority == nil {
		return Module{}, errors.New("ledger module requires an authority")
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.New([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{
		Ledger: application.NewService(repository, authority, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "post_" + base64.RawURLEncoding.EncodeToString(value)
}
