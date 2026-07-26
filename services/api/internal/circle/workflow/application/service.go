package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/domain"
)

var (
	ErrNotFound           = errors.New("circle workflow not found")
	ErrOptimisticConflict = errors.New("circle workflow optimistic conflict")
	ErrCommandApplied     = errors.New("circle workflow command already applied")
	ErrUnavailable        = errors.New("circle workflow unavailable")
	ErrAccessDenied       = errors.New("circle workflow access denied")
)

type Command struct {
	ID               string
	CircleID         string
	ActorID          string
	MemberID         string
	RequestID        string
	ExpectedRevision uint64
	ReasonCode       string
}

type InviteResult struct {
	Invite   domain.Invite
	Token    string
	Replayed bool
}

type RequestResult struct {
	Request  domain.Request
	Replayed bool
}

type Service struct {
	repository Repository
	authorizer Authorizer
	tokens     TokenIssuer
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, tokens TokenIssuer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, authorizer: authorizer, tokens: tokens, ids: ids, now: now}
}

func (service Service) CreateInvite(ctx context.Context, command Command, ttl time.Duration) (InviteResult, error) {
	if err := service.require(ctx, command, "invite.create", command.MemberID); err != nil {
		return InviteResult{}, err
	}
	raw, digest, err := service.tokens.NewToken()
	if err != nil {
		return InviteResult{}, ErrUnavailable
	}
	invite, err := domain.NewInvite(service.ids.NewID(), command.CircleID, digest, service.now().Add(ttl), service.domainCommand(command))
	if err != nil {
		return InviteResult{}, err
	}
	if err := service.repository.CreateInvite(ctx, invite); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			replayed, findErr := service.repository.FindInviteByCommand(ctx, command.ID)
			if findErr != nil || !replayed.HasCommand(command.ID) {
				return InviteResult{}, domain.ErrCommandMismatch
			}
			// The raw token is deliberately unrecoverable after creation.
			return InviteResult{Invite: replayed, Replayed: true}, nil
		}
		return InviteResult{}, service.translate(err)
	}
	return InviteResult{Invite: invite, Token: raw}, nil
}

func (service Service) Request(ctx context.Context, command Command) (RequestResult, error) {
	command.MemberID = command.ActorID
	if err := service.require(ctx, command, "membership.request", command.MemberID); err != nil {
		return RequestResult{}, err
	}
	request, err := domain.NewRequest(service.ids.NewID(), command.CircleID, command.MemberID, "direct", service.domainCommand(command))
	if err != nil {
		return RequestResult{}, err
	}
	if err := service.repository.CreateRequest(ctx, request); err != nil {
		return service.replayCreatedRequest(ctx, command, err)
	}
	return RequestResult{Request: request}, nil
}

func (service Service) RedeemInvite(ctx context.Context, command Command, rawToken string) (RequestResult, error) {
	digest, err := service.tokens.Digest(strings.TrimSpace(rawToken))
	if err != nil {
		return RequestResult{}, domain.ErrInvalidWorkflow
	}
	invite, err := service.repository.FindInviteByDigest(ctx, digest)
	if err != nil {
		return RequestResult{}, service.translate(err)
	}
	command.CircleID = invite.CircleID()
	command.MemberID = command.ActorID
	if err := service.require(ctx, command, "membership.request", command.MemberID); err != nil {
		return RequestResult{}, err
	}
	redeemed, err := invite.Redeem(service.domainCommand(command))
	if err != nil {
		return RequestResult{}, err
	}
	requestCommand := command
	requestCommand.ID += ".request"
	requestCommand.ExpectedRevision = 0
	request, err := domain.NewRequest(service.ids.NewID(), invite.CircleID(), command.ActorID, "invite", service.domainCommand(requestCommand))
	if err != nil {
		return RequestResult{}, err
	}
	if err := service.repository.Redeem(ctx, redeemed, request, invite.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			replayed, findErr := service.repository.FindRequestByCommand(ctx, requestCommand.ID)
			if findErr == nil && replayed.HasCommand(requestCommand.ID) {
				return RequestResult{Request: replayed, Replayed: true}, nil
			}
			return RequestResult{}, domain.ErrCommandMismatch
		}
		return RequestResult{}, service.translate(err)
	}
	return RequestResult{Request: request}, nil
}

func (service Service) Approve(ctx context.Context, command Command) (RequestResult, error) {
	return service.mutate(ctx, command, "membership.approve", func(request domain.Request, change domain.Command) (domain.Request, error) {
		return request.Approve(change)
	})
}

func (service Service) Decline(ctx context.Context, command Command) (RequestResult, error) {
	return service.mutate(ctx, command, "membership.decline", func(request domain.Request, change domain.Command) (domain.Request, error) {
		return request.Decline(change)
	})
}

func (service Service) Expel(ctx context.Context, command Command) (RequestResult, error) {
	return service.mutate(ctx, command, "membership.expel", func(request domain.Request, change domain.Command) (domain.Request, error) {
		return request.Expel(change)
	})
}

func (service Service) mutate(ctx context.Context, command Command, action string, transition func(domain.Request, domain.Command) (domain.Request, error)) (RequestResult, error) {
	request, err := service.repository.FindRequest(ctx, strings.TrimSpace(command.RequestID))
	if err != nil {
		return RequestResult{}, service.translate(err)
	}
	command.CircleID = request.CircleID()
	command.MemberID = request.MemberID()
	if err := service.require(ctx, command, action, request.MemberID()); err != nil {
		return RequestResult{}, err
	}
	wasApplied := request.HasCommand(command.ID)
	next, err := transition(request, service.domainCommand(command))
	if err != nil {
		return RequestResult{}, err
	}
	if wasApplied {
		return RequestResult{Request: next, Replayed: true}, nil
	}
	if err := service.repository.SaveRequest(ctx, next, request.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			reloaded, findErr := service.repository.FindRequestByCommand(ctx, command.ID)
			if findErr == nil && reloaded.HasCommand(command.ID) {
				verified, replayErr := transition(reloaded, service.domainCommand(command))
				if replayErr == nil {
					return RequestResult{Request: verified, Replayed: true}, nil
				}
				return RequestResult{}, replayErr
			}
			return RequestResult{}, domain.ErrCommandMismatch
		}
		return RequestResult{}, service.translate(err)
	}
	return RequestResult{Request: next}, nil
}

func (service Service) replayCreatedRequest(ctx context.Context, command Command, err error) (RequestResult, error) {
	if !errors.Is(err, ErrCommandApplied) {
		return RequestResult{}, service.translate(err)
	}
	replayed, findErr := service.repository.FindRequestByCommand(ctx, command.ID)
	if findErr != nil || !replayed.HasCommand(command.ID) {
		return RequestResult{}, domain.ErrCommandMismatch
	}
	return RequestResult{Request: replayed, Replayed: true}, nil
}

func (service Service) require(ctx context.Context, command Command, action, memberID string) error {
	if service.repository == nil || service.authorizer == nil || service.tokens == nil || service.ids == nil {
		return ErrUnavailable
	}
	if err := service.authorizer.Require(ctx, strings.TrimSpace(command.ActorID), strings.TrimSpace(command.CircleID), action, strings.TrimSpace(memberID)); err != nil {
		return ErrAccessDenied
	}
	return nil
}

func (service Service) domainCommand(command Command) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ActorID: strings.TrimSpace(command.ActorID),
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
