package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

var (
	ErrNotFound              = errors.New("profile not found")
	ErrOptimisticConflict    = errors.New("profile revision conflict")
	ErrCommandAlreadyApplied = errors.New("profile command already applied")
	ErrRepositoryUnavailable = errors.New("profile repository unavailable")
	ErrConsentDenied         = errors.New("profile consent denied")
)

// Repository is the persistence port owned by the profile application layer.
type Repository interface {
	Find(context.Context, string) (domain.Profile, error)
	Save(context.Context, domain.Profile, uint64, string) error
}

// ConsentEvaluator receives only an opaque consent reference. The profile
// context never loads consent evidence or changes consent state.
type ConsentEvaluator interface {
	Allows(context.Context, string, string) (bool, error)
}
