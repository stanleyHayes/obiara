package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
	"strings"
	"time"
)

type Service struct {
	auth Authorizer
	mfa  MFAGate
	keys Keyer
	repo Repository
	now  func() time.Time
}

func New(a Authorizer, m MFAGate, k Keyer, r Repository, n func() time.Time) Service {
	if n == nil {
		n = time.Now
	}
	return Service{a, m, k, r, n}
}
func (s Service) Queue(ctx context.Context, actor string, limit int) ([]Summary, error) {
	if s.auth.Authorize(ctx, actor, CapQueue) != nil {
		return nil, ErrDenied
	}
	cs, e := s.repo.List(ctx, limit)
	if e != nil {
		return nil, ErrUnavailable
	}
	out := make([]Summary, 0, len(cs))
	for _, c := range cs {
		out = append(out, Summary{c.ID(), c.Kind(), c.Status(), c.SummaryCode(), c.Created()})
	}
	return out, nil
}
func (s Service) Detail(ctx context.Context, actor, id string) (Summary, error) {
	if s.auth.Authorize(ctx, actor, CapRead) != nil {
		return Summary{}, ErrDenied
	}
	c, e := s.repo.Find(ctx, id)
	if e != nil {
		return Summary{}, ErrNotFound
	}
	return Summary{c.ID(), c.Kind(), c.Status(), c.SummaryCode(), c.Created()}, nil
}
func (s Service) Evidence(ctx context.Context, actor, id, purpose, correlation string) (Evidence, error) {
	if s.auth.Authorize(ctx, actor, CapEvidence) != nil {
		return Evidence{}, ErrDenied
	}
	now := s.now().UTC()
	if !s.mfa.Recent(ctx, actor, now) {
		return Evidence{}, ErrMFA
	}
	c, e := s.repo.Find(ctx, id)
	if e != nil {
		return Evidence{}, ErrNotFound
	}
	ak, e := s.keys.Key("community_auditor", actor)
	if e != nil {
		return Evidence{}, ErrUnavailable
	}
	ck, e := s.keys.Key("community_correlation", correlation)
	if e != nil {
		return Evidence{}, ErrUnavailable
	}
	pk, e := s.keys.Key("community_purpose", strings.TrimSpace(purpose))
	if e != nil {
		return Evidence{}, ErrUnavailable
	}
	if e = s.repo.RecordEvidenceAccess(ctx, c.ID(), ak, pk+":"+ck, now); e != nil {
		return Evidence{}, ErrUnavailable
	}
	return Evidence{c.EvidenceRef(), c.Kind()}, nil
}
func (s Service) Decide(ctx context.Context, actor, id, command, why, correlation string, approve bool, expected uint64) (domain.Case, error) {
	if s.auth.Authorize(ctx, actor, CapDecide) != nil {
		return domain.Case{}, ErrDenied
	}
	c, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Case{}, ErrNotFound
	}
	ak, e := s.keys.Key("community_auditor", actor)
	if e != nil {
		return domain.Case{}, ErrUnavailable
	}
	ck, e := s.keys.Key("community_correlation", correlation)
	if e != nil {
		return domain.Case{}, ErrUnavailable
	}
	n, e := c.Decide(approve, command, ak, why, ck, s.now(), expected)
	if e != nil {
		return domain.Case{}, e
	}
	if e = s.repo.Decide(ctx, n, c.Version(), command); e != nil {
		return domain.Case{}, ErrConflict
	}
	return n, nil
}
