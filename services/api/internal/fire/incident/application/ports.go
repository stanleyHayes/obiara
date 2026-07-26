package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Incident) error
	FindByCase(context.Context, string) (domain.Incident, error)
	FindByCommand(context.Context, string) (domain.Incident, error)
	Append(context.Context, domain.Incident, uint64, string) error
}
type ParticipantAuthority interface {
	RequireParticipant(context.Context, string, string) error
}
type SafetyAction interface {
	Apply(context.Context, string, string, string, string) error
}
type TrustSafetyRouter interface {
	Revalidate(context.Context) error
	Route(context.Context, domain.Case) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
