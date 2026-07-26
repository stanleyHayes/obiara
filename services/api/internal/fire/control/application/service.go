package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/domain"
	"strings"
	"time"
)

type OpenCommand struct {
	ID, FireRef, HostID string
	ParticipantIDs      []string
}
type Command struct {
	ID, ControlID, ActorID, TargetID, ReasonCode string
	ExpectedRevision                             uint64
}
type Result struct {
	Fire     domain.Fire
	Replayed bool
}
type Service struct {
	repository  Repository
	keyer       Keyer
	ids         IDSource
	revalidator Revalidator
	realtime    RealtimeControl
	now         func() time.Time
}

func New(r Repository, k Keyer, ids IDSource, v Revalidator, rt RealtimeControl, now func() time.Time) Service {
	return Service{r, k, ids, v, rt, now}
}
func (s Service) Open(ctx context.Context, c OpenCommand) (Result, error) {
	if !s.ready() {
		return Result{}, ErrNotAvailable
	}
	fireKey, err := s.key("fire", c.FireRef)
	if err != nil {
		return Result{}, err
	}
	host, err := s.key("participant", c.HostID)
	if err != nil {
		return Result{}, err
	}
	participants := make([]string, 0, len(c.ParticipantIDs))
	for _, id := range c.ParticipantIDs {
		key, keyErr := s.key("participant", id)
		if keyErr != nil {
			return Result{}, keyErr
		}
		participants = append(participants, key)
	}
	id := s.ids.NewID()
	reason := "fire.opened"
	fp := domain.Fingerprint(strings.TrimSpace(c.ID)+"|"+fireKey, domain.ActionOpened, host, host, reason, 0)
	fire, err := domain.Open(id, fireKey, host, participants, domain.Command{ID: c.ID, ActorKey: host, ReasonCode: reason, Fingerprint: fp, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.repository.Create(ctx, fire); err != nil {
		return Result{}, err
	}
	return Result{fire, false}, nil
}
func (s Service) Promote(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionPromoted)
}
func (s Service) Demote(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionDemoted)
}
func (s Service) Mute(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionMuted)
}
func (s Service) Eject(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionEjected)
}
func (s Service) change(ctx context.Context, c Command, action domain.Action) (Result, error) {
	if !s.ready() {
		return Result{}, ErrNotAvailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.ControlID))
	if err != nil {
		return Result{}, ErrNotAvailable
	}
	actor, err := s.key("participant", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	target, err := s.key("participant", c.TargetID)
	if err != nil {
		return Result{}, err
	}
	reason := strings.TrimSpace(c.ReasonCode)
	fp := domain.Fingerprint(current.ID(), action, actor, target, reason, c.ExpectedRevision)
	command := domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, ReasonCode: reason, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()}
	var next domain.Fire
	switch action {
	case domain.ActionPromoted:
		next, err = current.Promote(target, command)
	case domain.ActionDemoted:
		next, err = current.Demote(target, command)
	case domain.ActionMuted:
		next, err = current.Mute(target, command)
	default:
		next, err = current.Eject(target, command)
	}
	if err != nil {
		if errors.Is(err, domain.ErrDenied) {
			return Result{}, ErrNotAvailable
		}
		return Result{}, err
	}
	if next.Revision() == current.Revision() {
		return Result{next, true}, nil
	}
	if err = s.revalidator.Authorize(ctx, current.State(), action, actor, target); err != nil {
		return Result{}, ErrNotAvailable
	}
	fireKey := current.State().FireKey
	switch action {
	case domain.ActionPromoted:
		err = s.realtime.SetRole(ctx, fireKey, target, domain.RoleCohost, c.ID)
	case domain.ActionDemoted:
		err = s.realtime.SetRole(ctx, fireKey, target, domain.RoleParticipant, c.ID)
	case domain.ActionMuted:
		err = s.realtime.Mute(ctx, fireKey, target, c.ID)
	case domain.ActionEjected:
		err = s.realtime.EjectAndRevoke(ctx, fireKey, target, c.ID)
	}
	if err != nil {
		return Result{}, ErrNotAvailable
	}
	if err = s.repository.Append(ctx, next, current.Revision(), c.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			existing, findErr := s.repository.FindByCommand(ctx, c.ID)
			if findErr == nil {
				return Result{existing, true}, nil
			}
		}
		return Result{}, err
	}
	return Result{next, false}, nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrNotAvailable
	}
	return key, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.ids != nil && s.revalidator != nil && s.realtime != nil && s.now != nil
}
