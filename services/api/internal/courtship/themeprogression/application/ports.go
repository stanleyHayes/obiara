package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/domain"
)

var (
	ErrNotFound           = errors.New("theme progression not found")
	ErrOptimisticConflict = errors.New("theme progression optimistic conflict")
	ErrCommandApplied     = errors.New("theme progression command applied")
	ErrUnavailable        = errors.New("theme progression unavailable")
	ErrNotAvailable       = errors.New("theme progression not available")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Progression) error
	Find(context.Context, string) (domain.Progression, error)
	FindByCommand(context.Context, string) (domain.Progression, error)
	Append(context.Context, domain.Progression, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type Membership interface {
	RevalidatePair(context.Context, string, string) error
	RequireParticipant(context.Context, string, string) error
}
type ThemeOneEvidence interface {
	BothRevealed(context.Context, string, string) (evidenceRef string, revealed bool, err error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
