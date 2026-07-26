package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

var (
	ErrNotFound              = errors.New("circle not found")
	ErrOptimisticConflict    = errors.New("circle revision conflict")
	ErrCommandAlreadyApplied = errors.New("circle command already applied")
	ErrUnavailable           = errors.New("circle repository unavailable")
)

type Repository interface {
	Find(context.Context, string) (domain.Circle, error)
	Save(context.Context, domain.Circle, uint64, string) error
}
