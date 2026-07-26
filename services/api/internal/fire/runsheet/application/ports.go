package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.RunSheet) error
	Find(context.Context, string) (domain.RunSheet, error)
	FindByCommand(context.Context, string) (domain.RunSheet, error)
	Append(context.Context, domain.RunSheet, uint64, string) error
}
type Authority interface {
	RequireHostOrCohost(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
