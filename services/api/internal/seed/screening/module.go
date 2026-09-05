// Package screening is the composition root of sow screening.
//
// Under the owner's ruling every sow is read by a person before delivery, so
// this composes an advisor with no opinion and an adjudicator that never
// claims a human decision. The effect is that screening's only outcome is the
// review queue — which is the policy, stated as wiring rather than as a
// comment. See agent_plan.md §49.
package screening

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/deferred"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/locale"
	screeningmedia "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/media"
	screeningmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

type Module struct {
	// Screening is the port the sow context calls.
	Screening application.Adapter
	// Reviews is the queue a person works, exposed so the reviewer surface
	// can list and decide.
	Reviews *screeningmongo.ReviewStore
}

var ErrAssetsRequired = errors.New("screening module requires a media asset reader")

// NewModule composes screening against one database.
//
// reviewed is the list of languages somebody has actually read. It may be
// empty: an unreviewed language routes a sow to a person rather than refusing
// it, so an empty catalog is a safe configuration and not a broken one
// (CL-REG-07 means a tag belongs here only once a human has reviewed it).
func NewModule(
	ctx context.Context,
	database *mongo.Database,
	assets screeningmedia.Assets,
	sourceLocale string,
	reviewed []locale.Reviewed,
) (Module, error) {
	if assets == nil {
		return Module{}, ErrAssetsRequired
	}
	reviews := screeningmongo.NewReviewStore(database, time.Now)
	if err := reviews.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Screening: application.New(
			locale.NewSource(sourceLocale),
			locale.NewCatalog(reviewed...),
			screeningmedia.NewInspector(assets),
			deferred.NewAdvisor(),
			deferred.NewAdjudicator(),
			reviews,
			idSource{},
		),
		Reviews: reviews,
	}, nil
}

// idSource mints the review references screening hands back. They must be
// 64-hex: the adapter validates the shape, and a reference that fails it is a
// review nobody could find again.
type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
