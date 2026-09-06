// Package media owns authorized, time-limited access to member media.
package media

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/objectstore"
	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/sharingpolicy"
	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

type Module struct {
	Access application.AccessService
	Assets *mongodb.AssetRepository
	// Objects erases stored bytes. Exposed so retention can remove the object
	// itself; deleting only the row would leave the audio in the bucket with
	// nothing left that knows its key.
	Objects *objectstore.Signer
}

// NewModule composes media access against an S3-compatible bucket.
//
// purposes is the closed list this deployment will authorize. It is passed in
// rather than defaulted because the policy refuses anything not on it, and a
// default would decide, silently and elsewhere, who may read member media.
//
// entitlements, keyed by purpose, say who other than the owner may hear a
// recording under that purpose. A purpose with no entitlement stays
// owner-only. Passing none reproduces the behaviour this module had before
// there were any — which is correct as a floor and was wrong as the whole
// rule: with owner-only as the only policy, nobody could hear anybody, and
// the delivery chain ended in a refusal (agent_plan.md §63).
func NewModule(
	ctx context.Context,
	database *mongo.Database,
	storage objectstore.Config,
	purposes []string,
	entitlements map[string]sharingpolicy.Entitlement,
) (Module, error) {
	assets := mongodb.NewAssetRepository(database)
	if err := assets.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	signer, err := objectstore.NewSigner(storage, time.Now)
	if err != nil {
		return Module{}, err
	}
	return Module{
		Access: application.NewAccessService(
			assets, sharingpolicy.New(purposes, entitlements), signer, time.Now,
		),
		Assets:  assets,
		Objects: signer,
	}, nil
}
