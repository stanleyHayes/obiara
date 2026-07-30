package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

type Repository interface {
	Create(context.Context, domain.Round, string) error
	Find(context.Context, string) (domain.Round, error)
	FindByCommand(context.Context, string) (domain.Round, error)
	Append(context.Context, domain.Round, uint64, string) error
}

type Authority interface {
	Revalidate(context.Context, string, string, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type IDSource interface{ NewID() string }
