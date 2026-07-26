package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/domain"
)

var (
	ErrNotFound           = errors.New("assisted vouch request not found")
	ErrOptimisticConflict = errors.New("assisted vouch optimistic conflict")
	ErrCommandApplied     = errors.New("assisted vouch command already applied")
	ErrUnavailable        = errors.New("assisted vouch unavailable")
	ErrAccessDenied       = errors.New("assisted vouch access denied")
)

type Command struct {
	ID               string
	RequestID        string
	ActorID          string
	ExpectedRevision uint64
	ReasonCode       string
}

type CreateInput struct {
	SubjectID   string
	RequesterID string
	VoucherID   string
	TTL         time.Duration
}

type Result struct {
	Request  domain.Request
	Replayed bool
}

type Service struct {
	repository Repository
	authorizer Authorizer
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, authorizer: authorizer, keyer: keyer, ids: ids, now: now}
}

func (service Service) Request(ctx context.Context, command Command, input CreateInput) (Result, error) {
	if command.ActorID != input.RequesterID {
		return Result{}, ErrAccessDenied
	}
	if err := service.require(ctx, command.ActorID, "vouch.request", ""); err != nil {
		return Result{}, err
	}
	subjectKey, err := service.key("vouch:subject", input.SubjectID)
	if err != nil {
		return Result{}, err
	}
	requesterKey, err := service.key("vouch:actor", input.RequesterID)
	if err != nil {
		return Result{}, err
	}
	voucherKey, err := service.key("vouch:actor", input.VoucherID)
	if err != nil {
		return Result{}, err
	}
	request, err := domain.NewRequest(
		service.ids.NewID(), subjectKey, requesterKey, voucherKey,
		service.now().Add(input.TTL), service.domainCommand(command, requesterKey),
	)
	if err != nil {
		return Result{}, err
	}
	if err := service.repository.Create(ctx, request); err != nil {
		return service.replayCreate(ctx, command, err)
	}
	return Result{Request: request}, nil
}

func (service Service) Consent(ctx context.Context, command Command) (Result, error) {
	return service.mutate(ctx, command, "vouch.consent", func(request domain.Request, actorKey string, change domain.Command) (domain.Request, error) {
		if actorKey != request.VoucherKey() {
			return domain.Request{}, ErrAccessDenied
		}
		return request.Consent(change)
	})
}

func (service Service) Decide(ctx context.Context, command Command, decision domain.Decision) (Result, error) {
	return service.mutate(ctx, command, "vouch.decide", func(request domain.Request, _ string, change domain.Command) (domain.Request, error) {
		return request.Decide(decision, change)
	})
}

func (service Service) Withdraw(ctx context.Context, command Command) (Result, error) {
	return service.mutate(ctx, command, "vouch.withdraw", func(request domain.Request, actorKey string, change domain.Command) (domain.Request, error) {
		if actorKey != request.RequesterKey() {
			return domain.Request{}, ErrAccessDenied
		}
		return request.Withdraw(change)
	})
}

func (service Service) Expire(ctx context.Context, command Command) (Result, error) {
	return service.mutate(ctx, command, "vouch.expire", func(request domain.Request, _ string, change domain.Command) (domain.Request, error) {
		return request.Expire(change)
	})
}

func (service Service) mutate(ctx context.Context, command Command, action string, transition func(domain.Request, string, domain.Command) (domain.Request, error)) (Result, error) {
	request, err := service.repository.Find(ctx, strings.TrimSpace(command.RequestID))
	if err != nil {
		return Result{}, service.translate(err)
	}
	if err := service.require(ctx, command.ActorID, action, request.ID()); err != nil {
		return Result{}, err
	}
	actorKey, err := service.key("vouch:actor", command.ActorID)
	if err != nil {
		return Result{}, err
	}
	wasApplied := request.HasCommand(command.ID)
	next, err := transition(request, actorKey, service.domainCommand(command, actorKey))
	if err != nil {
		return Result{}, err
	}
	if wasApplied {
		return Result{Request: next, Replayed: true}, nil
	}
	if err := service.repository.Save(ctx, next, request.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			reloaded, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && reloaded.HasCommand(command.ID) {
				verified, replayErr := transition(reloaded, actorKey, service.domainCommand(command, actorKey))
				if replayErr == nil {
					return Result{Request: verified, Replayed: true}, nil
				}
				return Result{}, replayErr
			}
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{}, service.translate(err)
	}
	return Result{Request: next}, nil
}

func (service Service) replayCreate(ctx context.Context, command Command, err error) (Result, error) {
	if !errors.Is(err, ErrCommandApplied) {
		return Result{}, service.translate(err)
	}
	replayed, findErr := service.repository.FindByCommand(ctx, command.ID)
	if findErr != nil || !replayed.HasCommand(command.ID) {
		return Result{}, domain.ErrCommandMismatch
	}
	return Result{Request: replayed, Replayed: true}, nil
}

func (service Service) require(ctx context.Context, actorID, action, requestID string) error {
	if service.repository == nil || service.authorizer == nil || service.keyer == nil || service.ids == nil {
		return ErrUnavailable
	}
	if err := service.authorizer.Require(ctx, strings.TrimSpace(actorID), action, requestID); err != nil {
		return ErrAccessDenied
	}
	return nil
}

func (service Service) key(namespace, value string) (string, error) {
	key, err := service.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}

func (service Service) domainCommand(command Command, actorKey string) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ActorKey: actorKey,
		ExpectedRevision: command.ExpectedRevision, ReasonCode: strings.TrimSpace(command.ReasonCode),
		At: service.now().UTC(),
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
