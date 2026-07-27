package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type PolicyCatalog interface {
	Current(context.Context) (domain.Policy, error)
}
type Store interface {
	ClaimDue(context.Context, time.Time, int, string, time.Time) ([]domain.Candidate, error)
	Pseudonymize(context.Context, domain.Candidate, domain.Decision, string, string) error
	AggregateErase(context.Context, domain.Candidate, domain.Decision, string) error
}
type Pseudonymizer interface {
	Derive(string, uint64) (string, error)
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
