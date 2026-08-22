// Package seed composes the seed-stage slices a member meets before a
// courtship room exists: sprouting a doorway toward someone, the bounded
// exchange inside it, declining, and the safety bucket that rate-limits how
// often one member may reach toward others.
//
// They are composed together because they are one stage of the journey and
// share a privacy secret. Every participant and seed reference is keyed
// before it reaches storage, so who reached toward whom is never legible at
// rest.
package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	declinemongo "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/adapters/outbound/mongodb"
	declineprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/adapters/outbound/privacy"
	declineapp "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/application"
	safetymongo "github.com/stanleyHayes/obiara/services/api/internal/seed/safety/adapters/outbound/mongodb"
	safetyprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/safety/adapters/outbound/privacy"
	safetyapp "github.com/stanleyHayes/obiara/services/api/internal/seed/safety/application"
	sproutmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/adapters/outbound/mongodb"
	sproutprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/adapters/outbound/privacy"
	sproutapp "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/application"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("seed module requires a privacy secret")

// StageModule bundles the seed-stage slices.
type StageModule struct {
	Sprout  sproutapp.Service
	Decline declineapp.Service
	Safety  safetyapp.Service
}

// NewStageModule builds every seed-stage slice against one database and
// secret.
func NewStageModule(ctx context.Context, database *mongo.Database, secret string) (StageModule, error) {
	if len(secret) < 32 {
		return StageModule{}, ErrSecretRequired
	}
	key := []byte(secret)

	sproutRepository := sproutmongo.NewRepository(database)
	if err := sproutRepository.EnsureIndexes(ctx); err != nil {
		return StageModule{}, err
	}
	sproutKeyer, err := sproutprivacy.New(key)
	if err != nil {
		return StageModule{}, err
	}

	declineRepository := declinemongo.NewRepository(database)
	if err := declineRepository.EnsureIndexes(ctx); err != nil {
		return StageModule{}, err
	}
	declineKeyer, err := declineprivacy.New(key)
	if err != nil {
		return StageModule{}, err
	}

	safetyRepository := safetymongo.NewRepository(database)
	if err := safetyRepository.EnsureIndexes(ctx); err != nil {
		return StageModule{}, err
	}
	safetyKeyer, err := safetyprivacy.NewKeyer(key)
	if err != nil {
		return StageModule{}, err
	}

	ids := idSource{}
	return StageModule{
		Sprout:  sproutapp.New(sproutRepository, sproutKeyer, ids, time.Now),
		Decline: declineapp.New(declineRepository, declineKeyer, ids, time.Now),
		Safety:  safetyapp.NewService(safetyRepository, safetyKeyer, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "seed_" + base64.RawURLEncoding.EncodeToString(value)
}

// Stage adapts the seed-stage slices to the single inbound port the transport
// layer sees.
type Stage struct{ module StageModule }

func NewStage(module StageModule) Stage { return Stage{module: module} }

func (stage Stage) Sprout(ctx context.Context, command sproutapp.SproutCommand) (sproutapp.SproutResult, error) {
	return stage.module.Sprout.Sprout(ctx, command)
}

func (stage Stage) Exchange(ctx context.Context, command sproutapp.ExchangeCommand) (sproutapp.ExchangeResult, error) {
	result, err := stage.module.Sprout.Exchange(ctx, command)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// A doorway that does not exist must read exactly as one the caller
		// is not part of.
		return sproutapp.ExchangeResult{}, sproutapp.ErrNotFound
	}
	return result, err
}

func (stage Stage) Decline(ctx context.Context, command declineapp.Command) (declineapp.Result, error) {
	return stage.module.Decline.Decline(ctx, command)
}
