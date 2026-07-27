// Package calls is the composition root of the in-app call slice
// (E09-S09). Tokens come from the LiveKit adapter (S7-001); no phone
// number ever appears in a call flow (FR-304).
package calls

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/calls/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/application"
)

type Module struct {
	Calls application.CallService
}

func NewModule(ctx context.Context, database *mongo.Database, tokens application.TokenIssuer) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Calls: application.NewCallService(repository, mongodb.NewRoomMembership(database), tokens, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "call_" + base64.RawURLEncoding.EncodeToString(id)
}
