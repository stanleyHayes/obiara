// Package organization is the composition root of the organization context.
//
// An organization is a body Obiara has a relationship with, so that a discount
// code has an issuer and an audit trail has somebody to name. See
// agent_plan.md §41 for why there is no organization sign-in: codes are issued
// by staff on an organization's behalf, and a console is additive later.
package organization

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/organization/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/organization/application"
)

type Module struct {
	Organizations application.Service
	// Keyer is exposed because a payout records which operator approved it,
	// and an operator digest belongs to whichever context is keying
	// operators. This one already is.
	Keyer application.Keyer
}

// ErrSecretRequired reports a module built with no keying secret. The operator
// in an audit trail is a digest, and without a secret there is nothing to
// digest with — so this fails at startup rather than writing a trail that
// names people in the clear.
var ErrSecretRequired = errors.New("organization module requires a keying secret")

func NewModule(ctx context.Context, database *mongo.Database, secret string) (Module, error) {
	if secret == "" {
		return Module{}, ErrSecretRequired
	}
	keyer, err := newKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Organizations: application.New(repository, keyer, idSource{}, time.Now),
		Keyer:         keyer,
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "org_" + hex.EncodeToString(value)
}
