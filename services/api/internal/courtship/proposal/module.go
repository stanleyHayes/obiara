// Package proposal is the composition root of the courtship proposal slice:
// one member proposing a call, a meeting or exclusivity to another, and the
// other accepting, rejecting or the sender withdrawing it.
package proposal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/application"
)

// ErrSecretRequired reports a module built without a privacy secret. Member
// identities and proposal details are keyed with it, so a missing secret
// would mean writing them legibly to the store.
var ErrSecretRequired = errors.New("courtship proposal module requires a privacy secret")

type Module struct {
	Proposals application.Service
}

// NewModule builds the slice. secret keys member references and protects
// proposal details at rest.
func NewModule(ctx context.Context, database *mongo.Database, secret string) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
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
		Proposals: application.New(
			repository, keyer, privacy.NewProtector(keyer), idSource{}, time.Now,
		),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "prop_" + base64.RawURLEncoding.EncodeToString(value)
}
