// Package application is the organization's use cases.
//
// There is no organization sign-in here on purpose. Codes are issued by
// Obiara staff on an organization's behalf (agent_plan.md §41), so every
// entry point takes an operator, and the operator is keyed before anything is
// written down.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
)

type Service struct {
	repository Repository
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func New(repository Repository, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, keyer: keyer, ids: ids, now: now}
}

type RegisterCommand struct {
	CommandID    string
	OperatorID   string
	ReasonCode   string
	Name         string
	BillingEmail string
}

type Result struct {
	Organization domain.Organization
	Replayed     bool
}

// Register records a new organization.
func (service Service) Register(ctx context.Context, command RegisterCommand) (Result, error) {
	if !service.ready() {
		return Result{}, ErrUnavailable
	}
	actorKey, err := service.actor(command.OperatorID)
	if err != nil {
		return Result{}, err
	}
	organization, err := domain.Register(
		service.ids.NewID(), command.Name, command.BillingEmail,
		domain.Command{
			ID:         strings.TrimSpace(command.CommandID),
			ActorKey:   actorKey,
			ReasonCode: strings.TrimSpace(command.ReasonCode),
			At:         service.now().UTC(),
		},
	)
	if err != nil {
		return Result{}, err
	}
	if err := service.repository.Create(ctx, organization); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			// The same command id already registered somebody. Answering with
			// that organization is what makes a retried registration safe;
			// answering with a second one would leave two bodies where the
			// operator asked for one.
			existing, findErr := service.repository.FindByCommand(ctx, command.CommandID)
			if findErr != nil {
				return Result{}, ErrUnavailable
			}
			return Result{Organization: existing, Replayed: true}, nil
		}
		return Result{}, err
	}
	return Result{Organization: organization}, nil
}

type ChangeCommand struct {
	CommandID        string
	OperatorID       string
	ReasonCode       string
	OrganizationID   string
	ExpectedRevision uint64
	// Name is read only by Rename.
	Name string
}

// Suspend stops an organization being used as an issuer.
func (service Service) Suspend(ctx context.Context, command ChangeCommand) (Result, error) {
	return service.change(ctx, command, func(
		organization domain.Organization, applied domain.Command,
	) (domain.Organization, error) {
		return organization.Suspend(applied)
	})
}

// Restore returns a suspended organization to use.
func (service Service) Restore(ctx context.Context, command ChangeCommand) (Result, error) {
	return service.change(ctx, command, func(
		organization domain.Organization, applied domain.Command,
	) (domain.Organization, error) {
		return organization.Restore(applied)
	})
}

// Rename records a change of name.
func (service Service) Rename(ctx context.Context, command ChangeCommand) (Result, error) {
	return service.change(ctx, command, func(
		organization domain.Organization, applied domain.Command,
	) (domain.Organization, error) {
		return organization.Rename(command.Name, applied)
	})
}

func (service Service) change(
	ctx context.Context,
	command ChangeCommand,
	apply func(domain.Organization, domain.Command) (domain.Organization, error),
) (Result, error) {
	if !service.ready() {
		return Result{}, ErrUnavailable
	}
	actorKey, err := service.actor(command.OperatorID)
	if err != nil {
		return Result{}, err
	}
	current, err := service.repository.FindByID(ctx, strings.TrimSpace(command.OrganizationID))
	if err != nil {
		return Result{}, err
	}
	next, err := apply(current, domain.Command{
		ID:               strings.TrimSpace(command.CommandID),
		ActorKey:         actorKey,
		ReasonCode:       strings.TrimSpace(command.ReasonCode),
		ExpectedRevision: command.ExpectedRevision,
		At:               service.now().UTC(),
	})
	if err != nil {
		return Result{}, err
	}
	if next.Revision() == current.Revision() {
		// The aggregate recognised a replay and returned itself unchanged.
		// Writing would append nothing and could only fail.
		return Result{Organization: next, Replayed: true}, nil
	}
	if err := service.repository.Append(ctx, next, current.Revision(), command.CommandID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			existing, findErr := service.repository.FindByCommand(ctx, command.CommandID)
			if findErr == nil {
				return Result{Organization: existing, Replayed: true}, nil
			}
		}
		return Result{}, err
	}
	return Result{Organization: next}, nil
}

// List reads the organizations, for the operator surface that issues codes.
func (service Service) List(ctx context.Context, limit int) ([]domain.Organization, error) {
	if !service.ready() {
		return nil, ErrUnavailable
	}
	return service.repository.List(ctx, limit)
}

// Find reads one, which is what the promotion context asks before it will
// issue a code in somebody's name.
func (service Service) Find(ctx context.Context, id string) (domain.Organization, error) {
	if !service.ready() {
		return domain.Organization{}, ErrUnavailable
	}
	return service.repository.FindByID(ctx, strings.TrimSpace(id))
}

func (service Service) ready() bool {
	return service.repository != nil && service.keyer != nil &&
		service.ids != nil && service.now != nil
}

// actor keys the operator. An audit trail has to prove somebody acted without
// being a directory of who — the same rule every other trail in this codebase
// follows.
func (service Service) actor(operatorID string) (string, error) {
	key, err := service.keyer.Key("organization_operator", strings.TrimSpace(operatorID))
	if err != nil || key == "" {
		return "", ErrUnavailable
	}
	return key, nil
}
