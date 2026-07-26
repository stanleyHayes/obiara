package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/domain"
)

var (
	ErrRestrictionNotFound = errors.New("safeguarding restriction not found")
	ErrCommandConflict     = errors.New("safeguarding command conflicts with an existing command")
	ErrOptimisticConflict  = errors.New("safeguarding restriction changed")
)

// PurgeJob contains temporary lookup material and must be destroyed when the
// purge completes. It is never part of retained audit proof.
type PurgeJob struct {
	RestrictionID string
	SubjectID     string
	SourceRef     string
	PurgeDueAt    time.Time
}

type RestrictionStore interface {
	CreateBlocked(context.Context, domain.Restriction, PurgeJob) (domain.Restriction, bool, error)
	FindPending(context.Context, time.Time, int) ([]PurgeJob, error)
	FindByID(context.Context, string) (domain.Restriction, error)
	FindBySubjectKey(context.Context, string) (domain.Restriction, error)
	CompletePurge(context.Context, domain.Restriction, uint64) error
}

type ArtifactPurger interface {
	Purge(context.Context, string, string) error
}

// Keyer creates stable keyed digests. Production adapters must use HMAC with
// a rotated secret; an unkeyed digest is not sufficient for account IDs.
type Keyer interface {
	Key(string) (string, error)
}

type IDSource interface {
	NewID() string
}
