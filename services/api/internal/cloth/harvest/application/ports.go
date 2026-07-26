package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Harvest) error
	Find(context.Context, string) (domain.Harvest, error)
	FindByCommand(context.Context, string) (domain.Harvest, error)
	FindByHandoff(context.Context, string) (domain.Harvest, error)
	Append(context.Context, domain.Harvest, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type PairPolicy interface {
	Revalidate(context.Context, string, string) error
}
type Ownership interface {
	Revalidate(context.Context, string, string, string) error
}
type RecipeValidator interface {
	Revalidate(context.Context, domain.Payload) error
}
type ProviderAuth interface {
	Require(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
