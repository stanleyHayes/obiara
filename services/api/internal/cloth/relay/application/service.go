package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/domain"
	"time"
)

type Service struct {
	repo    Repository
	keys    Keyer
	auth    ReviewerAuthorization
	consent ConsentRevalidator
	now     func() time.Time
}

func NewService(r Repository, k Keyer, a ReviewerAuthorization, c ConsentRevalidator, n func() time.Time) Service {
	return Service{r, k, a, c, n}
}
func (s Service) load(ctx context.Context, pairID string) (domain.Relay, string, error) {
	pk, e := s.keys.Key("relay_pair", pairID)
	if e != nil {
		return domain.Relay{}, "", e
	}
	r, e := s.repo.Find(ctx, pk)
	return r, pk, e
}
func (s Service) Submit(ctx context.Context, pairID, reviewerID, cmdID, qref, pref string, expected uint64) (domain.Relay, error) {
	r, pk, e := s.load(ctx, pairID)
	if e != nil {
		return domain.Relay{}, e
	}
	rk, e := s.keys.Key("relay_reviewer", reviewerID)
	if e != nil {
		return domain.Relay{}, e
	}
	ok, e := s.auth.Allowed(ctx, pk, rk)
	if e != nil || !ok {
		return domain.Relay{}, domain.ErrDenied
	}
	return s.apply(ctx, r, r.Submit, domain.Command{ID: cmdID, ActorKey: rk, QuestionRef: qref, PromptRef: pref, ExpectedRevision: expected, At: s.now()})
}
func (s Service) Grant(ctx context.Context, pairID, memberID, cmdID, qref, rref string, expected uint64) (domain.Relay, error) {
	r, _, e := s.load(ctx, pairID)
	if e != nil {
		return domain.Relay{}, e
	}
	mk, e := s.keys.Key("relay_member", memberID)
	if e != nil {
		return domain.Relay{}, e
	}
	return s.apply(ctx, r, r.Grant, domain.Command{ID: cmdID, ActorKey: mk, QuestionRef: qref, ResponseRef: rref, ExpectedRevision: expected, At: s.now()})
}
func (s Service) Revoke(ctx context.Context, pairID, memberID, cmdID, qref string, expected uint64) (domain.Relay, error) {
	r, _, e := s.load(ctx, pairID)
	if e != nil {
		return domain.Relay{}, e
	}
	mk, e := s.keys.Key("relay_member", memberID)
	if e != nil {
		return domain.Relay{}, e
	}
	return s.apply(ctx, r, r.Revoke, domain.Command{ID: cmdID, ActorKey: mk, QuestionRef: qref, ExpectedRevision: expected, At: s.now()})
}
func (s Service) apply(ctx context.Context, r domain.Relay, fn func(domain.Command) (domain.Relay, error), c domain.Command) (domain.Relay, error) {
	n, e := fn(c)
	if e != nil {
		return domain.Relay{}, e
	}
	if n.Revision() != r.Revision() {
		e = s.repo.Save(ctx, n, r.Revision(), c.ID)
	}
	return n, e
}
func (s Service) Project(ctx context.Context, pairID, reviewerID, qref string) (domain.Projection, error) {
	r, pk, e := s.load(ctx, pairID)
	if e != nil {
		return domain.Projection{}, e
	}
	rk, e := s.keys.Key("relay_reviewer", reviewerID)
	if e != nil {
		return domain.Projection{}, e
	}
	ok, e := s.auth.Allowed(ctx, pk, rk)
	if e != nil || !ok {
		return domain.Projection{}, domain.ErrDenied
	}
	p, e := r.Project(rk, qref)
	if e != nil {
		return domain.Projection{}, e
	}
	ok, e = s.consent.Current(ctx, pk, qref, p.ResponseRef, r.Members())
	if e != nil || !ok {
		return domain.Projection{}, domain.ErrDenied
	}
	return p, nil
}
