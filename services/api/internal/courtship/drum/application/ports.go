package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/domain"
)

var (
	ErrNotFound           = errors.New("courtship drum stage not found")
	ErrOptimisticConflict = errors.New("courtship drum optimistic conflict")
	ErrCommandApplied     = errors.New("courtship drum command applied")
	ErrUnavailable        = errors.New("courtship drum unavailable")
	ErrNotAvailable       = errors.New("courtship drum not available")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Stage) error
	Find(context.Context, string) (domain.Stage, error)
	FindByCommand(context.Context, string) (domain.Stage, error)
	Append(context.Context, domain.Stage, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type Membership interface {
	RevalidatePair(context.Context, string, string) error
	RequireParticipant(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface {
	NewID() string
}
