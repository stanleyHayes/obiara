// Package catalog is the composition root of the commerce catalog: the SKUs
// an operator curates and a member can read once published.
package catalog

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/application"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("catalog module requires a privacy secret")

type Module struct {
	Catalog application.Service
}

// NewModule builds the slice. authority decides who may curate; the
// composition root bridges it to the admin context.
func NewModule(ctx context.Context, database *mongo.Database, authority application.Authority, secret string) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
	}
	if authority == nil {
		return Module{}, errors.New("catalog module requires an authority")
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
		Catalog: application.NewService(repository, authority, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "sku_" + base64.RawURLEncoding.EncodeToString(value)
}
