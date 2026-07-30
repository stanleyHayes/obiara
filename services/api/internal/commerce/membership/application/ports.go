package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	Create(context.Context, domain.Pass) error
	Find(context.Context, string) (domain.Pass, error)
	Save(context.Context, domain.Pass, uint64, string) error
}

type MemberRepository interface {
	FindForMember(context.Context, string) (domain.Pass, error)
}

// RefundConfirmation is a provider-confirmed opaque fact supplied by an
// adapter owned elsewhere. This boundary makes no provider call.
type RefundConfirmationSource interface {
	Confirmed(context.Context, string) (providerRef string, confirmedAt time.Time, err error)
}

type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
