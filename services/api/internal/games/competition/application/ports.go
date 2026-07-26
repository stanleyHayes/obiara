package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Competition) error
	Find(context.Context, string) (domain.Competition, error)
	FindByCommand(context.Context, string) (domain.Competition, error)
	Append(context.Context, domain.Competition, uint64, string) error
}
type Authority interface {
	RequireCohortMember(context.Context, string, string) error
	RequireReviewer(context.Context, string) error
	RevalidateCohort(context.Context, string, []string) error
}
type OptIn interface {
	Revalidate(context.Context, string, string) error
}
type ResultVerifier interface {
	Revalidate(context.Context, string, string, string) error
}
type FairPlayVerifier interface {
	Revalidate(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
