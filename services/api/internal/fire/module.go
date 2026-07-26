// Package fire is the composition root of the Fires and Realtime bounded
// context slice for scheduling and attendance (E09-S01). LiveKit room and
// host control adapters land with E09-S02/S03.
package fire

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/application"
)

type Module struct {
	Fires application.FireService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Fires: application.NewFireService(repository, time.Now, newID)}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "fire_" + base64.RawURLEncoding.EncodeToString(id)
}
