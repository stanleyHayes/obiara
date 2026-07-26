package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/domain"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("conduct signal not found")
	ErrConflict     = errors.New("conduct signal conflict")
	ErrApplied      = errors.New("conduct command applied")
	ErrNotAvailable = errors.New("conduct signal not available")
)

type Command struct {
	ID, SignalID, GameID, ActorID, SubjectID, EventRef string
	ExpectedRevision                                   uint64
}
type Service struct {
	r      Repository
	a      Authority
	events EventVerifier
	k      Keyer
	ids    IDSource
	now    func() time.Time
}

func NewService(r Repository, a Authority, e EventVerifier, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, e, k, ids, now}
}
func (s Service) Record(ctx context.Context, c Command, event domain.GameEvent) (domain.Projection, error) {
	if !s.ready() || s.a.RequireSubject(ctx, c.GameID, c.ActorID, c.SubjectID) != nil || s.events.Revalidate(ctx, c.GameID, c.EventRef, c.SubjectID, event) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	game, e := s.k.Key("game-conduct:game", strings.TrimSpace(c.GameID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	subject, e := s.k.Key("game-conduct:subject", strings.TrimSpace(c.SubjectID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	eventKey, e := s.k.Key("game-conduct:event", strings.TrimSpace(c.EventRef))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	now := s.now().UTC()
	signal, e := domain.Record(s.ids.NewID(), game, subject, eventKey, event, now, s.command(c))
	if e != nil || s.r.Create(ctx, signal) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return signal.Project(), nil
}
func (s Service) Appeal(ctx context.Context, c Command) (domain.Projection, error) {
	signal, e := s.subject(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	next, e := signal.Appeal(s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, signal, next, c.ID)
}
func (s Service) Resolve(ctx context.Context, c Command, result domain.AppealState) (domain.Projection, error) {
	if !s.ready() || s.a.RequireAppealReviewer(ctx, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	signal, e := s.r.Find(ctx, strings.TrimSpace(c.SignalID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := signal.Resolve(result, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, signal, next, c.ID)
}
func (s Service) View(ctx context.Context, c Command) (domain.Projection, error) {
	signal, e := s.subject(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	return signal.Project(), nil
}
func (s Service) subject(ctx context.Context, c Command) (domain.Signal, error) {
	if !s.ready() || s.a.RequireSubject(ctx, c.GameID, c.ActorID, c.SubjectID) != nil {
		return domain.Signal{}, ErrNotAvailable
	}
	signal, e := s.r.Find(ctx, strings.TrimSpace(c.SignalID))
	if e != nil {
		return domain.Signal{}, ErrNotAvailable
	}
	game, e := s.k.Key("game-conduct:game", strings.TrimSpace(c.GameID))
	if e != nil || game != signal.GameKey() {
		return domain.Signal{}, ErrNotAvailable
	}
	subject, e := s.k.Key("game-conduct:subject", strings.TrimSpace(c.SubjectID))
	if e != nil || subject != signal.SubjectKey() {
		return domain.Signal{}, ErrNotAvailable
	}
	return signal, nil
}
func (s Service) append(ctx context.Context, current, next domain.Signal, id string) (domain.Projection, error) {
	e := s.r.Append(ctx, next, current.Revision(), id)
	if e == nil {
		return next.Project(), nil
	}
	if errors.Is(e, ErrApplied) {
		old, x := s.r.FindByCommand(ctx, id)
		if x == nil {
			return old.Project(), nil
		}
	}
	return domain.Projection{}, ErrNotAvailable
}
func (s Service) command(c Command) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.events != nil && s.k != nil && s.ids != nil
}
