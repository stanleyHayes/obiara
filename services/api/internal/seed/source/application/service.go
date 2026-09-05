package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

var (
	ErrNotFound           = errors.New("introduction source request not found")
	ErrOptimisticConflict = errors.New("introduction source optimistic conflict")
	ErrCommandApplied     = errors.New("introduction source command applied")
	ErrUnavailable        = errors.New("introduction source unavailable")
	ErrNotAvailable       = errors.New("introduction source not available")
)

type Command struct {
	ID, RequestID, ActorID, ReasonCode string
	ExpectedRevision                   uint64
}
type Proposal struct {
	RequesterID, SourceRef string
	SourceType             domain.SourceType
	TTL                    time.Duration
}
type Result struct {
	Request  domain.Request
	Replayed bool
}
type Service struct {
	repository Repository
	authorizer Authorizer
	policy     SourcePolicy
	resolver   CandidateResolver
	visibility ConsentVisibility
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(r Repository, a Authorizer, p SourcePolicy, c CandidateResolver, v ConsentVisibility, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, p, c, v, k, ids, now}
}

func (s Service) Open(ctx context.Context, command Command, proposal Proposal) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	requester, source := strings.TrimSpace(proposal.RequesterID), strings.TrimSpace(proposal.SourceRef)
	if command.ActorID != requester {
		return Result{}, ErrNotAvailable
	}
	if err := s.authorizer.Require(ctx, requester, "seed.source.open", proposal.SourceType, source); err != nil {
		return Result{}, ErrNotAvailable
	}
	if err := s.policy.Allow(ctx, proposal.SourceType, source); err != nil {
		return Result{}, ErrNotAvailable
	}
	raw, err := s.resolver.CandidateIDs(ctx, proposal.SourceType, source, domain.MaxCandidates)
	if err != nil || len(raw) > domain.MaxCandidates {
		return Result{}, ErrNotAvailable
	}
	candidates := make([]string, 0, len(raw))
	for _, candidate := range raw {
		visible, visibilityErr := s.visibility.Visible(ctx, requester, candidate, proposal.SourceType, source)
		if visibilityErr != nil {
			return Result{}, ErrNotAvailable
		}
		if !visible {
			continue
		}
		key, keyErr := s.key("seed-source:candidate", candidate)
		if keyErr != nil {
			return Result{}, keyErr
		}
		candidates = append(candidates, key)
	}
	requesterKey, err := s.key("seed-source:requester", requester)
	if err != nil {
		return Result{}, err
	}
	sourceKey, err := s.key("seed-source:source", source)
	if err != nil {
		return Result{}, err
	}
	opened, err := domain.Open(s.ids.NewID(), requesterKey, domain.Source{Type: proposal.SourceType, Key: sourceKey}, candidates, s.now().UTC().Add(proposal.TTL), s.domainCommand(command, requesterKey))
	if err != nil {
		return Result{}, err
	}
	if err := s.repository.Create(ctx, opened); err != nil {
		if !errors.Is(err, ErrCommandApplied) {
			return Result{}, s.translate(err)
		}
		replayed, findErr := s.repository.FindByCommand(ctx, command.ID)
		if findErr != nil || !replayed.HasCommand(command.ID) {
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{Request: replayed, Replayed: true}, nil
	}
	return Result{Request: opened}, nil
}

func (s Service) Withdraw(ctx context.Context, command Command) (Result, error) {
	return s.mutate(ctx, command, "seed.source.withdraw", false)
}

func (s Service) Expire(ctx context.Context, command Command) (Result, error) {
	return s.mutate(ctx, command, "seed.source.expire", true)
}

// FindForRequester reads one request back for the member who opened it.
//
// Ownership is decided the same way mutate decides it: the caller's id is
// keyed and compared against the stored requester key, because the raw id was
// never written. A request belonging to someone else is reported as absent
// rather than refused — telling a caller that an id exists but is not theirs
// is a disclosure on a surface whose whole purpose is that reaching toward
// someone is not legible.
func (s Service) FindForRequester(ctx context.Context, requestID, requesterID string) (domain.Request, error) {
	if !s.ready() {
		return domain.Request{}, ErrUnavailable
	}
	request, err := s.repository.Find(ctx, strings.TrimSpace(requestID))
	if err != nil {
		return domain.Request{}, s.translate(err)
	}
	requesterKey, err := s.key("seed-source:requester", strings.TrimSpace(requesterID))
	if err != nil {
		return domain.Request{}, err
	}
	if requesterKey != request.RequesterKey() {
		return domain.Request{}, ErrNotFound
	}
	return request, nil
}

func (s Service) mutate(ctx context.Context, command Command, action string, expire bool) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(command.RequestID))
	if err != nil {
		return Result{}, s.translate(err)
	}
	if err := s.authorizer.Require(ctx, command.ActorID, action, current.Source().Type, current.ID()); err != nil {
		return Result{}, ErrNotAvailable
	}
	actorKey, err := s.key("seed-source:requester", command.ActorID)
	if err != nil {
		return Result{}, err
	}
	if !expire && actorKey != current.RequesterKey() {
		return Result{}, ErrNotAvailable
	}
	wasApplied := current.HasCommand(command.ID)
	change := s.domainCommand(command, actorKey)
	var next domain.Request
	if expire {
		next, err = current.Expire(change)
	} else {
		next, err = current.Withdraw(change)
	}
	if err != nil {
		return Result{}, err
	}
	if wasApplied {
		return Result{Request: next, Replayed: true}, nil
	}
	if err := s.repository.Append(ctx, next, current.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			replayed, findErr := s.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && replayed.HasCommand(command.ID) {
				return Result{Request: replayed, Replayed: true}, nil
			}
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{}, s.translate(err)
	}
	return Result{Request: next}, nil
}

func (s Service) ready() bool {
	return s.repository != nil && s.authorizer != nil && s.policy != nil && s.resolver != nil &&
		s.visibility != nil && s.keyer != nil && s.ids != nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (s Service) domainCommand(c Command, actorKey string) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actorKey, ExpectedRevision: c.ExpectedRevision, ReasonCode: strings.TrimSpace(c.ReasonCode), At: s.now().UTC()}
}
func (s Service) translate(err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrCommandApplied):
		return err
	case errors.Is(err, ErrOptimisticConflict):
		return domain.ErrStaleRevision
	default:
		return ErrUnavailable
	}
}
