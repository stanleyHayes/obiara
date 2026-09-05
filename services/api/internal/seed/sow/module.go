// Package sow is the composition root of the sow: the atomic gesture a member
// makes toward another, carrying their answer and costing a seed.
//
// It was built complete and left unreachable — no module, no routes — while a
// simpler contentless model behind /v1/seed/sprouts served the live path. See
// agent_plan.md §45.
package sow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

type Module struct {
	Sows application.Service
	// Acceptance is exposed because the reviewer surface settles a held sow
	// through the same store that accepted it.
	Acceptance *mongodb.Acceptance
}

// ErrDependenciesRequired reports a module built without the ports it must
// not substitute. Screening decides whether a member's words reach a
// stranger, and ownership decides whose voice a sow carries; a nil either
// would not fail loudly, it would ship the product without them.
var ErrDependenciesRequired = errors.New(
	"sow module requires screening, media ownership, a keying secret and a positive allowance",
)

// NewModule composes the sow against one database.
func NewModule(
	ctx context.Context,
	database *mongo.Database,
	screening application.Screening,
	ownership application.MediaOwnership,
	secret string,
	weeklyUnits int64,
) (Module, error) {
	if screening == nil || ownership == nil || secret == "" || weeklyUnits <= 0 {
		return Module{}, ErrDependenciesRequired
	}
	keyer, err := privacy.New([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	acceptance := mongodb.NewAcceptance(database)
	if err := acceptance.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Sows: application.New(screening, acceptance, keyer, idSource{}, time.Now, weeklyUnits).
			WithMediaOwnership(ownership),
		Acceptance: acceptance,
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		// A weak identifier on a record that costs a member a seed is worse
		// than a crash at startup, which is where this would surface.
		panic(err)
	}
	return "sow_" + hex.EncodeToString(value)
}
