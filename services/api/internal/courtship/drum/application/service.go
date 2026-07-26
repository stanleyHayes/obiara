package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/domain"
)

type Command struct {
	ID               string
	StageID          string
	ActorID          string
	ReasonCode       string
	ExpectedRevision uint64
}
type Pair struct {
	FirstMemberID  string
	SecondMemberID string
}
type Result struct {
	Stage    domain.Stage
	Replayed bool
}
type Service struct {
	repository Repository
	authorizer Authorizer
	membership Membership
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, membership Membership, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository, authorizer, membership, keyer, ids, now}
}

func (service Service) Open(ctx context.Context, command Command, pair Pair, voiceRef string) (Result, error) {
	if !service.ready() || command.ActorID != pair.FirstMemberID || strings.TrimSpace(voiceRef) == "" {
		return Result{}, ErrNotAvailable
	}
	if service.authorizer.Require(ctx, command.ActorID, "courtship.drum.open", "") != nil ||
		service.membership.RevalidatePair(ctx, pair.FirstMemberID, pair.SecondMemberID) != nil {
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
	voice, err := service.key("voice", voiceRef)
	if err != nil {
		return Result{}, err
	}
	stage, err := domain.Open(service.ids.NewID(), []string{first, second}, voice, service.command(command, first))
	if err != nil {
		return Result{}, err
	}
	if err := service.repository.Create(ctx, stage); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			stored, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil {
				if !equivalentOpening(stage, stored) {
					return Result{}, domain.ErrCommandMismatch
				}
				return Result{Stage: stored, Replayed: true}, nil
			}
		}
		return Result{}, service.translate(err)
	}
	return Result{Stage: stage}, nil
}

func (service Service) AddVoice(ctx context.Context, command Command, voiceRef string) (Result, error) {
	return service.add(ctx, command, domain.MediumVoice, "voice", voiceRef)
}
func (service Service) AddText(ctx context.Context, command Command, textRef string) (Result, error) {
	return service.add(ctx, command, domain.MediumText, "text", textRef)
}
func (service Service) add(ctx context.Context, command Command, medium domain.Medium, namespace, contentRef string) (Result, error) {
	if !service.ready() {
		return Result{}, ErrUnavailable
	}
	stage, err := service.repository.Find(ctx, strings.TrimSpace(command.StageID))
	if err != nil {
		return Result{}, service.translate(err)
	}
	if service.authorizer.Require(ctx, command.ActorID, "courtship.drum.turn", stage.ID()) != nil ||
		service.membership.RequireParticipant(ctx, command.ActorID, stage.ID()) != nil {
		return Result{}, ErrNotAvailable
	}
	actor, err := service.key("member", command.ActorID)
	if err != nil {
		return Result{}, err
	}
	content, err := service.key(namespace, contentRef)
	if err != nil {
		return Result{}, err
	}
	replayed := stage.HasCommand(command.ID)
	domainCommand := service.command(command, actor)
	next, err := stage.Add(medium, content, domainCommand)
	if err != nil {
		return Result{}, err
	}
	if replayed {
		return Result{Stage: next, Replayed: true}, nil
	}
	if err := service.repository.Append(ctx, next, stage.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			stored, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil {
				if stored.ID() != stage.ID() {
					return Result{}, domain.ErrCommandMismatch
				}
				verified, verifyErr := stored.Add(medium, content, domainCommand)
				if verifyErr != nil {
					return Result{}, verifyErr
				}
				return Result{Stage: verified, Replayed: true}, nil
			}
		}
		return Result{}, service.translate(err)
	}
	return Result{Stage: next}, nil
}

func equivalentOpening(candidate, stored domain.Stage) bool {
	if !slices.Equal(candidate.Members(), stored.Members()) {
		return false
	}
	candidateBeats, storedBeats := candidate.Beats(), stored.Beats()
	if len(candidateBeats) != 1 || len(storedBeats) == 0 {
		return false
	}
	left, right := candidateBeats[0], storedBeats[0]
	return left.CommandID == right.CommandID && left.ActorKey == right.ActorKey &&
		left.Medium == right.Medium && left.ContentKey == right.ContentKey &&
		left.ReasonCode == right.ReasonCode
}

func (service Service) ready() bool {
	return service.repository != nil && service.authorizer != nil && service.membership != nil &&
		service.keyer != nil && service.ids != nil
}
func (service Service) key(namespace, value string) (string, error) {
	key, err := service.keyer.Key("courtship-drum:"+namespace, strings.TrimSpace(value))
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
