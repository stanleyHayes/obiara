// Package identity is the composition root of the Identity and Access
// bounded context (agent_plan.md §7.3). OTP-based registration builds on
// this session/device kernel in E03-S01.
package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
)

type Module struct {
	Sessions application.SessionService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Sessions: application.NewSessionService(repository, time.Now, newSessionID),
	}, nil
}

func newSessionID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		// crypto/rand failure at ID minting is unrecoverable for a secure
		// session service; panicking beats issuing a weak identifier.
		panic(err)
	}
	return "sess_" + base64.RawURLEncoding.EncodeToString(id)
}
