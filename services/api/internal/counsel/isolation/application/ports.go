package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/counsel/isolation/domain"
)

const SafetyEscalationPurpose = "counsel.explicit_safety_escalation"

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

// Scope answers only whether both opaque members are currently within the
// private counsel scope. It never returns an attendance list.
type Scope interface {
	ContainsBoth(context.Context, string, string, string) (bool, error)
}

type Consent interface {
	CurrentAllows(context.Context, string, string, uint64) (bool, error)
}

type Authority interface {
	AuthorizeEscalation(context.Context, string, string) error
}

type SafetySink interface {
	Publish(context.Context, domain.SafetyEvent) error
}

type IDSource interface {
	NewID() string
}
