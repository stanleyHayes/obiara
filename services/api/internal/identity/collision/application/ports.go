package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	// RegisterSignal atomically records a pseudonymous subject binding and
	// reports whether the signal is now bound to more than one subject.
	RegisterSignal(context.Context, domain.Kind, string, string) (bool, error)
	Create(context.Context, domain.Case, domain.AuditEvent) (domain.Case, bool, error)
	FindByID(context.Context, string) (domain.Case, error)
	Resolve(context.Context, domain.Case, domain.AuditEvent, int64) error
}

type Keyer interface {
	Key(namespace, value string) (string, error)
}
