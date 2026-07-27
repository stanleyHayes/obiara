package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/domain"
)

var (
	ErrInvalid     = errors.New("invalid community operation request")
	ErrUnavailable = errors.New("community operation unavailable")
	ErrNotFound    = errors.New("community operation proposal not found")
	ErrConflict    = errors.New("community operation proposal conflict")
	ErrApplied     = errors.New("community operation command applied")
)

type Service struct {
	authority Authority
	densities DensitySource
	hosts     HostSource
	notices   NoticeCatalog
	repo      Repository
	ids       IDSource
	clock     Clock
}

func New(a Authority, d DensitySource, h HostSource, n NoticeCatalog, r Repository, i IDSource, c Clock) Service {
	return Service{a, d, h, n, r, i, c}
}

type ProposeRequest struct {
	ActorKey, CircleKey, FireKey, HostKey     string
	Operation                                 domain.Operation
	Reason                                    domain.Reason
	ReasonRef, TemplateKey, Locale, CommandID string
}

func (s Service) Propose(ctx context.Context, req ProposeRequest) (domain.Proposal, error) {
	if s.authority.RequireCommunityOperator(ctx, req.ActorKey) != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	density, e := s.densities.CurrentDensity(ctx, req.CircleKey, req.FireKey)
	if e != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	var host *domain.HostEligibility
	if req.Operation == domain.AssignHost || req.Operation == domain.ReplaceHost {
		current, e := s.hosts.CurrentEligibility(ctx, req.HostKey)
		if e != nil {
			return domain.Proposal{}, ErrUnavailable
		}
		host = &current
	}
	notice, e := s.notices.CurrentNotice(ctx, req.TemplateKey, req.Locale, density.Participants)
	if e != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	p, e := domain.Propose(s.ids.NewID(), req.ActorKey, req.Operation, req.Reason, req.ReasonRef, density, host, notice, req.CommandID, s.clock.Now())
	if e != nil {
		return domain.Proposal{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, p); e != nil {
		return domain.Proposal{}, e
	}
	return p, nil
}
func (s Service) AcknowledgeNotice(ctx context.Context, id, actor, digest, command string) (domain.Proposal, error) {
	if s.authority.RequireCommunityOperator(ctx, actor) != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	p, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Proposal{}, e
	}
	state := p.State()
	density, e := s.densities.CurrentDensity(ctx, state.Density.CircleKey, state.Density.FireKey)
	if e != nil || density.Version != state.Density.Version {
		return domain.Proposal{}, ErrUnavailable
	}
	if state.Host != nil {
		host, e := s.hosts.CurrentEligibility(ctx, state.Host.HostKey)
		if e != nil || host.VerificationVersion != state.Host.VerificationVersion || host.CertificationVersion != state.Host.CertificationVersion {
			return domain.Proposal{}, ErrUnavailable
		}
	}
	notice, e := s.notices.CurrentNotice(ctx, state.Notice.TemplateKey, state.Notice.Locale, state.Notice.AudienceCount)
	if e != nil || notice.Digest != state.Notice.Digest {
		return domain.Proposal{}, ErrUnavailable
	}
	next, e := p.AcknowledgeNotice(actor, digest, command, s.clock.Now())
	if e != nil {
		return domain.Proposal{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, next, p.Revision(), command); e != nil {
		return domain.Proposal{}, e
	}
	return next, nil
}
