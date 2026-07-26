package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	"strings"
	"time"
)

type CreateCommand struct {
	ID, SenderID, RecipientID, DetailRef string
	Kind                                 domain.Type
	ExpiresAt                            time.Time
}
type DecisionCommand struct {
	ID, ProposalID, ActorID string
	ExpectedRevision        uint64
}
type Result struct {
	Proposal domain.Proposal
	Replayed bool
}
type Service struct {
	repository Repository
	keyer      Keyer
	protector  DetailProtector
	ids        IDSource
	now        func() time.Time
}

func New(r Repository, k Keyer, p DetailProtector, ids IDSource, now func() time.Time) Service {
	return Service{r, k, p, ids, now}
}
func (s Service) Create(ctx context.Context, c CreateCommand) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	sender, err := s.key("member", c.SenderID)
	if err != nil {
		return Result{}, err
	}
	recipient, err := s.key("member", c.RecipientID)
	if err != nil {
		return Result{}, err
	}
	detail, err := s.protector.Protect(ctx, strings.TrimSpace(c.DetailRef))
	if err != nil {
		return Result{}, ErrUnavailable
	}
	id := s.ids.NewID()
	fp := createFingerprint(strings.TrimSpace(c.ID), c.Kind, sender, recipient, detail, c.ExpiresAt)
	proposal, err := domain.Create(id, c.Kind, sender, recipient, detail, c.ExpiresAt, domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: sender, Fingerprint: fp, ExpectedRevision: 0, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.repository.Create(ctx, proposal); err == nil {
		return Result{Proposal: proposal}, nil
	} else if !errors.Is(err, ErrCommandApplied) {
		return Result{}, s.neutral(err)
	}
	existing, findErr := s.repository.FindByCommand(ctx, c.ID)
	if findErr != nil {
		return Result{}, ErrNotAvailable
	}
	events := existing.State().Events
	for _, event := range events {
		if event.CommandID == c.ID {
			if event.Fingerprint != fp {
				return Result{}, domain.ErrCommandMismatch
			}
			return Result{existing, true}, nil
		}
	}
	return Result{}, domain.ErrCommandMismatch
}
func (s Service) Accept(ctx context.Context, c DecisionCommand) (Result, error) {
	return s.decide(ctx, c, domain.ActionAccepted)
}
func (s Service) Reject(ctx context.Context, c DecisionCommand) (Result, error) {
	return s.decide(ctx, c, domain.ActionRejected)
}
func (s Service) Withdraw(ctx context.Context, c DecisionCommand) (Result, error) {
	return s.decide(ctx, c, domain.ActionWithdrawn)
}
func (s Service) decide(ctx context.Context, c DecisionCommand, action domain.Action) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.ProposalID))
	if err != nil {
		return Result{}, ErrNotAvailable
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	fp := domain.Fingerprint(current.ID(), action, actor, c.ExpectedRevision)
	command := domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()}
	var next domain.Proposal
	switch action {
	case domain.ActionAccepted:
		next, err = current.Accept(command)
	case domain.ActionRejected:
		next, err = current.Reject(command)
	default:
		next, err = current.Withdraw(command)
	}
	if err != nil {
		if errors.Is(err, domain.ErrNotParticipant) {
			return Result{}, ErrNotAvailable
		}
		return Result{}, err
	}
	if next.Revision() == current.Revision() {
		return Result{next, true}, nil
	}
	if err = s.repository.Append(ctx, next, current.Revision(), c.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			replayed, findErr := s.repository.FindByCommand(ctx, c.ID)
			if findErr == nil {
				return Result{replayed, true}, nil
			}
		}
		return Result{}, s.neutral(err)
	}
	return Result{next, false}, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.protector != nil && s.ids != nil && s.now != nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (s Service) neutral(err error) error {
	if errors.Is(err, ErrConcurrentChange) {
		return err
	}
	return ErrNotAvailable
}
func createFingerprint(commandID string, kind domain.Type, sender, recipient, detail string, expiresAt time.Time) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q|%q|%q|%q|%q|%s", commandID, kind, sender, recipient, detail, expiresAt.UTC().Format(time.RFC3339Nano))))
	return hex.EncodeToString(sum[:])
}
