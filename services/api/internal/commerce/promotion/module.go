// Package promotion is the composition root of the discount code.
//
// Named for the thing rather than "voucher", because vouch/assisted already
// owns that word for a person who vouches for another member — which has
// nothing to do with money (agent_plan.md §41).
package promotion

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
)

type Module struct {
	Promotions application.Service
}

// ErrDependenciesRequired reports a module built without the ports that decide
// who may issue a code and under what digest a redemption is recorded. A nil
// issuer check would let codes be minted in a suspended organization's name;
// a nil keyer would record who used a code in the clear.
var ErrDependenciesRequired = errors.New(
	"promotion module requires an issuer check and a member keyer",
)

func NewModule(
	ctx context.Context, database *mongo.Database,
	issuers application.Issuers, keyer application.Keyer,
) (Module, error) {
	if issuers == nil || keyer == nil {
		return Module{}, ErrDependenciesRequired
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Promotions: application.New(repository, issuers, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "promo_" + hex.EncodeToString(value)
}
