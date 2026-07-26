package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

var (
	ErrAttemptNotFound        = errors.New("liveness attempt not found")
	ErrCommandConflict        = errors.New("liveness command conflicts with existing input")
	ErrOptimisticConflict     = errors.New("liveness attempt changed")
	ErrReviewQueueUnavailable = errors.New("liveness manual-review queue unavailable")
)

type ProviderRequest struct {
	CommandID        string
	AttemptID        string
	VoiceArtifactRef string
	FaceArtifactRef  string
}

type ProviderOutcome string

const (
	OutcomeLive      ProviderOutcome = "live"
	OutcomeNotLive   ProviderOutcome = "not_live"
	OutcomeUncertain ProviderOutcome = "uncertain"
)

type ProviderResult struct {
	Outcome     ProviderOutcome
	ProviderRef string
}

type Provider interface {
	Assess(context.Context, ProviderRequest) (ProviderResult, error)
}

type AttemptStore interface {
	// Create must be idempotent by CommandID and return the existing attempt
	// with replayed=true. Reusing a command for different keyed input returns
	// ErrCommandConflict.
	Create(context.Context, domain.Attempt) (domain.Attempt, bool, error)
	FindByID(context.Context, string) (domain.Attempt, error)
	Update(context.Context, domain.Attempt, uint64) error
}

type ReviewTask struct {
	AttemptID        string
	VoiceArtifactRef string
	FaceArtifactRef  string
	Reason           domain.Reason
}

type ManualReviewQueue interface {
	// Enqueue and Complete must be idempotent by AttemptID so application
	// retries can repair a partial queue write or cleanup.
	Enqueue(context.Context, ReviewTask) error
	Complete(context.Context, string) error
}

type Keyer interface {
	Key(string) (string, error)
}

type IDSource interface {
	NewID() string
}
