package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
)

var (
	ErrInvalid     = errors.New("invalid membership request")
	ErrNotFound    = errors.New("membership pass not found")
	ErrConflict    = errors.New("membership pass conflict")
	ErrApplied     = errors.New("membership command already applied")
	ErrUnavailable = errors.New("membership service unavailable")
)

type Service struct {
	repository    Repository
	confirmations RefundConfirmationSource
	ids           IDSource
	clock         Clock
}

func New(repository Repository, confirmations RefundConfirmationSource, ids IDSource, clock Clock) Service {
	return Service{repository: repository, confirmations: confirmations, ids: ids, clock: clock}
}

func (service Service) Current(ctx context.Context, memberKey string) (domain.Pass, error) {
	repository, ok := service.repository.(MemberRepository)
	if !ok || memberKey == "" {
		return domain.Pass{}, ErrUnavailable
	}
	return repository.FindForMember(ctx, memberKey)
}

func (service Service) Grant(ctx context.Context, memberKey, passID, receiptRef string, passVersion uint64, paidThrough time.Time, grace time.Duration, commandID string) (domain.Pass, error) {
	if service.repository == nil || service.ids == nil || service.clock == nil {
		return domain.Pass{}, ErrInvalid
	}
	now := service.clock.Now()
	pass, err := domain.New(domain.Grant{
		ID: service.ids.NewID(), MemberKey: memberKey, PassID: passID,
		PassVersion: passVersion, ReceiptRef: receiptRef, GrantedAt: now,
		PaidThrough: paidThrough, GraceUntil: paidThrough.Add(grace), GraceDuration: grace,
	}, commandID)
	if err != nil {
		return domain.Pass{}, ErrInvalid
	}
	if err = service.repository.Create(ctx, pass); err != nil {
		return domain.Pass{}, err
	}
	return pass, nil
}

func (service Service) Cancel(ctx context.Context, id, commandID string) (domain.Pass, error) {
	pass, err := service.repository.Find(ctx, id)
	if err != nil {
		return domain.Pass{}, err
	}
	next, err := pass.Cancel(commandID, service.clock.Now())
	if err != nil {
		return domain.Pass{}, ErrInvalid
	}
	if err = service.repository.Save(ctx, next, pass.Revision(), commandID); err != nil {
		return domain.Pass{}, err
	}
	return next, nil
}

func (service Service) RequestRefund(ctx context.Context, id, commandID string) (domain.Pass, error) {
	pass, err := service.repository.Find(ctx, id)
	if err != nil {
		return domain.Pass{}, err
	}
	requestRef := service.ids.NewID()
	next, err := pass.RequestRefund(commandID, requestRef, service.clock.Now())
	if err != nil {
		return domain.Pass{}, ErrInvalid
	}
	if err = service.repository.Save(ctx, next, pass.Revision(), commandID); err != nil {
		return domain.Pass{}, err
	}
	return next, nil
}

func (service Service) ConfirmRefund(ctx context.Context, id, commandID string) (domain.Pass, error) {
	if service.confirmations == nil {
		return domain.Pass{}, ErrInvalid
	}
	pass, err := service.repository.Find(ctx, id)
	if err != nil {
		return domain.Pass{}, err
	}
	state := pass.State()
	providerRef, confirmedAt, err := service.confirmations.Confirmed(ctx, state.RefundRequestRef)
	if err != nil || confirmedAt.IsZero() {
		return domain.Pass{}, ErrUnavailable
	}
	next, err := pass.ConfirmRefund(commandID, state.RefundRequestRef, providerRef, confirmedAt)
	if err != nil {
		return domain.Pass{}, ErrInvalid
	}
	if err = service.repository.Save(ctx, next, pass.Revision(), commandID); err != nil {
		return domain.Pass{}, err
	}
	return next, nil
}
