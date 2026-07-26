package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/scamarc/domain"
)

const MonitoringPurpose = "safety.scam_arc_monitoring"

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Consent interface {
	CurrentAllows(context.Context, string, string, uint64) (bool, error)
}

type Authority interface {
	AuthorizeEvaluation(context.Context, string) error
}

type RuleCatalog interface {
	Current(context.Context) (domain.RuleSet, error)
}

// EventSource returns already-categorized opaque summaries, never message,
// voice, payment, model, or vendor output.
type EventSource interface {
	Current(context.Context, string, int) (string, []domain.Event, error)
}

type HumanRoute interface {
	Route(context.Context, domain.Signal) error
}

type IDSource interface {
	NewID() string
}

type Clock interface {
	Now() time.Time
}
