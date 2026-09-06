// Package affiliate is the composition root of the referral scheme.
//
// Affiliates are outside parties — organizations, creators, campus reps.
// Members are never affiliates (agent_plan.md §41): paying somebody inside the
// community to recruit changes what "why is this person talking to me" means.
//
// Composed only when a commission and a withholding rate are both configured.
// A scheme that accrues but can never legally pay out is a liability that only
// grows, so an unset rate leaves the whole thing absent.
package affiliate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/application"
)

type Module struct {
	Affiliates application.Service
	Payouts    application.PayoutService
}

var ErrDependenciesRequired = errors.New(
	"affiliate module requires a keyer, a qualification check, a transfer rail, " +
		"a positive commission and a withholding rate",
)

// Settings is the scheme's commercial shape, which is a money decision rather
// than an engineering one and so is passed in rather than defaulted.
type Settings struct {
	CommissionPesewas      int64
	MinimumPayoutPesewas   int64
	WithholdingBasisPoints int64
}

func NewModule(
	ctx context.Context, database *mongo.Database,
	qualification application.Qualification, transfers application.Transfers,
	keyer application.Keyer, settings Settings,
) (Module, error) {
	if qualification == nil || transfers == nil || keyer == nil ||
		settings.CommissionPesewas <= 0 ||
		settings.WithholdingBasisPoints <= 0 || settings.WithholdingBasisPoints >= 10_000 {
		return Module{}, ErrDependenciesRequired
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	payouts := mongodb.NewPayouts(database)
	if err := payouts.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Affiliates: application.New(
			repository, repository, qualification, nil, keyer, idSource{"aff"},
			settings.CommissionPesewas, time.Now,
		),
		Payouts: application.NewPayoutService(
			repository, payouts, transfers, nil, idSource{"payout"},
			settings.MinimumPayoutPesewas, settings.WithholdingBasisPoints, time.Now,
		),
	}, nil
}

type idSource struct{ prefix string }

func (source idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return source.prefix + "_" + hex.EncodeToString(value)
}
