// Package allowance is the composition root of the weekly seed allowance:
// the server-authoritative, non-purchasable budget a member spends when
// sowing, renewed on the Monday of their own week.
package allowance

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
)

var (
	// ErrSecretRequired reports a module built without a privacy secret.
	ErrSecretRequired = errors.New("allowance module requires a privacy secret")
	// ErrAllowanceRequired reports a non-positive weekly budget. A zero
	// allowance would silently forbid sowing rather than gating it.
	ErrAllowanceRequired = errors.New("weekly seed allowance must be positive")
)

type Module struct {
	Allowances *application.Service
}

// NewModule builds the slice. timezone is the IANA zone whose Monday starts
// the member's week; weekly is the number of seeds granted each week.
func NewModule(ctx context.Context, database *mongo.Database, secret, timezone string, weekly int64) (Module, error) {
	if len(secret) < 32 {
		return Module{}, ErrSecretRequired
	}
	if weekly <= 0 {
		return Module{}, ErrAllowanceRequired
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	week, err := domain.NewWeekPolicy(timezone)
	if err != nil {
		return Module{}, err
	}
	service, err := application.NewService(repository, keyer, week, time.Now, weekly)
	if err != nil {
		return Module{}, err
	}
	return Module{Allowances: service}, nil
}
