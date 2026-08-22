// Package communityaudit is the composition root of the community audit
// desk: the queue of conduct cases an operator reviews and decides.
package communityaudit

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("community audit module requires a privacy secret")

type Module struct {
	Audit application.Service
}

// NewModule builds the desk. authority and mfa are bridged to the admin
// context by the composition root.
func NewModule(ctx context.Context, database *mongo.Database, authority application.Authorizer, mfa application.MFAGate, secret string) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
	}
	if authority == nil || mfa == nil {
		return Module{}, errors.New("community audit module requires an authority and an mfa gate")
	}
	repository := mongodb.New(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.New([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{Audit: application.New(authority, mfa, keyer, repository, time.Now)}, nil
}
