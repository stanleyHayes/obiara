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
	// listen is the FR-202 gate. A nil one makes the service unavailable
	// rather than permissive: a sow that skipped the listening is the exact
	// outcome the gate exists to prevent, so it must not be what happens
	// when the wiring is missing.
	listen ListenGate
	// allowance is the FR-201a spend. Nil refuses, for the same reason the
	// listen gate does: a sow that cost nothing is the outcome the rule
	// exists to prevent.
	allowance Allowance
}

// WithAllowance attaches the weekly seed allowance.
func (s Service) WithAllowance(allowance Allowance) Service {
	s.allowance = allowance
	return s
}

// WithListenGate attaches the listening requirement.
func (s Service) WithListenGate(listen ListenGate) Service {
	s.listen = listen
	return s
}

func New(repository Repository, keyer Keyer, ids IDSource, now func() time.Time) Service {
	return Service{repository: repository, keyer: keyer, ids: ids, now: now}
}

func (s Service) Sprout(ctx context.Context, command SproutCommand) (SproutResult, error) {
	if !s.ready() {
		return SproutResult{}, ErrUnavailable
	}
	// Only sowing needs the gate. Speaking inside a doorway both members
	// already opened does not, and requiring it there would shut existing
	// conversations whenever the gate was unavailable.
	if s.listen == nil {
		return SproutResult{}, ErrUnavailable
	}
	// Checked before anything is keyed or written: a sow that was never
	// armed should leave no trace, not a refused one.
	heard, err := s.listen.Heard(ctx, command.ActorID, command.TargetID)
	if err != nil {
		return SproutResult{}, ErrUnavailable
	}
	if !heard {
		return SproutResult{}, ErrNotHeard
	}
	// Charged after the gate, so a refused sow is never paid for, and before
	// the intent is recorded, so a sow that could not be paid for leaves no
	// trace. Both sides are idempotent by command id, so a client retry
	// completes the sow without charging twice.
	if s.allowance == nil {
		return SproutResult{}, ErrUnavailable
	}
	if err := s.allowance.Spend(ctx, command.ActorID, strings.TrimSpace(command.ID)); err != nil {
		return SproutResult{}, err
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
