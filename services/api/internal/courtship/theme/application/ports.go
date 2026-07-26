package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/theme/domain"
)

var (
	ErrNotFound           = errors.New("guided theme not found")
	ErrOptimisticConflict = errors.New("guided theme optimistic conflict")
	ErrCommandApplied     = errors.New("guided theme command applied")
	ErrUnavailable        = errors.New("guided theme unavailable")
	ErrNotAvailable       = errors.New("guided theme not available")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Theme) error
	Find(context.Context, string) (domain.Theme, error)
	FindByCommand(context.Context, string) (domain.Theme, error)
	Append(context.Context, domain.Theme, uint64, string) error
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
