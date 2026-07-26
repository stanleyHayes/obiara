package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
)

type SproutCommand struct{ ID, ActorID, TargetID, SeedRef string }
type ExchangeCommand struct{ ID, DoorwayID, ActorID, MessageRef string }
type SproutResult struct {
	Doorway  *domain.Doorway
	Replayed bool
}
type ExchangeResult struct {
	Doorway  domain.Doorway
	Replayed bool
}

type Service struct {
	repository Repository
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func New(repository Repository, keyer Keyer, ids IDSource, now func() time.Time) Service {
	return Service{repository, keyer, ids, now}
}

func (s Service) Sprout(ctx context.Context, command SproutCommand) (SproutResult, error) {
	if !s.ready() {
		return SproutResult{}, ErrUnavailable
	}
	actor, err := s.key("participant", command.ActorID)
	if err != nil {
		return SproutResult{}, err
	}
	target, err := s.key("participant", command.TargetID)
	if err != nil {
		return SproutResult{}, err
	}
	seed, err := s.key("seed", command.SeedRef)
	if err != nil {
		return SproutResult{}, err
	}
	fp := fingerprint("sprout", command.ID, actor, target, seed)
	intent, err := domain.NewIntent(s.ids.NewID(), actor, target, seed, strings.TrimSpace(command.ID), fp, s.now())
	if err != nil {
		return SproutResult{}, err
	}
	doorway, replayed, err := s.repository.RecordIntent(ctx, intent)
	if err != nil {
		return SproutResult{}, err
	}
	return SproutResult{doorway, replayed}, nil
}

func (s Service) Exchange(ctx context.Context, command ExchangeCommand) (ExchangeResult, error) {
	if !s.ready() {
		return ExchangeResult{}, ErrUnavailable
	}
	current, err := s.repository.FindDoorway(ctx, strings.TrimSpace(command.DoorwayID))
	if err != nil {
		return ExchangeResult{}, err
	}
	actor, err := s.key("participant", command.ActorID)
	if err != nil {
		return ExchangeResult{}, err
	}
	message, err := s.key("message", command.MessageRef)
	if err != nil {
		return ExchangeResult{}, err
	}
	fp := fingerprint("exchange", command.ID, current.ID(), actor, message)
	next, changed, err := current.Exchange(actor, message, strings.TrimSpace(command.ID), fp, s.now())
	if err != nil {
		return ExchangeResult{}, err
	}
	if !changed {
		return ExchangeResult{next, true}, nil
	}
	stored, replayed, err := s.repository.AppendExchange(ctx, next, current.Revision())
	if err != nil {
		return ExchangeResult{}, err
	}
	return ExchangeResult{stored, replayed}, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.ids != nil && s.now != nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func fingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q", parts)))
	return hex.EncodeToString(sum[:])
}
