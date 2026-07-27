package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type MasterCatalog interface {
	Current(context.Context) (domain.Master, error)
}
type Authority interface {
	RequireAuthor(context.Context, string) error
	RequireReviewer(context.Context, string, domain.ReviewStage) error
	RequireApprover(context.Context, string) error
}
type Repository interface {
	Create(context.Context, domain.Pack) error
	Find(context.Context, string) (domain.Pack, error)
	Save(context.Context, domain.Pack, uint64, string) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
