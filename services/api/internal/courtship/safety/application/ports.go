package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	// Create inserts the initial state. Save is for an aggregate that
	// already has events; it cannot open one.
	Create(context.Context, domain.Safety) error
	Find(context.Context, string) (domain.Safety, error)
	Save(context.Context, domain.Safety, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
