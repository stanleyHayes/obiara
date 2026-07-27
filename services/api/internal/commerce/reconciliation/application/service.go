package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid reconciliation request")
	ErrNotFound = errors.New("reconciliation record not found")
	ErrApplied  = errors.New("reconciliation event already applied")
	ErrConflict = errors.New("reconciliation conflict")
)

type Service struct {
	repo       Repository
	signatures SignatureVerifier
	ledger     Ledger
	keys       Keyer
	ids        IDSource
	clock      Clock
}

func New(r Repository, s SignatureVerifier, l Ledger, k Keyer, ids IDSource, c Clock) Service {
	return Service{repo: r, signatures: s, ledger: l, keys: k, ids: ids, clock: c}
}

func (s Service) Apply(ctx context.Context, envelope SignedEnvelope) (domain.StatementFact, domain.Decision, error) {
	if err := s.signatures.Verify(ctx, envelope); err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	provider, err := s.keys.Key("reconciliation-provider", envelope.Provider)
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	event, err := s.keys.Key("reconciliation-event:"+provider, envelope.EventID)
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	reference, err := s.keys.Key("reconciliation-reference:"+provider, envelope.Reference)
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	fact, err := domain.NewFact(s.ids.NewID(), provider, event, reference, envelope.LedgerCommand, domain.Currency(envelope.Currency), domain.ProviderStatus(envelope.Status), envelope.Minor, time.Unix(envelope.OccurredUnix, 0), s.clock.Now())
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	if err = s.repo.AppendFact(ctx, fact); err != nil {
		if !errors.Is(err, ErrApplied) {
			return domain.StatementFact{}, domain.Decision{}, err
		}
		existing, findErr := s.repo.FindFactByEvent(ctx, event)
		if findErr != nil {
			return domain.StatementFact{}, domain.Decision{}, findErr
		}
		if existing.Fingerprint() != fact.Fingerprint() {
			return domain.StatementFact{}, domain.Decision{}, ErrConflict
		}
		fact = existing
	}
	decision, err := s.compare(ctx, fact)
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, err
	}
	audit, err := domain.NewAudit(s.ids.NewID(), fact, decision, s.clock.Now())
	if err != nil {
		return domain.StatementFact{}, domain.Decision{}, ErrInvalid
	}
	if err = s.repo.AppendAudit(ctx, audit); err != nil && !errors.Is(err, ErrApplied) {
		return domain.StatementFact{}, domain.Decision{}, err
	}
	return fact, decision, nil
}

func (s Service) compare(ctx context.Context, fact domain.StatementFact) (domain.Decision, error) {
	proof, found, err := s.ledger.Proof(ctx, fact.LedgerCommand())
	if err != nil {
		return domain.Decision{}, err
	}
	return domain.Compare(fact, proof, found), nil
}

func (s Service) RunDay(ctx context.Context, day string) (domain.Checkpoint, error) {
	if existing, err := s.repo.FindCheckpoint(ctx, day); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return domain.Checkpoint{}, err
	}
	facts, err := s.repo.ListFactsForDay(ctx, day)
	if err != nil {
		return domain.Checkpoint{}, err
	}
	reconciled, excepted := 0, 0
	for _, fact := range facts {
		decision, compareErr := s.compare(ctx, fact)
		if compareErr != nil {
			return domain.Checkpoint{}, compareErr
		}
		if decision.Outcome() == domain.OutcomeReconciled {
			reconciled++
		}
		if decision.Outcome() == domain.OutcomeException {
			excepted++
		}
		audit, auditErr := domain.NewAudit(s.ids.NewID(), fact, decision, s.clock.Now())
		if auditErr != nil {
			return domain.Checkpoint{}, auditErr
		}
		if auditErr = s.repo.AppendAudit(ctx, audit); auditErr != nil && !errors.Is(auditErr, ErrApplied) {
			return domain.Checkpoint{}, auditErr
		}
	}
	checkpoint, err := domain.NewCheckpoint(s.ids.NewID(), day, len(facts), reconciled, excepted, s.clock.Now())
	if err != nil {
		return domain.Checkpoint{}, ErrInvalid
	}
	if err = s.repo.AppendCheckpoint(ctx, checkpoint); err != nil {
		if !errors.Is(err, ErrApplied) {
			return domain.Checkpoint{}, err
		}
		existing, findErr := s.repo.FindCheckpoint(ctx, day)
		if findErr != nil {
			return domain.Checkpoint{}, findErr
		}
		if existing.Fingerprint() != checkpoint.Fingerprint() {
			return domain.Checkpoint{}, ErrConflict
		}
		return existing, nil
	}
	return checkpoint, nil
}
