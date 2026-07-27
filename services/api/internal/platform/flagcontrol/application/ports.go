package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	CreateWithAudit(context.Context, domain.Proposal, domain.Audit) error
	Find(context.Context, string) (domain.Proposal, error)
	FindByCommand(context.Context, string) (domain.Proposal, error)
	SaveWithAudit(context.Context, domain.Proposal, uint64, domain.Audit) error
}
type Authority interface {
	RequireSteppedController(context.Context, string, domain.Capability) (string, error)
}
type Runtime interface {
	Apply(context.Context, domain.Environment, domain.Market, domain.RuntimeChange) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
