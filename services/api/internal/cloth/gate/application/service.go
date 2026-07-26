package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/domain"
	"strings"
	"time"
)

type OpenCommand struct {
	ID, Version, ActorID string
	MemberIDs            [2]string
}
type ChangeCommand struct {
	ID, PolicyID, ActorID, ReviewerID, QuestionRef, MaterialRef string
	ExpectedRevision                                            uint64
}
type Result struct {
	Policy   domain.Policy
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
func (s Service) Open(ctx context.Context, c OpenCommand) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	members, err := s.members(c.MemberIDs)
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	id := s.ids.NewID()
	fp := domain.Fingerprint(strings.TrimSpace(c.ID)+"|"+strings.TrimSpace(c.Version)+"|"+members[0]+"|"+members[1], domain.ActionOpened, actor, domain.Capability{}, 0)
	policy, err := domain.Open(id, strings.TrimSpace(c.Version), members, domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, Fingerprint: fp, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.repository.Create(ctx, policy); err == nil {
		return Result{policy, false}, nil
	} else if !errors.Is(err, ErrCommandApplied) {
		return Result{}, err
	}
	existing, findErr := s.repository.FindByCommand(ctx, c.ID)
	if findErr != nil {
		return Result{}, err
	}
	events := existing.State().Events
	if len(events) == 0 || events[0].Fingerprint != fp {
		return Result{}, domain.ErrCommandMismatch
	}
	return Result{existing, true}, nil
}
func (s Service) Grant(ctx context.Context, c ChangeCommand) (Result, error) {
	return s.change(ctx, c, domain.ActionGranted)
}
func (s Service) Revoke(ctx context.Context, c ChangeCommand) (Result, error) {
	return s.change(ctx, c, domain.ActionRevoked)
}
func (s Service) change(ctx context.Context, c ChangeCommand, action domain.Action) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.PolicyID))
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	capability, err := s.capability(c)
	if err != nil {
		return Result{}, err
	}
	fp := domain.Fingerprint(current.ID(), action, actor, capability, c.ExpectedRevision)
	command := domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()}
	var next domain.Policy
	if action == domain.ActionGranted {
		next, err = current.Grant(capability, command)
	} else {
		next, err = current.Revoke(capability, command)
	}
	if err != nil {
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
		return Result{}, err
	}
	return Result{next, false}, nil
}
func (s Service) capability(c ChangeCommand) (domain.Capability, error) {
	reviewer, err := s.key("reviewer", c.ReviewerID)
	if err != nil {
		return domain.Capability{}, err
	}
	question, err := s.key("question", c.QuestionRef)
	if err != nil {
		return domain.Capability{}, err
	}
	material, err := s.key("material", c.MaterialRef)
	if err != nil {
		return domain.Capability{}, err
	}
	return domain.Capability{ReviewerKey: reviewer, QuestionKey: question, MaterialKey: material}, nil
}
func (s Service) members(ids [2]string) ([2]string, error) {
	var out [2]string
	for i, id := range ids {
		key, err := s.key("member", id)
		if err != nil {
			return out, err
		}
		out[i] = key
	}
	return out, nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.ids != nil && s.now != nil
}
