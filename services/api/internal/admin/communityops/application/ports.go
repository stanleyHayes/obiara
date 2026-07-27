package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Authority interface {
	RequireCommunityOperator(context.Context, string) error
}
type DensitySource interface {
	CurrentDensity(context.Context, string, string) (domain.Density, error)
}
type HostSource interface {
	CurrentEligibility(context.Context, string) (domain.HostEligibility, error)
}
type NoticeCatalog interface {
	CurrentNotice(context.Context, string, string, int) (domain.Notice, error)
}
type Repository interface {
	Create(context.Context, domain.Proposal) error
	Find(context.Context, string) (domain.Proposal, error)
	Save(context.Context, domain.Proposal, uint64, string) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
