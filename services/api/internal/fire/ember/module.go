// Package ember is the composition root of the ember slice (E06-S10).
// The instant-doorway port for mutual embers stays nil until the sprout
// module composes; mutual status is recorded either way (Doc 06 S-65).
package ember

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/application"
)

type Module struct {
	Embers application.EmberService
}

// NewModule builds the ember service. opener may be nil until the sprout
// module is wired at composition.
func NewModule(ctx context.Context, database *mongo.Database, opener application.DoorwayOpener) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Embers: application.NewEmberService(repository, repository, opener, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "ember_" + base64.RawURLEncoding.EncodeToString(id)
}
