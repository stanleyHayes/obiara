package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

var (
	ErrNotFound              = errors.New("consent record not found")
	ErrOptimisticConflict    = errors.New("consent record changed")
	ErrCommandAlreadyApplied = errors.New("consent command already applied")
	ErrRepositoryUnavailable = errors.New("consent repository unavailable")
)

type Key struct {
	SubjectID string
	PurposeID string
}

// Repository is the outbound persistence port. Save must atomically enforce
// both expectedRevision and global command-ID uniqueness for replay safety.
type Repository interface {
	Find(context.Context, Key) (domain.Record, error)
	Save(context.Context, domain.Record, uint64, string) error
}

// PurposeCatalog resolves the exact immutable purpose version requested by a
// command. Adapters must never silently substitute a newer version.
type PurposeCatalog interface {
	FindVersion(context.Context, string, uint64) (domain.Purpose, error)
}
