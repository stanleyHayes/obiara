package application

import (
	"context"
	"errors"
	session "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("oware session not found")
	ErrConflict     = errors.New("oware session conflict")
	ErrApplied      = errors.New("oware session command applied")
	ErrNotAvailable = errors.New("oware session not available")
)

type Command struct {
	ID, SessionID, RoomID, ActorID string
	ExpectedRevision               uint64
}
type Service struct {
	r     Repository
	rooms RoomEmbedding
	a     Authorizer
	k     Keyer
	ids   IDSource
	now   func() time.Time
}

func NewService(r Repository, rooms RoomEmbedding, a Authorizer, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, rooms, a, k, ids, now}
}
func (s Service) Create(ctx context.Context, c Command, secondPlayerID string, window time.Duration) (session.Projection, error) {
	if !s.ready() || s.a.RequireParticipant(ctx, c.RoomID, c.ActorID) != nil || s.rooms.Revalidate(ctx, c.RoomID, c.ActorID, secondPlayerID) != nil {
		return session.Projection{}, ErrNotAvailable
	}
	room, e := s.k.Key("oware-session:room", strings.TrimSpace(c.RoomID))
	if e != nil {
		return session.Projection{}, ErrNotAvailable
	}
	first, e := s.k.Key("oware-session:player", strings.TrimSpace(c.ActorID))
	if e != nil {
		return session.Projection{}, ErrNotAvailable
	}
	second, e := s.k.Key("oware-session:player", strings.TrimSpace(secondPlayerID))
	if e != nil {
		return session.Projection{}, ErrNotAvailable
	}
	now := s.now().UTC()
	game, e := session.Create(s.ids.NewID(), room, []string{first, second}, window, now, s.command(c))
	if e != nil || s.r.Create(ctx, game) != nil {
		return session.Projection{}, ErrNotAvailable
	}
	return game.Project(now), nil
}
func (s Service) Move(ctx context.Context, c Command, pit int) (session.Projection, error) {
	game, actor, e := s.current(ctx, c)
	if e != nil {
		return session.Projection{}, e
	}
	next, e := game.Move(actor, pit, s.now().UTC(), s.command(c))
	if e != nil {
		return session.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, game, next, c.ID)
}
func (s Service) Expire(ctx context.Context, c Command) (session.Projection, error) {
	game, _, e := s.current(ctx, c)
	if e != nil {
		return session.Projection{}, e
	}
	next, e := game.Expire(s.now().UTC(), s.command(c))
	if e != nil {
		return session.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, game, next, c.ID)
}
func (s Service) View(ctx context.Context, c Command) (session.Projection, error) {
	game, _, e := s.current(ctx, c)
	if e != nil {
		return session.Projection{}, e
	}
	return game.Project(s.now().UTC()), nil
}
func (s Service) current(ctx context.Context, c Command) (session.Session, string, error) {
	if !s.ready() || s.a.RequireParticipant(ctx, c.RoomID, c.ActorID) != nil {
		return session.Session{}, "", ErrNotAvailable
	}
	game, e := s.r.Find(ctx, strings.TrimSpace(c.SessionID))
	if e != nil {
		return session.Session{}, "", ErrNotAvailable
	}
	room, e := s.k.Key("oware-session:room", strings.TrimSpace(c.RoomID))
	if e != nil || room != game.RoomRef() {
		return session.Session{}, "", ErrNotAvailable
	}
	actor, e := s.k.Key("oware-session:player", strings.TrimSpace(c.ActorID))
	if e != nil {
		return session.Session{}, "", ErrNotAvailable
	}
	found := false
	for _, p := range game.Players() {
		if p == actor {
			found = true
		}
	}
	if !found {
		return session.Session{}, "", ErrNotAvailable
	}
	return game, actor, nil
}
func (s Service) append(ctx context.Context, current, next session.Session, id string) (session.Projection, error) {
	e := s.r.Append(ctx, next, current.Revision(), id)
	if e == nil {
		return next.Project(s.now().UTC()), nil
	}
	if errors.Is(e, ErrApplied) {
		old, x := s.r.FindByCommand(ctx, id)
		if x == nil {
			return old.Project(s.now().UTC()), nil
		}
	}
	return session.Projection{}, ErrNotAvailable
}
func (s Service) command(c Command) session.Command {
	return session.Command{ID: strings.TrimSpace(c.ID), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) ready() bool {
	return s.r != nil && s.rooms != nil && s.a != nil && s.k != nil && s.ids != nil
}
