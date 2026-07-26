package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/coldstart/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

// Authority revalidates the requester's current authority to generate private
// introductions. Generate calls it both before input reads and immediately
// before returning.
type Authority interface {
	AuthorizeColdStart(context.Context, string) error
}

// Preferences returns privacy-minimized, explicit reciprocal summaries.
type Preferences interface {
	Reciprocal(context.Context, string, int) ([]domain.ReciprocalPreference, error)
}

// TrustPaths returns privacy-scoped summaries only, never raw paths.
type TrustPaths interface {
	Summaries(context.Context, string, int) ([]domain.TrustSummary, error)
}

// Visibility is revalidated immediately before each candidate projection.
type Visibility interface {
	CanIntroduce(context.Context, string, string) (bool, error)
}
