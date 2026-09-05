// Package safeguarding owns age hard-block and personal-data purge policy.
//
// The context was built complete and left unreachable: the domain carries the
// minimum age, correct birthday arithmetic and a purge SLA, and until this
// module existed nothing could call any of it. A date of birth arrived at
// verification, was written to Mongo in plain text, and no code anywhere asked
// how old the person was.
package safeguarding

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/application"
)

type Module struct {
	Safeguarding application.Service
}

// ErrSecretRequired reports a module built without keying material. The
// restriction record is the only thing that outlives the purge, and an
// unkeyed subject digest would make it a list of which accounts belonged to
// children — the opposite of what retaining it is for.
var ErrSecretRequired = errors.New("safeguarding module requires a keying secret")

// NewModule composes the age gate over its Mongo store and purger.
func NewModule(ctx context.Context, database *mongo.Database, secret []byte) (Module, error) {
	if len(secret) == 0 {
		return Module{}, ErrSecretRequired
	}
	keyer, err := privacy.NewHMACKeyer(secret)
	if err != nil {
		return Module{}, err
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Safeguarding: application.NewService(
			repository, mongodb.NewArtifactPurger(database), keyer, idSource{}, time.Now,
		),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		// A weak identifier on a compliance record is worse than a crash at
		// startup, which is where this failure would surface.
		panic(err)
	}
	return "sg_" + base64.RawURLEncoding.EncodeToString(value)
}
