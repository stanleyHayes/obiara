// Package profile contains the transport-neutral profile bounded context.
//
// The package intentionally does not expose HTTP handlers or import identity,
// consent, privacy, or authorization implementations. Composition supplies
// outbound ports after those boundaries have made their own access decisions.
package profile

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
)

// Module exposes the profile services currently wired at composition.
type Module struct {
	Profile  application.Service
	Doorway  application.DoorwayService
	Vault    application.VaultService
	profiles *mongodb.Repository
}

// NewModule builds the doorway question and photo vault slice (E03-S09).
// The core profile service (E03-S07) composes separately.
func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	vaultRepository := mongodb.NewVaultRepository(database)
	if err := vaultRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	profileRepository := mongodb.NewRepository(database)
	if err := profileRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Profile:  application.NewService(profileRepository, nil, time.Now),
		Doorway:  application.NewDoorwayService(mongodb.NewDoorwayRepository(database), time.Now),
		Vault:    application.NewVaultService(vaultRepository, time.Now, newID),
		profiles: profileRepository,
	}, nil
}

func (module Module) WithConsent(evaluator application.ConsentEvaluator) Module {
	module.Profile = application.NewService(module.profiles, evaluator, time.Now)
	return module
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "vi_" + base64.RawURLEncoding.EncodeToString(id)
}
