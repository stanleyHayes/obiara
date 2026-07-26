package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

var ErrIdempotencyKeyRequired = errors.New("idempotency key is required")

type RegisterMemberCommand struct {
	ID             string
	Email          string
	IdempotencyKey string
}

type RegisterMember struct {
	repository MemberRepository
	now        func() time.Time
}

func NewRegisterMember(repository MemberRepository, now func() time.Time) RegisterMember {
	return RegisterMember{repository: repository, now: now}
}

func (handler RegisterMember) Handle(ctx context.Context, command RegisterMemberCommand) (domain.Member, error) {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return domain.Member{}, ErrIdempotencyKeyRequired
	}

	member, err := domain.NewMember(command.ID, command.Email, handler.now())
	if err != nil {
		return domain.Member{}, err
	}
	if err := handler.repository.Create(ctx, member); err != nil {
		return domain.Member{}, err
	}
	return member, nil
}
