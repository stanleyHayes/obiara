// Package verification is the composition root of the Member Profile and
// Verification bounded context slice delivered in E03-S03. Liveness
// orchestration (E03-S04) and the age gate (E03-S05) extend this module.
package verification

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

type Module struct {
	Verification application.VerificationService
	// Provider is the active identity provider adapter (simulator until a
	// scored Ghana Card vendor is selected).
	Provider application.VerificationProvider
}

// NewModule builds the verification context. tiers bridges to the identity
// context's tier state machine; composition roots provide it.
func NewModule(ctx context.Context, database *mongo.Database, tiers application.TierTransitions) (Module, error) {
	caseRepository := mongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	provider := simulator.NewProvider()
	return Module{
		Verification: application.NewVerificationService(caseRepository, provider, tiers, time.Now, newID),
		Provider:     provider,
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "vc_" + base64.RawURLEncoding.EncodeToString(id)
}
