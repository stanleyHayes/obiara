// Package application orchestrates Ghana Card verification (E03-S03):
// provider decision with timeout, fallback to the human queue on
// uncertainty or outage (never a silent pass), and tier promotion on
// approval via the identity context port.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

var (
	ErrCaseNotFound     = errors.New("verification case not found")
	ErrProviderRejected = errors.New("identity provider rejected the document")
	// ErrIdentityAlreadyVerified reports a card that is already the verified
	// identity of a different account.
	//
	// FR-102 is "exactly one active account per verified identity", and the
	// verified identity is the document, not the phone number or the email
	// address it was submitted from. Without this, one person could hold a
	// Tier 1 account per contact they control — which is also how a blocked
	// member walks back onto the romantic surfaces, since blocking is per
	// account.
	ErrIdentityAlreadyVerified = errors.New("this identity is already verified on another account")
)

// ProviderRequest is the provider-neutral verification submission. Provider
// SDK types never cross this port (agent_plan.md §7.2).
type ProviderRequest struct {
	CaseID      string
	CardNumber  string
	DateOfBirth time.Time
}

// ProviderResult is the provider-neutral decision.
type ProviderResult struct {
	// Outcome: "match", "mismatch", "uncertain".
	Outcome     string
	ProviderRef string
	Reason      string
}

// VerificationProvider is the outbound port for the national-ID issuer
// service. Adapters implement timeout, retry, circuit-breaker and
// idempotency rules (agent_plan.md §11).
type VerificationProvider interface {
	Verify(context.Context, ProviderRequest) (ProviderResult, error)
}

// CaseRepository persists verification cases.
type CaseRepository interface {
	Create(context.Context, domain.VerificationCase) error
	FindByID(context.Context, string) (domain.VerificationCase, error)
	Update(context.Context, domain.VerificationCase) error
	// NextQueued returns the oldest manual-queue cases for the fallback desk.
	NextQueued(context.Context, int) ([]domain.VerificationCase, error)
	// ApprovedAccountByCardKey returns the account already verified with this
	// card, or ErrCaseNotFound when the identity is unclaimed.
	ApprovedAccountByCardKey(ctx context.Context, cardKey string) (string, error)
}

// CardKeyer derives the stable, non-reversible key a card is recognised by.
type CardKeyer interface {
	Key(value string) (string, error)
}

// TierTransitions is the identity-context port used to promote accounts on
// approval (E03-S06 state machine owns the transition rules).
type TierTransitions interface {
	Transition(ctx context.Context, accountID string, target int, reason, actorID string) error
}

// VerificationService runs the submission → decision → tier flow.
type VerificationService struct {
	cases    CaseRepository
	provider VerificationProvider
	tiers    TierTransitions
	keyer    CardKeyer
	now      func() time.Time
	newID    func() string
}

func NewVerificationService(cases CaseRepository, provider VerificationProvider, tiers TierTransitions, keyer CardKeyer, now func() time.Time, newID func() string) VerificationService {
	return VerificationService{cases: cases, provider: provider, tiers: tiers, keyer: keyer, now: now, newID: newID}
}

// SubmitGhanaCard opens a case and asks the provider. Provider outage or
// uncertainty routes the case to the human fallback queue; the member is
// never silently failed or passed (FR-103).
func (service VerificationService) SubmitGhanaCard(ctx context.Context, accountID, cardNumber string, dateOfBirth time.Time) (domain.VerificationCase, error) {
	cardKey, err := service.keyer.Key(cardNumber)
	if err != nil {
		return domain.VerificationCase{}, err
	}

	// Refuse before opening a case, so a member who is trying to hold a
	// second account learns immediately rather than after a provider round
	// trip. The unique index behind Update is what actually guarantees it —
	// this check exists to give a clear answer, not to be the enforcement.
	switch owner, lookupErr := service.cases.ApprovedAccountByCardKey(ctx, cardKey); {
	case lookupErr == nil && owner != accountID:
		return domain.VerificationCase{}, ErrIdentityAlreadyVerified
	case lookupErr != nil && !errors.Is(lookupErr, ErrCaseNotFound):
		return domain.VerificationCase{}, lookupErr
	}

	verificationCase, err := domain.NewCase(service.newID(), accountID, cardKey, maskCard(cardNumber), dateOfBirth, service.now())
	if err != nil {
		return domain.VerificationCase{}, err
	}
	if err := service.cases.Create(ctx, verificationCase); err != nil {
		return domain.VerificationCase{}, err
	}
	// The plaintext travels to the provider in memory and is never stored.
	return service.decideWithProvider(ctx, verificationCase, cardNumber)
}

