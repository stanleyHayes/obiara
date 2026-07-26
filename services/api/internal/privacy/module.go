// Package privacy is the composition root of the Consent and Privacy
// bounded context slice for data-subject requests (E03-S10). The export
// assembler and erasure runner ports are provided by composition when the
// worker processor is wired; this module ships the member-facing service.
package privacy

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/privacy/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/application"
)

type Module struct {
	Privacy application.PrivacyService
	// Requests exposes the repository for the worker processor build.
	Requests application.RequestRepository
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRequestRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Privacy:  application.NewPrivacyService(repository, repository, time.Now, newID),
		Requests: repository,
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "pr_" + base64.RawURLEncoding.EncodeToString(id)
}
