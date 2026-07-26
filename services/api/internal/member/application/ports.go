package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

// MemberRepository is an outbound port owned by the application layer.
type MemberRepository interface {
	Create(context.Context, domain.Member) error
	FindByID(context.Context, string) (domain.Member, error)
}
