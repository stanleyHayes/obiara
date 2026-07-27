package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type DefinitionCatalog interface {
	Current(context.Context) (domain.Definition, error)
}
type AggregateSource interface {
	Current(context.Context, string) (domain.Snapshot, error)
}
type Repository interface {
	Insert(context.Context, domain.Report) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
