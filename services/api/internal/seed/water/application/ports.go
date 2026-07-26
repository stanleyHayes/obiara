package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Water) error
	Find(context.Context, string) (domain.Water, error)
	FindByCommand(context.Context, string) (domain.Water, error)
	Append(context.Context, domain.Water, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type PairConsent interface {
	Revalidate(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
