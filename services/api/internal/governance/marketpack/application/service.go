package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/domain"
)

var (
	ErrInvalid     = errors.New("invalid market-pack request")
	ErrUnavailable = errors.New("market-pack governance unavailable")
	ErrNotFound    = errors.New("market pack not found")
	ErrConflict    = errors.New("market pack conflict")
	ErrApplied     = errors.New("market-pack command applied")
)

type Service struct {
	masters   MasterCatalog
	authority Authority
	repo      Repository
	ids       IDSource
	clock     Clock
}

func New(m MasterCatalog, a Authority, r Repository, ids IDSource, c Clock) Service {
	return Service{m, a, r, ids, c}
}
func (s Service) Propose(ctx context.Context, author, market, locale string, version uint64, translations []domain.Translation, command string) (domain.Pack, error) {
	if s.authority.RequireAuthor(ctx, author) != nil {
		return domain.Pack{}, ErrUnavailable
	}
	master, e := s.masters.Current(ctx)
	if e != nil {
		return domain.Pack{}, ErrUnavailable
	}
	pack, e := domain.Propose(s.ids.NewID(), market, locale, author, version, master, translations, command, s.clock.Now())
	if e != nil {
		return domain.Pack{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, pack); e != nil {
		return domain.Pack{}, e
	}
	return pack, nil
}
func (s Service) Review(ctx context.Context, packID, reviewer, command string, stage domain.ReviewStage, checks []domain.Check, evidenceRef string) (domain.Pack, error) {
	if s.authority.RequireReviewer(ctx, reviewer, stage) != nil {
		return domain.Pack{}, ErrUnavailable
	}
	pack, e := s.repo.Find(ctx, packID)
	if e != nil {
		return domain.Pack{}, e
	}
	if !s.current(ctx, pack) {
		return domain.Pack{}, ErrUnavailable
	}
	next, e := pack.AddReview(domain.Review{Stage: stage, ReviewerKey: reviewer, Checks: checks, EvidenceRef: evidenceRef, ReviewedAt: s.clock.Now()}, command)
	if e != nil {
		return domain.Pack{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, next, pack.Revision(), command); e != nil {
		return domain.Pack{}, e
	}
	return next, nil
}
func (s Service) Approve(ctx context.Context, packID, approver, reasoningRef, command string) (domain.Pack, error) {
	if s.authority.RequireApprover(ctx, approver) != nil {
		return domain.Pack{}, ErrUnavailable
	}
	pack, e := s.repo.Find(ctx, packID)
	if e != nil {
		return domain.Pack{}, e
	}
	if !s.current(ctx, pack) {
		return domain.Pack{}, ErrUnavailable
	}
	next, e := pack.Approve(approver, reasoningRef, command, s.clock.Now())
	if e != nil {
		return domain.Pack{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, next, pack.Revision(), command); e != nil {
		return domain.Pack{}, e
	}
	return next, nil
}
func (s Service) current(ctx context.Context, pack domain.Pack) bool {
	master, e := s.masters.Current(ctx)
	state := pack.State()
	return e == nil && master.Spec().ID == state.MasterID && master.Spec().Version == state.MasterVersion
}
