package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

var (
	ErrNotFound              = errors.New("trust edge not found")
	ErrOptimisticConflict    = errors.New("trust edge revision conflict")
	ErrCommandAlreadyApplied = errors.New("trust edge command already applied")
	ErrUnavailable           = errors.New("trust dependency unavailable")
)

type Repository interface {
	Find(context.Context, string) (domain.Edge, error)
	Save(context.Context, domain.Edge, uint64, string) error
	Outgoing(context.Context, []string) ([]domain.Edge, error)
}

type ProjectionAuthorizer interface {
	CanProject(context.Context, string, string) (bool, error)
}

type ConsentEvaluator interface {
	Allows(context.Context, string, string) (bool, error)
}
