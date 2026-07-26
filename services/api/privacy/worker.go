// Package privacy exposes the deliberately small worker boundary of the
// API-owned privacy context. Policy and persistence remain internal.
package privacy

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	identitymongodb "github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	privacymongodb "github.com/stanleyHayes/obiara/services/api/internal/privacy/adapters/outbound/mongodb"
	privacyapplication "github.com/stanleyHayes/obiara/services/api/internal/privacy/application"
)

type WorkerProcessor interface {
	RunBatch(context.Context, int) error
}

func NewWorkerProcessor(ctx context.Context, database *mongo.Database) (WorkerProcessor, error) {
	requests := privacymongodb.NewRequestRepository(database)
	if err := requests.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	assembler := privacymongodb.NewArchiveAssembler(database, time.Now)
	if err := assembler.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	sessions := identitymongodb.NewRepository(database)
	if err := sessions.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	sessionService := identityapplication.NewSessionService(sessions, time.Now, func() string {
		return "worker-session-unused"
	})
	processor := privacyapplication.NewProcessor(
		requests, assembler, privacymongodb.NewErasureRunner(database, time.Now),
		sessionService, time.Now,
	)
	return processor, nil
}
