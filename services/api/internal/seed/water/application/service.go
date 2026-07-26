package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
	"strings"
	"time"
)

var (
	ErrNotFound           = errors.New("mutual water not found")
	ErrOptimisticConflict = errors.New("mutual water optimistic conflict")
	ErrCommandApplied     = errors.New("mutual water command applied")
	ErrUnavailable        = errors.New("mutual water unavailable")
	ErrNotAvailable       = errors.New("mutual water not available")
)

type Command struct {
	ID, WaterID, ActorID, ReasonCode string
	ExpectedRevision                 uint64
}
type StartProposal struct{ FirstMemberID, SecondMemberID string }
type Result struct {
	Water    domain.Water
	Replayed bool
}
type Service struct {
	r                 Repository
	a                 Authorizer
	c                 PairConsent
	k                 Keyer
	waterIDs, roomIDs IDSource
	now               func() time.Time
}

func NewService(r Repository, a Authorizer, c PairConsent, k Keyer, wids, rids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, c, k, wids, rids, now}
}
func (s Service) Start(ctx context.Context, c Command, p StartProposal) (Result, error) {
	if !s.ready() || c.ActorID != p.FirstMemberID {
		return Result{}, ErrNotAvailable
	}
	if e := s.a.Require(ctx, c.ActorID, "seed.water.start", ""); e != nil {
		return Result{}, ErrNotAvailable
	}
	if e := s.c.Revalidate(ctx, p.FirstMemberID, p.SecondMemberID); e != nil {
		return Result{}, ErrNotAvailable
	}
	first, e := s.key(p.FirstMemberID)
	if e != nil {
		return Result{}, e
	}
	second, e := s.key(p.SecondMemberID)
	if e != nil {
		return Result{}, e
	}
	w, e := domain.Start(s.waterIDs.NewID(), []string{first, second}, s.command(c, first))
	if e != nil {
		return Result{}, e
	}
	if e = s.r.Create(ctx, w); e != nil {
		if !errors.Is(e, ErrCommandApplied) {
			return Result{}, s.translate(e)
		}
		old, x := s.r.FindByCommand(ctx, c.ID)
		if x != nil {
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{Water: old, Replayed: true}, nil
	}
	return Result{Water: w}, nil
}
func (s Service) Water(ctx context.Context, c Command) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, e := s.r.Find(ctx, strings.TrimSpace(c.WaterID))
	if e != nil {
		return Result{}, s.translate(e)
	}
	if e = s.a.Require(ctx, c.ActorID, "seed.water.mutual", current.ID()); e != nil {
		return Result{}, ErrNotAvailable
	}
	actor, e := s.key(c.ActorID)
	if e != nil || !current.IsMember(actor) {
		return Result{}, ErrNotAvailable
	}
	members := current.Members()
	if e = s.c.Revalidate(ctx, members[0], members[1]); e != nil {
		return Result{}, ErrNotAvailable
	}
	replay := current.HasCommand(c.ID)
	room := ""
	if len(current.Watered()) == 1 {
		room, e = s.k.Key("seed-water:room", s.roomIDs.NewID())
		if e != nil {
			return Result{}, ErrUnavailable
		}
	}
	next, e := current.Water(s.command(c, actor), room)
	if e != nil {
		return Result{}, e
	}
	if replay {
		return Result{Water: next, Replayed: true}, nil
	}
	if e = s.r.Append(ctx, next, current.Revision(), c.ID); e != nil {
		if errors.Is(e, ErrCommandApplied) {
			old, x := s.r.FindByCommand(ctx, c.ID)
			if x == nil {
				return Result{Water: old, Replayed: true}, nil
			}
		}
		return Result{}, s.translate(e)
	}
	return Result{Water: next}, nil
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.c != nil && s.k != nil && s.waterIDs != nil && s.roomIDs != nil
}
func (s Service) key(v string) (string, error) {
	x, e := s.k.Key("seed-water:member", strings.TrimSpace(v))
	if e != nil {
		return "", ErrUnavailable
	}
	return x, nil
}
func (s Service) command(c Command, a string) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: a, ReasonCode: strings.TrimSpace(c.ReasonCode), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) translate(e error) error {
	if errors.Is(e, ErrNotFound) || errors.Is(e, ErrCommandApplied) {
		return e
	}
	if errors.Is(e, ErrOptimisticConflict) {
		return domain.ErrStaleRevision
	}
	return ErrUnavailable
}
