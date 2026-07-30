package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/domain"
)

type Repository interface {
	Create(context.Context, domain.Cohort) error
	Find(context.Context, string) (domain.Cohort, error)
	FindByCommand(context.Context, string) (domain.Cohort, error)
	Append(context.Context, domain.Cohort, uint64, string) error
}

type AdminAuthority interface {
	RequireTournamentManager(context.Context, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type IDSource interface{ NewID() string }
