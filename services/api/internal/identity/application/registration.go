package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var (
	ErrChallengeNotFound = errors.New("otp challenge not found")
	ErrAccountNotFound   = errors.New("account not found")
	// ErrAccountExists reports a unique-phone conflict at write time (a
	// racing registration); transport layers map it to a conflict without
	// importing persistence types.
	ErrAccountExists = errors.New("account already exists")
)

// OtpChallengeRepository persists OTP challenges.
type OtpChallengeRepository interface {
	Create(context.Context, domain.OtpChallenge) error
	// LatestByPhone returns the most recent challenge for a phone number.
	LatestByPhone(context.Context, string) (domain.OtpChallenge, error)
	// Update persists attempt/consumption changes. The phone+createdAt
	// identity of a challenge never changes.
	Update(context.Context, domain.OtpChallenge) error
}

// AccountRepository persists phone-bound accounts.
type AccountRepository interface {
	Create(context.Context, domain.Account) error
	FindByPhone(context.Context, string) (domain.Account, error)
	FindByID(context.Context, string) (domain.Account, error)
	// UpdateWithAudit applies a tier transition atomically with its audit
	// record, rejecting stale versions with ErrStaleSession.
	UpdateWithAudit(context.Context, domain.Account, domain.TierTransition) error
}

// OtpSender is the outbound provider port for OTP delivery (SMS primary,
// WhatsApp fallback per agent_plan.md §11). Production adapters are scored
// and selected separately; the simulator adapter serves dev and tests.
type OtpSender interface {
	Send(ctx context.Context, phone, code string) error
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
func (service RegistrationService) RequestOtp(ctx context.Context, phone string) (OtpRequest, error) {
	now := service.now()

	var latest *domain.OtpChallenge
	if challenge, err := service.challenges.LatestByPhone(ctx, phone); err == nil {
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
	challenge, err := domain.NewChallenge(service.newID(), phone, code, sentCount, now)
	if err != nil {
		return OtpRequest{}, err
	}
	if err := service.sender.Send(ctx, phone, code); err != nil {
		return OtpRequest{}, err
	}
	if err := service.challenges.Create(ctx, challenge); err != nil {
		return OtpRequest{}, err
	}
	return OtpRequest{ChallengeID: challenge.ID(), ExpiresAt: challenge.ExpiresAt()}, nil
}

// VerifyOtp checks a code, consumes the challenge, finds or creates the
// account, and issues a session for the device.
func (service RegistrationService) VerifyOtp(ctx context.Context, phone, code, deviceID string) (IssuedSession, error) {
	challenge, err := service.challenges.LatestByPhone(ctx, phone)
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

	account, err := service.accounts.FindByPhone(ctx, phone)
	if errors.Is(err, ErrAccountNotFound) {
		account, err = domain.NewAccount(service.newID(), phone, service.now())
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
