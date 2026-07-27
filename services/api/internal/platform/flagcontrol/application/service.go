package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
)

var (
	ErrNotFound = errors.New("flag control not found")
	ErrApplied  = errors.New("flag control already applied")
	ErrConflict = errors.New("flag control conflict")
)

type Service struct {
	repo      Repository
	authority Authority
	runtime   Runtime
	ids       IDSource
	clock     Clock
}

func New(r Repository, a Authority, rt Runtime, ids IDSource, c Clock) Service {
	return Service{r, a, rt, ids, c}
}

type ProposeCommand struct {
	CommandID, SessionID string
	Capability           domain.Capability
	Environment          domain.Environment
	Market               domain.Market
	Action               domain.Action
	Reason               domain.Reason
}

func (s Service) Propose(ctx context.Context, c ProposeCommand) (domain.Proposal, error) {
	actor, err := s.authority.RequireSteppedController(ctx, c.SessionID, c.Capability)
	if err != nil {
		return domain.Proposal{}, err
	}
	now := s.clock.Now()
	p, err := domain.NewProposal(s.ids.NewID(), c.CommandID, actor, c.Capability, c.Environment, c.Market, c.Action, c.Reason, now, now.Add(domain.MaxLifetime))
	if err != nil {
		return domain.Proposal{}, err
	}
	audit, _ := domain.NewAudit(s.ids.NewID(), p, actor, domain.AuditProposed, now)
	if err = s.repo.CreateWithAudit(ctx, p, audit); err != nil {
		if !errors.Is(err, ErrApplied) {
			return domain.Proposal{}, err
		}
		existing, e := s.repo.FindByCommand(ctx, c.CommandID)
		if e != nil {
			return domain.Proposal{}, e
		}
		if existing.Fingerprint() != p.Fingerprint() {
			return domain.Proposal{}, ErrConflict
		}
		return existing, nil
	}
	return p, nil
}
func (s Service) Approve(ctx context.Context, id, session string) (domain.Proposal, error) {
	p, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Proposal{}, err
	}
	actor, err := s.authority.RequireSteppedController(ctx, session, p.Capability())
	if err != nil {
		return domain.Proposal{}, err
	}
	next, err := p.Approve(actor, s.clock.Now())
	if err != nil {
		return domain.Proposal{}, err
	}
	audit, _ := domain.NewAudit(s.ids.NewID(), next, actor, domain.AuditApproved, s.clock.Now())
	if err = s.repo.SaveWithAudit(ctx, next, p.Version(), audit); err != nil {
		return domain.Proposal{}, err
	}
	return next, nil
}
func (s Service) Apply(ctx context.Context, id, session string) (domain.Proposal, error) {
	p, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Proposal{}, err
	}
	actor, err := s.authority.RequireSteppedController(ctx, session, p.Capability())
	if err != nil {
		return domain.Proposal{}, err
	}
	next, change, err := p.Apply(s.clock.Now())
	if err != nil {
		return domain.Proposal{}, err
	}
	st := p.State()
	if actor != st.ApproverKey {
		return domain.Proposal{}, domain.ErrInvalid
	}
	if err = s.runtime.Apply(ctx, st.Environment, st.Market, change); err != nil {
		return domain.Proposal{}, err
	}
	audit, _ := domain.NewAudit(s.ids.NewID(), next, actor, domain.AuditApplied, s.clock.Now())
	if err = s.repo.SaveWithAudit(ctx, next, p.Version(), audit); err != nil {
		return domain.Proposal{}, err
	}
	return next, nil
}
func (s Service) Expire(ctx context.Context, id string) (domain.Proposal, error) {
	p, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Proposal{}, err
	}
	next, change, err := p.Expire(s.clock.Now())
	if err != nil {
		return domain.Proposal{}, err
	}
	st := p.State()
	if err = s.runtime.Apply(ctx, st.Environment, st.Market, change); err != nil {
		return domain.Proposal{}, err
	}
	audit, _ := domain.NewAudit(s.ids.NewID(), next, st.ProposerKey, domain.AuditExpired, s.clock.Now())
	if err = s.repo.SaveWithAudit(ctx, next, p.Version(), audit); err != nil {
		return domain.Proposal{}, err
	}
	return next, nil
}
