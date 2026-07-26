package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("run sheet not found")
	ErrConflict     = errors.New("run sheet conflict")
	ErrApplied      = errors.New("run sheet command applied")
	ErrNotAvailable = errors.New("run sheet not available")
)

type Command struct {
	ID, RunSheetID, FireID, ActorID string
	ExpectedRevision                uint64
}
type Service struct {
	r   Repository
	a   Authority
	k   Keyer
	ids IDSource
	now func() time.Time
}

func NewService(r Repository, a Authority, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, k, ids, now}
}
func (s Service) Create(ctx context.Context, c Command, version uint64, segments []domain.Segment) (domain.RunSheet, error) {
	if !s.ready() || s.a.RequireHostOrCohost(ctx, c.FireID, c.ActorID) != nil {
		return domain.RunSheet{}, ErrNotAvailable
	}
	fire, e := s.k.Key("fire-runsheet:fire", strings.TrimSpace(c.FireID))
	if e != nil {
		return domain.RunSheet{}, ErrNotAvailable
	}
	host, e := s.k.Key("fire-runsheet:host", strings.TrimSpace(c.ActorID))
	if e != nil {
		return domain.RunSheet{}, ErrNotAvailable
	}
	r, e := domain.Create(s.ids.NewID(), fire, host, version, segments, s.command(c))
	if e != nil || s.r.Create(ctx, r) != nil {
		return domain.RunSheet{}, ErrNotAvailable
	}
	return r, nil
}
func (s Service) Start(ctx context.Context, c Command) (domain.Projection, error) {
	return s.change(ctx, c, func(r domain.RunSheet) (domain.RunSheet, error) { return r.Start(s.now().UTC(), s.command(c)) })
}
func (s Service) Advance(ctx context.Context, c Command) (domain.Projection, error) {
	return s.change(ctx, c, func(r domain.RunSheet) (domain.RunSheet, error) { return r.Advance(s.now().UTC(), s.command(c)) })
}
func (s Service) Skip(ctx context.Context, c Command) (domain.Projection, error) {
	return s.change(ctx, c, func(r domain.RunSheet) (domain.RunSheet, error) { return r.Skip(s.now().UTC(), s.command(c)) })
}
func (s Service) Extend(ctx context.Context, c Command, d time.Duration) (domain.Projection, error) {
	return s.change(ctx, c, func(r domain.RunSheet) (domain.RunSheet, error) { return r.Extend(d, s.command(c)) })
}
func (s Service) View(ctx context.Context, c Command) (domain.Projection, error) {
	if !s.ready() || s.a.RequireHostOrCohost(ctx, c.FireID, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	r, e := s.r.Find(ctx, strings.TrimSpace(c.RunSheetID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return r.Project(s.now().UTC()), nil
}
func (s Service) change(ctx context.Context, c Command, f func(domain.RunSheet) (domain.RunSheet, error)) (domain.Projection, error) {
	if !s.ready() || s.a.RequireHostOrCohost(ctx, c.FireID, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	r, e := s.r.Find(ctx, strings.TrimSpace(c.RunSheetID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	fire, e := s.k.Key("fire-runsheet:fire", strings.TrimSpace(c.FireID))
	if e != nil || fire != r.FireKey() {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := f(r)
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	if e = s.r.Append(ctx, next, r.Revision(), c.ID); e != nil {
		if errors.Is(e, ErrApplied) {
			old, x := s.r.FindByCommand(ctx, c.ID)
			if x == nil {
				return old.Project(s.now().UTC()), nil
			}
		}
		return domain.Projection{}, ErrNotAvailable
	}
	return next.Project(s.now().UTC()), nil
}
func (s Service) command(c Command) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) ready() bool { return s.r != nil && s.a != nil && s.k != nil && s.ids != nil }
