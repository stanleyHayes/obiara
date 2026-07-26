package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/host/domain"
)

var (
	ErrNotFound              = errors.New("host application not found")
	ErrOptimisticConflict    = errors.New("host application changed")
	ErrCommandAlreadyUsed    = errors.New("host command already used")
	ErrDependencyUnavailable = errors.New("host dependency unavailable")
	ErrManualReviewRequired  = errors.New("host application requires manual review")
	ErrInstitutionRejected   = errors.New("institutional verification rejected")
)

type Repository interface {
	Create(context.Context, domain.Application) (domain.Application, bool, error)
	FindByID(context.Context, string) (domain.Application, error)
	Update(context.Context, domain.Application, uint64, string) error
	ListRecheckDue(context.Context, time.Time, int) ([]domain.Application, error)
}

type ProviderOutcome string

const (
	OutcomeVerified  ProviderOutcome = "verified"
	OutcomeRejected  ProviderOutcome = "rejected"
	OutcomeUncertain ProviderOutcome = "uncertain"
)

type ProviderRequest struct {
	CommandID       string
	ApplicationID   string
	ProofReference  string
	InstitutionKind domain.InstitutionKind
}

type ProviderResult struct {
	Outcome     ProviderOutcome
	ProviderRef string
}

type InstitutionalProvider interface {
	Verify(context.Context, ProviderRequest) (ProviderResult, error)
}

type ReviewTask struct {
	ApplicationID  string
	ProofReference string
	Reason         domain.Reason
}

type ManualReviewQueue interface {
	Enqueue(context.Context, ReviewTask) error
	Complete(context.Context, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type IDSource interface {
	NewID() string
}