// maskCard keeps the last four digits, which is all the review desk ever
// displayed of a card.
func maskCard(card string) string {
	if len(card) < 4 {
		return "••••"
	}
	return "•••• " + card[len(card)-4:]
}

func (service VerificationService) decideWithProvider(ctx context.Context, verificationCase domain.VerificationCase, cardNumber string) (domain.VerificationCase, error) {
	result, err := service.provider.Verify(ctx, ProviderRequest{
		CaseID:      verificationCase.ID(),
		CardNumber:  cardNumber,
		DateOfBirth: verificationCase.DateOfBirth(),
	})
	if err != nil {
		return service.queueManual(ctx, verificationCase, "provider unavailable")
	}

	switch result.Outcome {
	case "match":
		if err := verificationCase.Approve(result.ProviderRef, result.Reason, service.now()); err != nil {
			return domain.VerificationCase{}, err
		}
		if err := service.cases.Update(ctx, verificationCase); err != nil {
			return domain.VerificationCase{}, err
		}
		// Tier promotion failures do not reopen the approved case; the
		// transition is retried by reconciliation. The error is surfaced so
		// operators can observe it.
		if err := service.tiers.Transition(ctx, verificationCase.AccountID(), 1, "ghana card verified", "verification:"+verificationCase.ID()); err != nil {
			return verificationCase, err
		}
		return verificationCase, nil
	case "mismatch":
		if err := verificationCase.Reject(result.ProviderRef, result.Reason, service.now()); err != nil {
			return domain.VerificationCase{}, err
		}
		if err := service.cases.Update(ctx, verificationCase); err != nil {
			return domain.VerificationCase{}, err
		}
		return verificationCase, ErrProviderRejected
	default:
		return service.queueManual(ctx, verificationCase, "provider uncertain")
	}
}

func (service VerificationService) queueManual(ctx context.Context, verificationCase domain.VerificationCase, reason string) (domain.VerificationCase, error) {
	if err := verificationCase.QueueForManualReview(reason, service.now()); err != nil {
		return domain.VerificationCase{}, err
	}
	if err := service.cases.Update(ctx, verificationCase); err != nil {
		return domain.VerificationCase{}, err
	}
	return verificationCase, nil
}

// ManualQueue lists the oldest cases awaiting human review (fallback desk
// skeleton; the admin console lands with E03-S11).
func (service VerificationService) ManualQueue(ctx context.Context, limit int) ([]domain.VerificationCase, error) {
	return service.cases.NextQueued(ctx, limit)
}

// DecideManual records a human fallback-desk decision on a queued case.
func (service VerificationService) DecideManual(ctx context.Context, caseID string, approve bool, reason, actorID string) (domain.VerificationCase, error) {
	verificationCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return domain.VerificationCase{}, err
	}
	if approve {
		if err := verificationCase.Approve("manual:"+actorID, reason, service.now()); err != nil {
			return domain.VerificationCase{}, err
		}
		if err := service.cases.Update(ctx, verificationCase); err != nil {
			return domain.VerificationCase{}, err
		}
		if err := service.tiers.Transition(ctx, verificationCase.AccountID(), 1, "manual verification approved", actorID); err != nil {
			return verificationCase, err
		}
		return verificationCase, nil
	}
	if err := verificationCase.Reject("manual:"+actorID, reason, service.now()); err != nil {
		return domain.VerificationCase{}, err
	}
	return verificationCase, service.cases.Update(ctx, verificationCase)
}
