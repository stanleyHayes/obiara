package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/domain"
	"strings"
	"time"
)

var (
	ErrNotFound           = errors.New("courtship room not found")
	ErrOptimisticConflict = errors.New("courtship room optimistic conflict")
	ErrCommandApplied     = errors.New("courtship room command applied")
	ErrUnavailable        = errors.New("courtship room unavailable")
	ErrNotAvailable       = errors.New("courtship room not available")
)

type Command struct {
	ID, RoomID, ActorID, ReasonCode string
	ExpectedRevision                uint64
}
type Proposal struct{ FirstMemberID, SecondMemberID string }
type Result struct {
	Room     domain.Room
	Replayed bool
}
type Service struct {
	r   Repository
	a   Authorizer
	m   Membership
	k   Keyer
	ids IDSource
	now func() time.Time
}

func NewService(r Repository, a Authorizer, m Membership, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, m, k, ids, now}
}
func (s Service) Open(ctx context.Context, c Command, p Proposal) (Result, error) {
	if !s.ready() || c.ActorID != p.FirstMemberID {
		return Result{}, ErrNotAvailable
	}
	if s.a.Require(ctx, c.ActorID, "courtship.room.open", "") != nil || s.m.RevalidatePair(ctx, p.FirstMemberID, p.SecondMemberID) != nil {
		return Result{}, ErrNotAvailable
	}
	a, e := s.member(p.FirstMemberID)
	if e != nil {
		return Result{}, e
	}
	b, e := s.member(p.SecondMemberID)
	if e != nil {
		return Result{}, e
	}
	r, e := domain.Open(s.ids.NewID(), []string{a, b}, s.command(c, a))
	if e != nil {
		return Result{}, e
	}
	if e = s.r.Create(ctx, r); e != nil {
		return Result{}, s.translate(e)
	}
	return Result{Room: r}, nil
}
func (s Service) Message(ctx context.Context, c Command, contentRef string) (Result, error) {
	content, e := s.k.Key("courtship-room:content", strings.TrimSpace(contentRef))
	if e != nil {
		return Result{}, ErrUnavailable
	}
	return s.mutate(ctx, c, "courtship.room.message", func(r domain.Room, d domain.Command) (domain.Room, error) { return r.Message(content, d) })
}
func (s Service) Close(ctx context.Context, c Command) (Result, error) {
	return s.mutate(ctx, c, "courtship.room.close", func(r domain.Room, d domain.Command) (domain.Room, error) { return r.Close(d) })
}
func (s Service) mutate(ctx context.Context, c Command, action string, f func(domain.Room, domain.Command) (domain.Room, error)) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	r, e := s.r.Find(ctx, strings.TrimSpace(c.RoomID))
	if e != nil {
		return Result{}, s.translate(e)
	}
	if s.a.Require(ctx, c.ActorID, action, r.ID()) != nil {
		return Result{}, ErrNotAvailable
	}
	actor, e := s.member(c.ActorID)
	if e != nil || !r.IsMember(actor) {
		return Result{}, ErrNotAvailable
	}
	members := r.Members()
	if s.m.RevalidatePair(ctx, members[0], members[1]) != nil {
		return Result{}, ErrNotAvailable
	}
	replay := r.HasCommand(c.ID)
	next, e := f(r, s.command(c, actor))
	if e != nil {
		return Result{}, e
	}
	if replay {
		return Result{Room: next, Replayed: true}, nil
	}
	if e = s.r.Append(ctx, next, r.Revision(), c.ID); e != nil {
		if errors.Is(e, ErrCommandApplied) {
			old, x := s.r.FindByCommand(ctx, c.ID)
			if x == nil {
				return Result{Room: old, Replayed: true}, nil
			}
		}
		return Result{}, s.translate(e)
	}
	return Result{Room: next}, nil
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.m != nil && s.k != nil && s.ids != nil
}
func (s Service) member(v string) (string, error) {
	x, e := s.k.Key("courtship-room:member", strings.TrimSpace(v))
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
