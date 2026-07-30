package application

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

var (
	ErrNotFound     = errors.New("ampe round not found")
	ErrConflict     = errors.New("ampe round conflict")
	ErrApplied      = errors.New("ampe command applied")
	ErrNotAvailable = errors.New("ampe round not available")
)

type Command struct {
	ID, RoundID, RoomID, ActorID, FirstPlayerID, SecondPlayerID string
	ExpectedSequence                                            uint64
}

type PlayerProjection struct {
	Ready, Connected, Locked bool
}

type Projection struct {
	ID                      string
	Sequence                uint64
	You, Other              PlayerProjection
	Paused                  bool
	OwnChoice               *domain.Choice
	YourReveal, OtherReveal *domain.Choice
	Complete                bool
}

type Service struct {
	r   Repository
	a   Authority
	k   Keyer
	ids IDSource
}

func NewService(r Repository, a Authority, k Keyer, ids IDSource) Service {
	return Service{r: r, a: a, k: k, ids: ids}
}

func (s Service) Create(ctx context.Context, c Command) (Projection, error) {
	if !s.ready() || c.ActorID != c.FirstPlayerID ||
		s.a.Revalidate(ctx, c.RoomID, c.FirstPlayerID, c.SecondPlayerID) != nil {
		return Projection{}, ErrNotAvailable
	}
	room, players, err := s.keys(c.RoomID, c.FirstPlayerID, c.SecondPlayerID)
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	round, err := domain.Open(domain.Spec{
		ID: s.ids.NewID(), RoomKey: room,
		PlayerKeys: [2]string{players[0], players[1]},
	})
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	if err = s.r.Create(ctx, round, strings.TrimSpace(c.ID)); err != nil {
		if errors.Is(err, ErrApplied) {
			if prior, findErr := s.r.FindByCommand(ctx, strings.TrimSpace(c.ID)); findErr == nil {
				actor, keyErr := s.k.Key("ampe:player", strings.TrimSpace(c.ActorID))
				if keyErr == nil {
					return project(prior, actor)
				}
			}
		}
		return Projection{}, ErrNotAvailable
	}
	actor, err := s.k.Key("ampe:player", strings.TrimSpace(c.ActorID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	return project(round, actor)
}

func (s Service) Apply(ctx context.Context, c Command, action domain.Action, choice domain.Choice) (Projection, error) {
	round, actor, err := s.current(ctx, c)
	if err != nil {
		return Projection{}, err
	}
	next, replayed, err := round.Apply(domain.Command{
		ID: strings.TrimSpace(c.ID), ActorKey: actor, Action: action,
		Choice: choice, ExpectedSequence: c.ExpectedSequence,
	})
	if replayed {
		return project(round, actor)
	}
	if err != nil {
		if errors.Is(err, domain.ErrStaleSequence) {
			return Projection{}, ErrConflict
		}
		return Projection{}, ErrNotAvailable
	}
	if err = s.r.Append(ctx, next, round.Sequence(), strings.TrimSpace(c.ID)); err != nil {
		if errors.Is(err, ErrApplied) {
			if prior, findErr := s.r.FindByCommand(ctx, strings.TrimSpace(c.ID)); findErr == nil {
				return project(prior, actor)
			}
		}
		return Projection{}, ErrConflict
	}
	return project(next, actor)
}

func (s Service) View(ctx context.Context, c Command) (Projection, error) {
	round, actor, err := s.current(ctx, c)
	if err != nil {
		return Projection{}, err
	}
	return project(round, actor)
}

func (s Service) current(ctx context.Context, c Command) (domain.Round, string, error) {
	if !s.ready() || s.a.Revalidate(ctx, c.RoomID, c.FirstPlayerID, c.SecondPlayerID) != nil {
		return domain.Round{}, "", ErrNotAvailable
	}
	round, err := s.r.Find(ctx, strings.TrimSpace(c.RoundID))
	if err != nil {
		return domain.Round{}, "", ErrNotAvailable
	}
	room, players, err := s.keys(c.RoomID, c.FirstPlayerID, c.SecondPlayerID)
	spec := round.Specification()
	if err != nil || room != spec.RoomKey || spec.PlayerKeys != [2]string{players[0], players[1]} {
		return domain.Round{}, "", ErrNotAvailable
	}
	actor, err := s.k.Key("ampe:player", strings.TrimSpace(c.ActorID))
	if err != nil || !slices.Contains(players, actor) {
		return domain.Round{}, "", ErrNotAvailable
	}
	return round, actor, nil
}

func (s Service) keys(roomID, firstID, secondID string) (string, []string, error) {
	room, err := s.k.Key("ampe:room", strings.TrimSpace(roomID))
	if err != nil {
		return "", nil, err
	}
	first, err := s.k.Key("ampe:player", strings.TrimSpace(firstID))
	if err != nil {
		return "", nil, err
	}
	second, err := s.k.Key("ampe:player", strings.TrimSpace(secondID))
	if err != nil {
		return "", nil, err
	}
	players := []string{first, second}
	slices.Sort(players)
	return room, players, nil
}

func project(round domain.Round, actor string) (Projection, error) {
	view, err := round.View(actor)
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	index := 0
	if view.Players[1].Key == actor {
		index = 1
	}
	result := Projection{
		ID: view.ID, Sequence: view.Sequence, Paused: view.Paused,
		You:       PlayerProjection{view.Players[index].Ready, view.Players[index].Connected, view.Players[index].Locked},
		Other:     PlayerProjection{view.Players[1-index].Ready, view.Players[1-index].Connected, view.Players[1-index].Locked},
		OwnChoice: view.OwnChoice, Complete: view.Reveal != nil,
	}
	if view.Reveal != nil {
		yours, other := view.Reveal.Choices[index], view.Reveal.Choices[1-index]
		result.YourReveal, result.OtherReveal = &yours, &other
	}
	return result, nil
}

func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.k != nil && s.ids != nil
}
