package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/domain"
)

var (
	ErrUnavailable = errors.New("P2 companion unavailable")
	ErrApplied     = errors.New("P2 companion command already applied")
	ErrConflict    = errors.New("P2 companion command conflict")
)

type Service struct {
	consent ConsentSource
	facts   CompanionSource
	auth    SessionAuthenticator
	repo    Repository
	ids     IDSource
	clock   Clock
}

func New(c ConsentSource, f CompanionSource, a SessionAuthenticator, r Repository, ids IDSource, clock Clock) Service {
	return Service{consent: c, facts: f, auth: a, repo: r, ids: ids, clock: clock}
}

type ProposeCommand struct {
	CommandID    string
	ActorRef     string
	CourtshipRef string
	ReviewerRef  string
	PackVersion  uint64
	Items        []domain.PackItem
}

func (s Service) ProposeGateLink(ctx context.Context, cmd ProposeCommand) (domain.Proposal, error) {
	if err := s.auth.Authenticate(ctx, cmd.ActorRef, cmd.CourtshipRef); err != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	consent, err := s.consent.CurrentGateConsent(ctx, cmd.CourtshipRef)
	if err != nil {
		return domain.Proposal{}, ErrUnavailable
	}
	proposal, err := domain.Propose(
		s.ids.NewID(), cmd.CommandID, cmd.CourtshipRef, cmd.ReviewerRef,
		s.ids.NewTokenRef(), s.ids.NewWatermarkRef(), cmd.PackVersion, cmd.Items,
		consent, s.clock.Now(),
	)
	if err != nil {
		return domain.Proposal{}, err
	}
	if err := s.repo.Create(ctx, proposal); err != nil {
		return domain.Proposal{}, err
	}
	return proposal, nil
}

func (s Service) ViewUSSD(ctx context.Context, sessionRef, memberRef string) (domain.USSDView, error) {
	if err := s.auth.Authenticate(ctx, sessionRef, memberRef); err != nil {
		return domain.USSDView{}, ErrUnavailable
	}
	facts, err := s.facts.CurrentCompanionFacts(ctx, memberRef)
	if err != nil || facts.MemberRef != memberRef {
		return domain.USSDView{}, ErrUnavailable
	}
	view, err := domain.NewUSSDView(facts, s.clock.Now())
	if err != nil {
		return domain.USSDView{}, ErrUnavailable
	}
	return view, nil
}
