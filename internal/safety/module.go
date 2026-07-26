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

	"github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/safety/application"
)

type Module struct {
	Safety application.SafetyService
	Cases  application.CaseService
}

func NewModule(ctx context.Context, database *mongo.Database, outboxStore application.OutboxAppender) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	caseRepository := mongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureCaseIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Safety: application.NewSafetyService(repository, repository, outboxStore, time.Now, newID),
		Cases:  application.NewCaseService(caseRepository, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "rep_" + base64.RawURLEncoding.EncodeToString(id)
}
