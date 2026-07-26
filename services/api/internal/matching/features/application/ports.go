package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Catalog interface {
	Put(context.Context, domain.Definition) error
	FindDefinition(context.Context, string, uint64) (domain.Definition, error)
	Current(context.Context, string) (domain.Definition, error)
	ListCurrent(context.Context) ([]domain.Definition, error)
}
type GrantRepository interface {
	Create(context.Context, domain.Grant) error
	Find(context.Context, string, string) (domain.Grant, error)
	FindByCommand(context.Context, string) (domain.Grant, error)
	ListEffective(context.Context, string) ([]domain.Grant, error)
	Append(context.Context, domain.Grant, uint64, string) error
}
type DecisionRepository interface {
	CreateDecision(context.Context, domain.Decision) error
	FindDecision(context.Context, string) (domain.Decision, error)
}
type Authority interface {
	RequireMember(context.Context, string, string) error
	RequirePair(context.Context, string, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
