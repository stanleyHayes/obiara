package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
)

var (
	ErrInvalid     = errors.New("invalid escrow request")
	ErrNotFound    = errors.New("escrow not found")
	ErrConflict    = errors.New("escrow conflict")
	ErrApplied     = errors.New("escrow command applied")
	ErrUnavailable = errors.New("escrow unavailable")
)

type Service struct {
	repo  Repository
	ids   IDSource
	clock Clock
}

func New(r Repository, ids IDSource, c Clock) Service { return Service{r, ids, c} }
func (s Service) ForOwner(ctx context.Context, ownerKey string) ([]domain.Escrow, error) {
	return s.repo.ListForOwner(ctx, ownerKey)
}
func (s Service) FindForOwner(ctx context.Context, id, ownerKey string) (domain.Escrow, error) {
	escrow, err := s.repo.Find(ctx, id)
	if err != nil || escrow.State().OwnerKey != ownerKey {
		return domain.Escrow{}, ErrNotFound
	}
	return escrow, nil
}
func (s Service) Fund(ctx context.Context, ownerKey, engagementID, fundingRef string, amount uint64, terms domain.Terms, command string) (domain.Escrow, error) {
	x, e := domain.Fund(s.ids.NewID(), ownerKey, engagementID, fundingRef, amount, terms, command, s.clock.Now())
	if e != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, x); errors.Is(e, ErrApplied) {
		existing, findErr := s.repo.FindByCommand(ctx, command)
		if findErr == nil && existing.State().OwnerKey == ownerKey {
			return existing, nil
		}
	}
	if e != nil {
		return domain.Escrow{}, e
	}
	return x, nil
}
func (s Service) FundAudited(ctx context.Context, ownerKey, engagementID, fundingRef string, amount uint64, terms domain.Terms, command, actorID string) (domain.Escrow, error) {
	escrow, err := domain.Fund(s.ids.NewID(), ownerKey, engagementID, fundingRef, amount, terms, command, s.clock.Now())
	if err != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if err = s.repo.CreateAudited(ctx, escrow, actorID); errors.Is(err, ErrApplied) {
		existing, findErr := s.repo.FindByCommand(ctx, command)
		if findErr == nil && existing.State().OwnerKey == ownerKey {
			return existing, nil
		}
	}
	if err != nil {
		return domain.Escrow{}, err
	}
	return escrow, nil
}
func (s Service) AddEvidence(ctx context.Context, id, ownerKey, milestone string, role domain.EvidenceRole, command string) (domain.Escrow, error) {
	current, err := s.FindForOwner(ctx, id, ownerKey)
	if err != nil {
		return domain.Escrow{}, err
	}
	return s.mutateOwned(ctx, current, ownerKey, command, func(value domain.Escrow) (domain.Escrow, error) {
		return value.AddEvidence(milestone, role, s.ids.NewID(), command, s.clock.Now())
	})
}
func (s Service) AddEvidenceAudited(ctx context.Context, id, milestone string, role domain.EvidenceRole, command, actorID string) (domain.Escrow, error) {
	current, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Escrow{}, err
	}
	next, err := current.AddEvidence(milestone, role, s.ids.NewID(), command, s.clock.Now())
	if err != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if err = s.repo.SaveAudited(ctx, next, current.Revision(), command, actorID, "admin.escrow.delivery_evidence"); errors.Is(err, ErrApplied) {
		return s.repo.FindByCommand(ctx, command)
	}
	if err != nil {
		return domain.Escrow{}, err
	}
	return next, nil
}
func (s Service) Dispute(ctx context.Context, id, ownerKey, command string) (domain.Escrow, error) {
	current, err := s.FindForOwner(ctx, id, ownerKey)
	if err != nil {
		return domain.Escrow{}, err
	}
	return s.mutateOwned(ctx, current, ownerKey, command, func(value domain.Escrow) (domain.Escrow, error) {
		return value.RaiseDispute(s.ids.NewID(), s.ids.NewID(), command, s.clock.Now())
	})
}
func (s Service) mutateOwned(ctx context.Context, current domain.Escrow, ownerKey, command string, fn func(domain.Escrow) (domain.Escrow, error)) (domain.Escrow, error) {
	next, err := fn(current)
	if err != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if err = s.repo.Save(ctx, next, current.Revision(), command); errors.Is(err, ErrApplied) {
		replayed, findErr := s.repo.FindByCommand(ctx, command)
		if findErr == nil && replayed.State().OwnerKey == ownerKey {
			return replayed, nil
		}
	}
	if err != nil {
		return domain.Escrow{}, err
	}
	return next, nil
}
func (s Service) Mutate(ctx context.Context, id, command string, fn func(domain.Escrow) (domain.Escrow, error)) (domain.Escrow, error) {
	x, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Escrow{}, e
	}
	n, e := fn(x)
	if e != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, x.Revision(), command); e != nil {
		return domain.Escrow{}, e
	}
	return n, nil
}
func (s Service) SettleAudited(ctx context.Context, id, milestone, command, actorID string) (domain.Escrow, domain.Statement, error) {
	x, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Escrow{}, domain.Statement{}, e
	}
	n, statement, e := x.Settle(milestone, s.ids.NewID(), command, s.clock.Now())
	if e != nil {
		return domain.Escrow{}, domain.Statement{}, ErrInvalid
	}
	if e = s.repo.SettleAudited(ctx, n, x.Revision(), command, actorID, statement); errors.Is(e, ErrApplied) {
		replayed, findErr := s.repo.FindByCommand(ctx, command)
		if findErr == nil {
			state := replayed.State()
			var settledAt time.Time
			for _, event := range state.Events {
				if event.CommandID == command {
					settledAt = event.At
					break
				}
			}
			for _, item := range state.Milestones {
				if item.Term.ID == milestone && item.StatementRef != "" {
					return replayed, domain.Statement{
						Ref: item.StatementRef, EscrowID: state.ID, MilestoneID: milestone,
						TermsID: state.TermsID, TermsVersion: state.TermsVersion,
						GrossPesewas: item.Term.GrossPesewas, FeePesewas: item.Term.FeePesewas,
						NetPesewas: item.Term.GrossPesewas - item.Term.FeePesewas,
						SettledAt:  settledAt,
					}, nil
				}
			}
		}
	}
	if e != nil {
		return domain.Escrow{}, domain.Statement{}, e
	}
	return n, statement, nil
}
