// Package verification is the composition root of the Member Profile and
// Verification bounded context slice delivered in E03-S03. Liveness
// orchestration (E03-S04) and the age gate (E03-S05) extend this module.
package verification

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

type Module struct {
	Verification application.VerificationService
	// Provider is the active identity provider adapter, chosen by the
	// composition root from IDENTITY_VERIFICATION_PROVIDER.
	Provider application.VerificationProvider
}

// ErrProviderRequired reports a module built with no identity provider. A
// nil provider cannot decide a case, so this fails at startup rather than at
// a member's first verification attempt.
var ErrProviderRequired = errors.New("verification module requires an identity provider")

// NewModule builds the verification context. provider decides cases and must
// not be nil; tiers bridges to the identity context's tier state machine.
func NewModule(ctx context.Context, database *mongo.Database, provider application.VerificationProvider, tiers application.TierTransitions, cardSecret []byte) (Module, error) {
	if provider == nil {
		return Module{}, ErrProviderRequired
	}
	keyer, err := privacy.NewHMACKeyer(cardSecret)
	if err != nil {
		return Module{}, err
	}
	caseRepository := mongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	documents := mongodb.NewDocumentRepository(database)
	if err := documents.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	// Card photographs are sealed under a key derived from the same secret but
	// a different label, so the identity store and the liveness store cannot
	// be opened with each other's key.
	sealer, err := privacy.NewAESGCMSealer(cardSecret)
	if err != nil {
		return Module{}, err
	}
	return Module{
		Verification: application.
			NewVerificationService(caseRepository, provider, tiers, keyer, time.Now, newID).
			WithDocuments(documents, sealer, sealer),
		Provider: provider,
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "vc_" + base64.RawURLEncoding.EncodeToString(id)
}
