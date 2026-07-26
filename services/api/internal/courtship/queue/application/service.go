package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
)

type SubmitCommand struct {
	ID, RoomRef, DeviceRef, ActorID, PayloadRef string
	BaseSequence                                uint64
}
type Result struct {
	Event    domain.Event
	Replayed bool
}
type Service struct {
	repository Repository
	keyer      Keyer
	now        func() time.Time
}

func New(repository Repository, keyer Keyer, now func() time.Time) Service {
	return Service{repository, keyer, now}
}
func (s Service) Submit(ctx context.Context, c SubmitCommand) (Result, error) {
	if s.repository == nil || s.keyer == nil || s.now == nil {
		return Result{}, ErrUnavailable
	}
	room, err := s.key("room", c.RoomRef)
	if err != nil {
		return Result{}, err
	}
	device, err := s.key("device", c.DeviceRef)
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("actor", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	payload, err := s.key("payload", c.PayloadRef)
	if err != nil {
		return Result{}, err
	}
	state, err := s.repository.State(ctx, room)
	if err != nil {
		return Result{}, err
	}
	fp := fingerprint(c.ID, room, device, actor, payload, c.BaseSequence)
	next, event, err := state.Accept(domain.Command{ID: strings.TrimSpace(c.ID), DeviceKey: device, ActorKey: actor, PayloadKey: payload, Fingerprint: fp, BaseSequence: c.BaseSequence, At: s.now()})
	if err != nil {
		if err == domain.ErrStaleDevice {
			existing, findErr := s.repository.EventByCommand(ctx, room, strings.TrimSpace(c.ID))
			if findErr == nil {
				if existing.Fingerprint != fp {
					return Result{}, domain.ErrCommandMismatch
				}
				return Result{existing, true}, nil
			}
		}
		return Result{}, err
	}
	stored, replayed, err := s.repository.Append(ctx, next, event, state.Revision)
	if err != nil {
		return Result{}, err
	}
	return Result{stored, replayed}, nil
}
func (s Service) Events(ctx context.Context, roomRef string, after uint64, limit int) ([]domain.Event, error) {
	room, err := s.key("room", roomRef)
	if err != nil {
		return nil, err
	}
	return s.repository.Events(ctx, room, after, limit)
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func fingerprint(id, room, device, actor, payload string, base uint64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q|%q|%q|%q|%q|%d", id, room, device, actor, payload, base)))
	return hex.EncodeToString(sum[:])
}
