// Package identity is the composition root of the Identity and Access
// bounded context (agent_plan.md §7.3). OTP-based registration (E03-S01)
// builds on the session/device kernel.
package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
)

type Module struct {
	Sessions     application.SessionService
	Registration application.RegistrationService
	Tiers        application.TierService
	// Sender is the active OTP delivery adapter, chosen by the composition
	// root from OTP_PROVIDERS.
	Sender application.OtpSender
}

// ErrSenderRequired reports a module built without an OTP transport.
// Registration would mint codes that reach nobody, so this fails at startup
// instead of at the member's first sign-up.
var ErrSenderRequired = errors.New("identity module requires an otp sender")

// NewModule builds the identity slice. sender is the OTP delivery ladder the
// composition root assembled from configuration; it must not be nil.
func NewModule(ctx context.Context, database *mongo.Database, sender application.OtpSender) (Module, error) {
	if sender == nil {
		return Module{}, ErrSenderRequired
	}
	sessionRepository := mongodb.NewRepository(database)
	if err := sessionRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	challengeRepository := mongodb.NewOtpChallengeRepository(database)
	if err := challengeRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	accountRepository := mongodb.NewAccountRepository(database)
	if err := accountRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}

	sessions := application.NewSessionService(sessionRepository, time.Now, newID)
	return Module{
		Sessions:     sessions,
		Registration: application.NewRegistrationService(challengeRepository, accountRepository, sender, sessions, time.Now, newID),
		Tiers:        application.NewTierService(accountRepository, time.Now),
		Sender:       sender,
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		// crypto/rand failure at ID minting is unrecoverable for a secure
		// identity service; panicking beats issuing a weak identifier.
		panic(err)
	}
	return "id_" + base64.RawURLEncoding.EncodeToString(id)
}
