package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type DefinitionCatalog interface {
	Current(context.Context) (domain.Definition, error)
}
type AggregateSource interface {
	CurrentQuarter(context.Context, string) (domain.Snapshot, error)
}
type Repository interface {
	Insert(context.Context, domain.Report) error
	Find(context.Context, string, uint64) (domain.Report, error)
}
type Authority interface {
	RequireProjector(context.Context, string) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
