package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/domain"
)

type Command struct {
	ID               string
	ProgressionID    string
	ActorID          string
	ReasonCode       string
	ExpectedRevision uint64
}
type Pair struct{ FirstMemberID, SecondMemberID string }
type Result struct {
	Progression domain.Progression
	Replayed    bool
}
type Service struct {
	repository Repository
	authorizer Authorizer
	membership Membership
	evidence   ThemeOneEvidence
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, membership Membership, evidence ThemeOneEvidence, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository, authorizer, membership, evidence, keyer, ids, now}
}

func (service Service) Open(ctx context.Context, command Command, pair Pair) (Result, error) {
	if !service.ready() || command.ActorID != pair.FirstMemberID {
		return Result{}, ErrNotAvailable
	}
	if service.authorizer.Require(ctx, command.ActorID, "courtship.themeprogression.open", "") != nil ||
		service.membership.RevalidatePair(ctx, pair.FirstMemberID, pair.SecondMemberID) != nil {
		return Result{}, ErrNotAvailable
	}
	evidenceRef, revealed, err := service.evidence.BothRevealed(ctx, pair.FirstMemberID, pair.SecondMemberID)
	if err != nil || !revealed {
		return Result{}, ErrNotAvailable
	}
	first, err := service.key("member", pair.FirstMemberID)
	if err != nil {
		return Result{}, err
	}
	second, err := service.key("member", pair.SecondMemberID)
	if err != nil {
		return Result{}, err
	}
	evidenceKey, err := service.key("theme-one-evidence", evidenceRef)
	if err != nil {
		return Result{}, err
	}
	progression, err := domain.Open(service.ids.NewID(), []string{first, second}, evidenceKey, service.command(command, first))
	if err != nil {
		return Result{}, err
	}
	if err := service.repository.Create(ctx, progression); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			stored, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && equivalentOpening(progression, stored) {
				return Result{Progression: stored, Replayed: true}, nil
			}
			if findErr == nil {
				return Result{}, domain.ErrCommandMismatch
			}
		}
		return Result{}, service.translate(err)
	}
	return Result{Progression: progression}, nil
}

func (service Service) Submit(ctx context.Context, command Command, theme domain.ThemeNumber, encryptedContentRef string) (Result, error) {
	if !service.ready() {
		return Result{}, ErrUnavailable
	}
	progression, err := service.repository.Find(ctx, strings.TrimSpace(command.ProgressionID))
	if err != nil {
		return Result{}, service.translate(err)
	}
	if service.authorizer.Require(ctx, command.ActorID, "courtship.themeprogression.submit", progression.ID()) != nil ||
		service.membership.RequireParticipant(ctx, command.ActorID, progression.ID()) != nil {
		return Result{}, ErrNotAvailable
	}
	actor, err := service.key("member", command.ActorID)
	if err != nil {
		return Result{}, err
	}
	content, err := service.key("encrypted-content", encryptedContentRef)
	if err != nil {
		return Result{}, err
	}
	domainCommand := service.command(command, actor)
	replayed := progression.HasCommand(command.ID)
	next, err := progression.Submit(theme, content, domainCommand)
	if err != nil {
		return Result{}, err
	}
	if replayed {
		return Result{Progression: next, Replayed: true}, nil
	}
	if err := service.repository.Append(ctx, next, progression.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			stored, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && stored.ID() == progression.ID() {
				verified, verifyErr := stored.Submit(theme, content, domainCommand)
				if verifyErr != nil {
					return Result{}, verifyErr
				}
				return Result{Progression: verified, Replayed: true}, nil
			}
			if findErr == nil {
				return Result{}, domain.ErrCommandMismatch
			}
		}
		return Result{}, service.translate(err)
	}
	return Result{Progression: next}, nil
}

func (service Service) ready() bool {
	return service.repository != nil && service.authorizer != nil && service.membership != nil &&
		service.evidence != nil && service.keyer != nil && service.ids != nil
}
func (service Service) key(namespace, value string) (string, error) {
	key, err := service.keyer.Key("courtship-themeprogression:"+namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (service Service) command(command Command, actorKey string) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ActorKey: actorKey,
		ReasonCode:       strings.TrimSpace(command.ReasonCode),
		ExpectedRevision: command.ExpectedRevision, At: service.now().UTC(),
	}
}
func (service Service) translate(err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrCommandApplied):
		return err
	case errors.Is(err, ErrOptimisticConflict):
		return domain.ErrStaleRevision
	default:
		return ErrUnavailable
	}
}
func equivalentOpening(candidate, stored domain.Progression) bool {
	if !slices.Equal(candidate.Members(), stored.Members()) {
		return false
	}
	left, right := candidate.Events(), stored.Events()
	return len(left) == 1 && len(right) > 0 && left[0].CommandID == right[0].CommandID &&
		left[0].ActorKey == right[0].ActorKey && left[0].EvidenceKey == right[0].EvidenceKey &&
		left[0].ReasonCode == right[0].ReasonCode
}
