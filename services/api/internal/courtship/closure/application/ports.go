package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	// Create inserts the initial state. Save is for an aggregate that
	// already has events; it cannot open one.
	Create(context.Context, domain.Closure) error
	Find(context.Context, string) (domain.Closure, error)
	Save(context.Context, domain.Closure, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
