package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, now: now}
}

type Command struct {
	ID               string
	CircleID         string
	ActorID          string
	ExpectedRevision uint64
}

type Result struct {
	Circle   domain.Circle
	Replayed bool
}

func (service Service) Create(ctx context.Context, command Command, kind domain.Type) (Result, error) {
	if service.repository == nil {
		return Result{}, ErrUnavailable
	}
	change := service.domainCommand(command, "circle.create", string(kind))
	circle, err := domain.Create(command.CircleID, kind, command.ActorID, change)
	if err != nil {
		return Result{}, err
	}
	return service.save(ctx, circle, command.ExpectedRevision, command.ID, func(reloaded domain.Circle) (domain.Circle, error) {
		return reloaded, reloaded.VerifyReplay(change)
	})
}

func (service Service) Request(ctx context.Context, command Command) (Result, error) {
	change := service.domainCommand(command, "membership.request", command.ActorID)
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.Request(command.ActorID, change)
	})
}

func (service Service) Approve(ctx context.Context, command Command, memberID string) (Result, error) {
	change := service.domainCommand(command, "membership.approve", memberID)
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.Approve(memberID, change)
	})
}

func (service Service) PromoteHost(ctx context.Context, command Command, memberID string) (Result, error) {
	change := service.domainCommand(command, "membership.promote_host", memberID)
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.PromoteHost(memberID, change)
	})
}

func (service Service) Leave(ctx context.Context, command Command) (Result, error) {
	change := service.domainCommand(command, "membership.leave", command.ActorID)
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.Leave(command.ActorID, change)
	})
}

func (service Service) Expel(ctx context.Context, command Command, memberID string) (Result, error) {
	change := service.domainCommand(command, "membership.expel", memberID)
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.Expel(memberID, change)
	})
}

func (service Service) SetVisibility(ctx context.Context, command Command, visibility domain.Visibility) (Result, error) {
	change := service.domainCommand(command, "circle.visibility", string(visibility))
	return service.mutate(ctx, command, func(circle domain.Circle) (domain.Circle, error) {
		return circle.SetVisibility(visibility, change)
	})
}

func (service Service) Allows(ctx context.Context, circleID, memberID string, capability domain.Capability) (bool, error) {
	if service.repository == nil {
		return false, ErrUnavailable
	}
	circle, err := service.repository.Find(ctx, strings.TrimSpace(circleID))
	if err != nil {
		return false, err
	}
	return circle.Allows(strings.TrimSpace(memberID), capability), nil
}

func (service Service) mutate(ctx context.Context, command Command, mutate func(domain.Circle) (domain.Circle, error)) (Result, error) {
	if service.repository == nil {
		return Result{}, ErrUnavailable
	}
	current, err := service.repository.Find(ctx, strings.TrimSpace(command.CircleID))
	if err != nil {
		return Result{}, err
	}
	wasApplied := current.HasCommand(command.ID)
	next, err := mutate(current)
	if err != nil {
		return Result{}, err
	}
	if wasApplied {
		return Result{Circle: next, Replayed: true}, nil
	}
	return service.save(ctx, next, command.ExpectedRevision, command.ID, mutate)
}

func (service Service) save(ctx context.Context, circle domain.Circle, expected uint64, commandID string, replay func(domain.Circle) (domain.Circle, error)) (Result, error) {
	err := service.repository.Save(ctx, circle, expected, strings.TrimSpace(commandID))
	if err == nil {
		return Result{Circle: circle}, nil
	}
	if errors.Is(err, ErrCommandAlreadyApplied) {
		reloaded, findErr := service.repository.Find(ctx, circle.ID())
		if findErr == nil && reloaded.HasCommand(commandID) {
			verified, replayErr := replay(reloaded)
			if replayErr == nil {
				return Result{Circle: verified, Replayed: true}, nil
			}
			return Result{}, replayErr
		}
		return Result{}, domain.ErrCommandMismatch
	}
	if errors.Is(err, ErrOptimisticConflict) {
		return Result{}, domain.ErrStaleRevision
	}
	return Result{}, ErrUnavailable
}

func (service Service) domainCommand(command Command, kind, payload string) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ExpectedRevision: command.ExpectedRevision,
		ActorID: strings.TrimSpace(command.ActorID), Kind: kind,
		Payload: strings.TrimSpace(payload), RecordedAt: service.now().UTC(),
	}
}
