package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
)

var (
	ErrNotFound       = errors.New("organization not found")
	ErrConflict       = errors.New("organization changed concurrently")
	ErrCommandApplied = errors.New("organization command already applied")
	ErrUnavailable    = errors.New("organization service unavailable")
	// ErrNameTaken refuses a second organization with a name already in use.
	// Two bodies with one name make an audit trail ambiguous about which of
	// them a code was issued for, which is the one thing this context is for.
	ErrNameTaken = errors.New("an organization with that name is already registered")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Organization) error
	FindByID(context.Context, string) (domain.Organization, error)
	FindByCommand(context.Context, string) (domain.Organization, error)
	Append(ctx context.Context, organization domain.Organization, expected uint64, commandID string) error
	List(ctx context.Context, limit int) ([]domain.Organization, error)
}

// Keyer turns an operator id into the digest the audit trail records.
type Keyer interface {
	Key(namespace, value string) (string, error)
}

type IDSource interface{ NewID() string }
