// Package introduction owns the consent-bound Voice of Introduction lifecycle.
package introduction

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	mediaadapter "github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/media"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/transcription"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
)

// ConsentPurposeID is the purpose a Voice of Introduction is recorded under.
// It is re-checked on every transition, not only at creation: a member who
// withdraws it must stop the transcription that is already running.
const ConsentPurposeID = "voice.introduction"

type Module struct {
	Introductions application.Service
	Store         *mongodb.Store
}

// ErrDependenciesRequired reports a module built without the ports it cannot
// substitute. Failing at startup is the point: a nil consent gate would mean
// recordings kept without a lawful basis, discovered later.
var ErrDependenciesRequired = errors.New(
	"introduction module requires a consent gate, media access and an asset registry",
)

type idSource struct{}

func (idSource) NewID(prefix string) string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(value)
}

// NewModule composes the Voice of Introduction.
//
// The domain and application layers of this context have been complete and
// tested since S2-021; what was missing was this file. Nothing imported the
// package, no route reached it, and the feature the product is named around —
// a member introduced by their recorded voice — could not run.
func NewModule(
	ctx context.Context,
	database *mongo.Database,
	consent application.ConsentGate,
	access mediaadapter.Access,
	assets mediaadapter.Assets,
	remover mediaadapter.Remover,
	hmacSecret string,
) (Module, error) {
	if consent == nil || access == nil || assets == nil {
		return Module{}, ErrDependenciesRequired
	}
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewHMACKeyer([]byte(hmacSecret))
	if err != nil {
		return Module{}, err
	}
	manager := mediaadapter.NewManager(access, assets, remover, ConsentPurposeID, time.Now)

	return Module{
		Introductions: application.NewService(
			store,
			consent,
			manager,
			// No speech vendor is contracted. The deferred transcriber reports
			// uncertain rather than blocking the recording or inventing words
			// nobody said — the same call already made for liveness.
			transcription.NewDeferred(),
			keyer,
			idSource{},
			time.Now,
		),
		Store: store,
	}, nil
}
