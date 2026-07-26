package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	"strings"
	"time"
)

type Service struct {
	auth  Authorizer
	repo  Repository
	keyer Keyer
	ids   IDs
	now   func() time.Time
}

func NewService(a Authorizer, r Repository, k Keyer, i IDs, n func() time.Time) Service {
	if n == nil {
		n = time.Now
	}
	return Service{a, r, k, i, n}
}

type Create struct {
	CommandID, CircleID, ActorID, ContentRef string
	Media                                    domain.MediaRef
	StartsAt, EndsAt                         time.Time
	Retention                                time.Duration
}

func (s Service) Voice(ctx context.Context, r Create) (domain.Entry, error) {
	return s.create(ctx, r, domain.KindVoice, CapabilityPost)
}
func (s Service) Event(ctx context.Context, r Create) (domain.Entry, error) {
	return s.create(ctx, r, domain.KindEvent, CapabilityHost)
}
func (s Service) Notice(ctx context.Context, r Create) (domain.Entry, error) {
	return s.create(ctx, r, domain.KindNotice, CapabilityHost)
}
func (s Service) create(ctx context.Context, r Create, k domain.Kind, c Capability) (domain.Entry, error) {
	if s.auth == nil || s.repo == nil || s.keyer == nil || s.ids == nil {
		return domain.Entry{}, ErrUnavailable
	}
	if s.auth.Authorize(ctx, Decision{r.CircleID, r.ActorID, c}) != nil {
		return domain.Entry{}, ErrDenied
	}
	key, err := s.keyer.Key("circle_room_actor", strings.TrimSpace(r.ActorID))
	if err != nil {
		return domain.Entry{}, ErrUnavailable
	}
	now := s.now().UTC()
	e, err := domain.New(domain.Params{ID: s.ids.NewID(), CircleID: r.CircleID, AuthorKey: key, ContentRef: r.ContentRef, CommandID: r.CommandID, Kind: k, Media: r.Media, StartsAt: r.StartsAt, EndsAt: r.EndsAt, CreatedAt: now, ExpiresAt: now.Add(r.Retention)})
	if err != nil {
		return domain.Entry{}, err
	}
	got, _, err := s.repo.Create(ctx, e)
	if err != nil {
		return domain.Entry{}, ErrUnavailable
	}
	return got, nil
}
func (s Service) List(ctx context.Context, circle, actor string, limit int) ([]domain.Entry, error) {
	if s.auth.Authorize(ctx, Decision{circle, actor, CapabilityRead}) != nil {
		return nil, ErrDenied
	}
	return s.repo.List(ctx, circle, s.now().UTC(), limit)
}
func (s Service) Delete(ctx context.Context, id, actor, command string) (domain.Entry, error) {
	e, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Entry{}, ErrNotFound
	}
	if s.auth.Authorize(ctx, Decision{e.CircleID(), actor, CapabilityHost}) != nil {
		return domain.Entry{}, ErrDenied
	}
	key, err := s.keyer.Key("circle_room_actor", actor)
	if err != nil {
		return domain.Entry{}, ErrUnavailable
	}
	next, err := e.Delete(command, key, s.now(), e.Revision())
	if err != nil {
		return domain.Entry{}, err
	}
	if err = s.repo.Delete(ctx, next, e.Revision(), command); err != nil {
		return domain.Entry{}, ErrConflict
	}
	return next, nil
}
