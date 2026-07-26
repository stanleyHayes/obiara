package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Intent) error
	Find(context.Context, string) (domain.Intent, error)
	Save(context.Context, domain.Intent, uint64, string) error
}
type Provider interface {
	RequestCollection(context.Context, ProviderRequest) (string, error)
}
type ProviderRequest struct {
	RequestRef, PhoneRef string
	AmountPesewas        uint64
	Currency             string
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
