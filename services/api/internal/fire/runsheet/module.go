// Package runsheet is the composition root of the fire run sheet: the
// ordered segments a host works through while running a fire.
package runsheet

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/application"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("run sheet module requires a privacy secret")

type Module struct {
	RunSheets application.Service
}

// NewModule builds the slice. authority decides who may run a fire; the
// composition root bridges it to the fire context.
func NewModule(ctx context.Context, database *mongo.Database, authority application.Authority, secret string) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
	}
	if authority == nil {
		return Module{}, errors.New("run sheet module requires an authority")
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{
		RunSheets: application.NewService(repository, authority, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "runsheet_" + base64.RawURLEncoding.EncodeToString(value)
}
