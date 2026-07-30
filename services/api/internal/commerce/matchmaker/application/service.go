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
func (s Service) Marketplace(ctx context.Context) ([]domain.LicensedProfile, error) {
	profiles, err := s.licenses.ListCurrent(ctx, s.clock.Now())
	if err != nil {
		return nil, ErrUnavailable
	}
	return profiles, nil
}
func (s Service) ForMember(ctx context.Context, memberKey string) ([]domain.Engagement, error) {
	return s.repo.ListForMember(ctx, memberKey)
}
func (s Service) FindForMember(ctx context.Context, id, memberKey string) (domain.Engagement, error) {
	engagement, err := s.repo.Find(ctx, id)
	if err != nil || engagement.State().MemberKey != memberKey {
		return domain.Engagement{}, ErrNotFound
	}
	return engagement, nil
}

// FindForOperations exposes an engagement only to the separately authorized
// operations transport. Member-facing transports must use FindForMember.
func (s Service) FindForOperations(ctx context.Context, id string) (domain.Engagement, error) {
	return s.repo.Find(ctx, id)
}
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
		if errors.Is(e, ErrApplied) {
			existing, findErr := s.repo.FindByCommand(ctx, command)
			if findErr == nil && existing.State().MemberKey == memberKey {
				return existing, nil
			}
		}
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
