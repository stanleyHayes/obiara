package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var (
	ErrChallengeNotFound = errors.New("otp challenge not found")
	ErrAccountNotFound   = errors.New("account not found")
	// ErrAccountExists reports a unique-contact conflict at write time (a
	// racing registration); transport layers map it to a conflict without
	// importing persistence types.
	ErrAccountExists = errors.New("account already exists")
	// ErrCodeDeliveryFailed reports that an OTP was minted but no transport
	// in the ladder could deliver it. It is distinct from an internal fault:
	// the member did nothing wrong and retrying will not help until the
	// provider is fixed, so transport layers surface it as an actionable
	// service error rather than a bare 500.
	ErrCodeDeliveryFailed = errors.New("otp code could not be delivered")
	// ErrChannelUnavailable reports a contact whose channel this deployment
	// has no transport for. It is deliberately not a delivery failure: no
	// retry can help, so telling the member to try again shortly would loop
	// them forever on a provider that was never configured.
	ErrChannelUnavailable = errors.New("no otp sender is configured for this channel")
)

// OtpChallengeRepository persists OTP challenges.
type OtpChallengeRepository interface {
	Create(context.Context, domain.OtpChallenge) error
	// LatestByContact returns the most recent challenge for a contact.
	LatestByContact(context.Context, domain.Contact) (domain.OtpChallenge, error)
	// Update persists attempt/consumption changes. The contact+createdAt
	// identity of a challenge never changes.
	Update(context.Context, domain.OtpChallenge) error
}

// AccountRepository persists phone-bound accounts.
type AccountRepository interface {
	Create(context.Context, domain.Account) error
	FindByContact(context.Context, domain.Contact) (domain.Account, error)
	FindByID(context.Context, string) (domain.Account, error)
	// List returns accounts newest first for bounded, redacted operational
	// projections. Callers must never project the stored contact value.
	List(context.Context, int) ([]domain.Account, error)
	// UpdateWithAudit applies a tier transition atomically with its audit
	// record, rejecting stale versions with ErrStaleSession.
	UpdateWithAudit(context.Context, domain.Account, domain.TierTransition) error
	// Update persists status transitions (suspend, block, reactivate).
	Update(context.Context, domain.Account) error
	// ListSuspendedExpired returns suspensions past their end time.
	ListSuspendedExpired(context.Context, time.Time, int) ([]domain.Account, error)
}

// OtpSender is the outbound provider port for OTP delivery. The contact
// carries its own channel, so the adapter behind this port decides which
// transport to use rather than the service guessing: an SMS contact goes to
// the SMS provider, an email contact to the email provider.
type OtpSender interface {
	Send(ctx context.Context, contact domain.Contact, code string) error
}

// RegistrationService runs phone OTP registration and login: challenge
// issuance with rate limits, verification, find-or-create account, and
// session issuance (E03-S01).
type RegistrationService struct {
	challenges OtpChallengeRepository
	accounts   AccountRepository
	sender     OtpSender
	sessions   SessionService
	now        func() time.Time
	newID      func() string
}

func NewRegistrationService(
	challenges OtpChallengeRepository,
	accounts AccountRepository,
	sender OtpSender,
	sessions SessionService,
	now func() time.Time,
	newID func() string,
) RegistrationService {
	return RegistrationService{
		challenges: challenges,
		accounts:   accounts,
		sender:     sender,
		sessions:   sessions,
		now:        now,
		newID:      newID,
	}
}

// OtpRequest is the safe response to an OTP request: never the code.
type OtpRequest struct {
	ChallengeID string
	ExpiresAt   time.Time
}

// RequestOtp issues a new challenge subject to the resend policy.
func (service RegistrationService) RequestOtp(ctx context.Context, contact domain.Contact) (OtpRequest, error) {
	now := service.now()

	var latest *domain.OtpChallenge
	if challenge, err := service.challenges.LatestByContact(ctx, contact); err == nil {
		latest = &challenge
	} else if !errors.Is(err, ErrChallengeNotFound) {
		return OtpRequest{}, err
	}
	if err := domain.CheckResend(latest, now); err != nil {
		return OtpRequest{}, err
	}

	sentCount := 1
	if latest != nil && now.Sub(latest.CreatedAt()) < time.Hour {
		sentCount = latest.SentCount() + 1
	}

	code, err := domain.GenerateCode()
	if err != nil {
		return OtpRequest{}, err
	}
	challenge, err := domain.NewChallenge(service.newID(), contact, code, sentCount, now)
	if err != nil {
		return OtpRequest{}, err
	}
	if err := service.sender.Send(ctx, contact, code); err != nil {
		// The provider cause is kept in the chain so operators can read the
		// real reason in the logs, while callers match on the sentinel. The
		// challenge is deliberately not persisted: a member must never be
		// asked for a code that was never sent.
		return OtpRequest{}, fmt.Errorf("%w: %w", ErrCodeDeliveryFailed, err)
	}
	if err := service.challenges.Create(ctx, challenge); err != nil {
		return OtpRequest{}, err
	}
	return OtpRequest{ChallengeID: challenge.ID(), ExpiresAt: challenge.ExpiresAt()}, nil
}

// VerifyOtp checks a code, consumes the challenge, finds or creates the
// account, and issues a session for the device.
func (service RegistrationService) VerifyOtp(ctx context.Context, contact domain.Contact, code, deviceID string) (IssuedSession, error) {
	challenge, err := service.challenges.LatestByContact(ctx, contact)
	if err != nil {
		return IssuedSession{}, err
	}
	if err := challenge.Verify(code, service.now()); err != nil {
		// Failed attempts must be persisted to bound brute force; a
		// persistence failure here does not leak a successful verification.
		_ = service.challenges.Update(ctx, challenge)
		return IssuedSession{}, err
	}
	if err := service.challenges.Update(ctx, challenge); err != nil {
		return IssuedSession{}, err
	}

	account, err := service.accounts.FindByContact(ctx, contact)
	if errors.Is(err, ErrAccountNotFound) {
		account, err = domain.NewAccount(service.newID(), contact, service.now())
		if err != nil {
			return IssuedSession{}, err
		}
		if err := service.accounts.Create(ctx, account); err != nil {
			return IssuedSession{}, err
		}
	} else if err != nil {
		return IssuedSession{}, err
	}
	if err := account.Usable(); err != nil {
		return IssuedSession{}, err
	}

	return service.sessions.Issue(ctx, account.ID(), deviceID)
}
