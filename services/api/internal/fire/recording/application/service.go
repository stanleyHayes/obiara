package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/domain"
	"slices"
	"strings"
	"time"
)

type OpenCommand struct {
	ID, FireRef, HostID string
	ParticipantIDs      []string
}
type Command struct {
	ID, PolicyID, ActorID, SubjectID string
	Purpose                          domain.Purpose
	Retention                        time.Duration
	ExpectedRevision                 uint64
}
type Result struct {
	Policy   domain.Policy
	Replayed bool
}
type Service struct {
	repository Repository
	keyer      Keyer
	ids        IDSource
	authority  Authority
	membership Membership
	recorder   Recorder
	now        func() time.Time
}

func New(r Repository, k Keyer, ids IDSource, a Authority, m Membership, rec Recorder, now func() time.Time) Service {
	return Service{r, k, ids, a, m, rec, now}
}
func (s Service) Open(ctx context.Context, c OpenCommand) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	fire, err := s.key("fire", c.FireRef)
	if err != nil {
		return Result{}, err
	}
	host, err := s.key("participant", c.HostID)
	if err != nil {
		return Result{}, err
	}
	participants := make([]string, 0, len(c.ParticipantIDs))
	for _, id := range c.ParticipantIDs {
		k, e := s.key("participant", id)
		if e != nil {
			return Result{}, e
		}
		participants = append(participants, k)
	}
	id := s.ids.NewID()
	fp := domain.Fingerprint(c.ID+"|"+fire, domain.ActionOpened, host, host, "", 0, 0)
	p, err := domain.Open(id, fire, host, participants, domain.Command{ID: c.ID, ActorKey: host, Fingerprint: fp, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.repository.Create(ctx, p); err != nil {
		return Result{}, err
	}
	return Result{p, false}, nil
}
func (s Service) Propose(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionProposed)
}
func (s Service) OptIn(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionOptedIn)
}
func (s Service) Revoke(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionRevoked)
}
func (s Service) Join(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionJoined)
}
func (s Service) Leave(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionLeft)
}
func (s Service) Start(ctx context.Context, c Command) (Result, error) {
	return s.change(ctx, c, domain.ActionStarted)
}
func (s Service) change(ctx context.Context, c Command, action domain.Action) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.PolicyID))
	if err != nil {
		return Result{}, ErrUnavailable
	}
	actor, err := s.key("participant", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	subject := actor
	if c.SubjectID != "" {
		subject, err = s.key("participant", c.SubjectID)
		if err != nil {
			return Result{}, err
		}
	}
	state := current.State()
	fp := domain.Fingerprint(current.ID(), action, actor, subject, c.Purpose, c.Retention, c.ExpectedRevision)
	command := domain.Command{ID: c.ID, ActorKey: actor, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()}
	var next domain.Policy
	switch action {
	case domain.ActionProposed:
		next, err = current.Propose(c.Purpose, c.Retention, command)
	case domain.ActionOptedIn:
		next, err = current.OptIn(command)
	case domain.ActionRevoked:
		next, err = current.Revoke(command)
	case domain.ActionJoined:
		next, err = current.Join(subject, command)
	case domain.ActionLeft:
		next, err = current.Leave(subject, command)
	case domain.ActionStarted:
		if state.Proposal == nil {
			return Result{}, domain.ErrConsentRequired
		}
		if _, validationErr := current.Start(strings.Repeat("0", 64), command); validationErr != nil {
			return Result{}, validationErr
		}
		ref, startErr := s.prepareStart(ctx, current, actor, c.ID)
		if startErr != nil {
			return Result{}, startErr
		}
		next, err = current.Start(ref, command)
	}
	if err != nil {
		return Result{}, err
	}
	if next.Revision() == current.Revision() {
		return Result{next, true}, nil
	}
	if action != domain.ActionStarted {
		if err = s.authority.Authorize(ctx, state, action, actor); err != nil {
			return Result{}, ErrUnavailable
		}
		if err = s.validateMembership(ctx, next.State()); err != nil {
			return Result{}, err
		}
		if state.Active && (action == domain.ActionRevoked || action == domain.ActionJoined || action == domain.ActionLeft || action == domain.ActionProposed) {
			if err = s.recorder.Stop(ctx, state.FireKey, c.ID); err != nil {
				return Result{}, ErrUnavailable
			}
		}
	}
	if err = s.repository.Append(ctx, next, current.Revision(), c.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			existing, e := s.repository.FindByCommand(ctx, c.ID)
			if e == nil {
				return Result{existing, true}, nil
			}
		}
		return Result{}, err
	}
	return Result{next, false}, nil
}
func (s Service) prepareStart(ctx context.Context, current domain.Policy, actor, commandID string) (string, error) {
	state := current.State()
	if err := s.authority.Authorize(ctx, state, domain.ActionStarted, actor); err != nil {
		return "", ErrUnavailable
	}
	if err := s.validateMembership(ctx, state); err != nil {
		return "", err
	}
	ref, err := s.recorder.Start(ctx, state.FireKey, state.Proposal.Purpose, state.Proposal.Retention, commandID)
	if err != nil {
		return "", ErrUnavailable
	}
	return ref, nil
}
func (s Service) validateMembership(ctx context.Context, state domain.State) error {
	current, err := s.membership.Current(ctx, state.FireKey)
	if err != nil {
		return ErrUnavailable
	}
	slices.Sort(current)
	if !slices.Equal(current, state.Participants) {
		return ErrUnavailable
	}
	return nil
}
func (s Service) key(namespace, value string) (string, error) {
	k, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return k, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.ids != nil && s.authority != nil && s.membership != nil && s.recorder != nil && s.now != nil
}
