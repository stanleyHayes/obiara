// Package safety is the composition root of the Trust, Safety and Care
// bounded context slice for report/block intake (E12-S01). Triage queues,
// evidence viewers and action ladders land with E12-S02+.
package safety

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/application"
)

type Module struct {
	Safety application.SafetyService
}

func NewModule(ctx context.Context, database *mongo.Database, outboxStore *outbox.Store) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Safety: application.NewSafetyService(repository, repository, outboxStore, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "rep_" + base64.RawURLEncoding.EncodeToString(id)
}
