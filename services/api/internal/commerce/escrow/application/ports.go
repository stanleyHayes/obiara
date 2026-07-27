package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Escrow) error
	Find(context.Context, string) (domain.Escrow, error)
	Save(context.Context, domain.Escrow, uint64, string) error
}
type Ledger interface {
	// RecordSettlement is idempotent by commandID.
	RecordSettlement(context.Context, string, domain.Statement) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
