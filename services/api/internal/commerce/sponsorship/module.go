// Package sponsorship is the composition root of organization-funded seats.
//
// An organization pre-funds a balance and its sponsored codes draw from it.
// Deposits are recorded by an operator rather than collected by a rail, which
// is why this needs no B2B payment integration to let a university sponsor
// fifty seats (agent_plan.md §78).
package sponsorship

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/application"
)

type Module struct {
	Funds application.Service
}

// ErrIssuersRequired reports a module built without a way to tell whether an
// organization is still a live relationship. Without it, money could be taken
// in the name of a body Obiara has stopped dealing with.
var ErrIssuersRequired = errors.New("sponsorship module requires an issuer check")

func NewModule(
	ctx context.Context, database *mongo.Database,
	issuers application.Issuers, ledger application.Ledger,
) (Module, error) {
	if issuers == nil {
		return Module{}, ErrIssuersRequired
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Funds: application.New(repository, issuers, ledger, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "fund_" + hex.EncodeToString(value)
}
