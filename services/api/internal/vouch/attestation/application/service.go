package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/domain"
)

var (
	ErrNotFound           = errors.New("vouch attestation not found")
	ErrOptimisticConflict = errors.New("vouch attestation optimistic conflict")
	ErrCommandApplied     = errors.New("vouch attestation command applied")
	ErrUnavailable        = errors.New("vouch attestation unavailable")
	ErrAccessDenied       = errors.New("vouch attestation access denied")
	ErrPolicyDenied       = errors.New("vouch reputation stake denied by policy")
)

type Command struct {
	ID               string
	AttestationID    string
	ActorID          string
	ExpectedRevision uint64
	ReasonCode       string
}
type Proposal struct {
	SubjectID  string
	VoucherID  string
	ScopeKind  string
	ScopeID    string
	StakeUnits uint8
	TTL        time.Duration
}
type Result struct {
	Attestation domain.Attestation
	Replayed    bool
}
type Service struct {
	repository Repository
	authorizer Authorizer
	keyer      Keyer
	policy     StakePolicy
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, keyer Keyer, policy StakePolicy, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, authorizer: authorizer, keyer: keyer, policy: policy, ids: ids, now: now}
}

func (s Service) Propose(ctx context.Context, command Command, proposal Proposal) (Result, error) {
	if command.ActorID != proposal.VoucherID {
		return Result{}, ErrAccessDenied
	}
	if err := s.require(ctx, command.ActorID, "vouch.attestation.propose", ""); err != nil {
		return Result{}, err
	}
	policyVersion, err := s.policy.Validate(ctx, proposal.ScopeKind, proposal.StakeUnits)
	if err != nil {
		return Result{}, ErrPolicyDenied
	}
	subjectKey, err := s.key("vouch-attestation:subject", proposal.SubjectID)
	if err != nil {
		return Result{}, err
	}
	voucherKey, err := s.key("vouch-attestation:actor", proposal.VoucherID)
	if err != nil {
		return Result{}, err
	}
	scopeKey, err := s.key("vouch-attestation:scope", proposal.ScopeID)
	if err != nil {
		return Result{}, err
	}
	attestation, err := domain.Propose(
		s.ids.NewID(), subjectKey, voucherKey,
		domain.SubjectScope{Kind: proposal.ScopeKind, Key: scopeKey},
		proposal.StakeUnits, policyVersion, s.now().Add(proposal.TTL),
		s.domainCommand(command, voucherKey),
	)
	if err != nil {
		return Result{}, err
	}
	if err := s.repository.Create(ctx, attestation); err != nil {
		return s.replayCreate(ctx, command, err)
	}
	return Result{Attestation: attestation}, nil
}

func (s Service) Consent(ctx context.Context, command Command) (Result, error) {
	return s.mutate(ctx, command, "vouch.attestation.consent", func(a domain.Attestation, actor string, change domain.Command) (domain.Attestation, error) {
		if actor != a.VoucherKey() {
			return domain.Attestation{}, ErrAccessDenied
		}
		return a.Consent(change)
	})
}
func (s Service) Revoke(ctx context.Context, command Command) (Result, error) {
	return s.mutate(ctx, command, "vouch.attestation.revoke", func(a domain.Attestation, _ string, change domain.Command) (domain.Attestation, error) {
		return a.Revoke(change)
	})
}
func (s Service) Expire(ctx context.Context, command Command) (Result, error) {
	return s.mutate(ctx, command, "vouch.attestation.expire", func(a domain.Attestation, _ string, change domain.Command) (domain.Attestation, error) {
		return a.Expire(change)
	})
}

func (s Service) mutate(ctx context.Context, command Command, action string, transition func(domain.Attestation, string, domain.Command) (domain.Attestation, error)) (Result, error) {
	current, err := s.repository.Find(ctx, strings.TrimSpace(command.AttestationID))
	if err != nil {
		return Result{}, s.translate(err)
	}
	if err := s.require(ctx, command.ActorID, action, current.ID()); err != nil {
		return Result{}, err
	}
	actorKey, err := s.key("vouch-attestation:actor", command.ActorID)
	if err != nil {
		return Result{}, err
	}
	wasApplied := current.HasCommand(command.ID)
	next, err := transition(current, actorKey, s.domainCommand(command, actorKey))
	if err != nil {
		return Result{}, err
	}
	if wasApplied {
		return Result{Attestation: next, Replayed: true}, nil
	}
	if err := s.repository.Append(ctx, next, current.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			reloaded, findErr := s.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && reloaded.HasCommand(command.ID) {
				verified, replayErr := transition(reloaded, actorKey, s.domainCommand(command, actorKey))
				if replayErr == nil {
					return Result{Attestation: verified, Replayed: true}, nil
				}
				return Result{}, replayErr
			}
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{}, s.translate(err)
	}
	return Result{Attestation: next}, nil
}
func (s Service) replayCreate(ctx context.Context, command Command, err error) (Result, error) {
	if !errors.Is(err, ErrCommandApplied) {
		return Result{}, s.translate(err)
	}
	replayed, findErr := s.repository.FindByCommand(ctx, command.ID)
	if findErr != nil || !replayed.HasCommand(command.ID) {
		return Result{}, domain.ErrCommandMismatch
	}
	return Result{Attestation: replayed, Replayed: true}, nil
}
func (s Service) require(ctx context.Context, actor, action, id string) error {
	if s.repository == nil || s.authorizer == nil || s.keyer == nil || s.policy == nil || s.ids == nil {
		return ErrUnavailable
	}
	if err := s.authorizer.Require(ctx, strings.TrimSpace(actor), action, id); err != nil {
		return ErrAccessDenied
	}
	return nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (s Service) domainCommand(command Command, actorKey string) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ActorKey: actorKey,
		ExpectedRevision: command.ExpectedRevision, ReasonCode: strings.TrimSpace(command.ReasonCode),
		At: s.now().UTC(),
	}
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
