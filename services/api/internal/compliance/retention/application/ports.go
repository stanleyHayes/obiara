package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Record) error
	Find(context.Context, string) (domain.Record, error)
	FindByCommand(context.Context, string) (domain.Record, error)
	Append(context.Context, domain.Record, uint64, string) error
}
type PolicyCatalog interface {
	Find(context.Context, string, string, uint64) (domain.Policy, error)
}
type Authority interface {
	RequireSubject(context.Context, string, string) error
	RequireLegalHoldOfficer(context.Context, string) error
}
type ErasureVerifier interface {
	Verify(context.Context, string, string, string) (string, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
