package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
)

var (
	ErrInvalid     = errors.New("invalid matchmaker request")
	ErrNotFound    = errors.New("engagement not found")
	ErrConflict    = errors.New("engagement conflict")
	ErrApplied     = errors.New("command applied")
	ErrUnavailable = errors.New("matchmaker unavailable")
)

type Service struct {
	repo     Repository
	licenses LicenseCatalog
	ids      IDSource
	clock    Clock
}

func New(r Repository, l LicenseCatalog, ids IDSource, c Clock) Service { return Service{r, l, ids, c} }
func (s Service) Book(ctx context.Context, memberKey, matchmakerKey string, terms domain.Terms, command string) (domain.Engagement, error) {
	license, e := s.licenses.Current(ctx, matchmakerKey)
	if e != nil || license.MatchmakerKey != matchmakerKey {
		return domain.Engagement{}, ErrUnavailable
	}
	x, e := domain.Book(s.ids.NewID(), memberKey, license, terms, command, s.clock.Now())
	if e != nil {
		return domain.Engagement{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, x); e != nil {
		return domain.Engagement{}, e
	}
	return x, nil
}
func (s Service) Mutate(ctx context.Context, id, command string, fn func(domain.Engagement) (domain.Engagement, error)) (domain.Engagement, error) {
	x, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Engagement{}, e
	}
	license, e := s.licenses.Current(ctx, x.State().MatchmakerKey)
	if e != nil || !license.Current(s.clock.Now()) || license.ID != x.State().LicenseID || license.Version != x.State().LicenseVersion {
		return domain.Engagement{}, ErrUnavailable
	}
	n, e := fn(x)
	if e != nil {
		return domain.Engagement{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, x.Revision(), command); e != nil {
		return domain.Engagement{}, e
	}
	return n, nil
}
