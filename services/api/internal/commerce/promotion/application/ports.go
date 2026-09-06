package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
)

var (
	ErrNotFound    = errors.New("promotion not found")
	ErrConflict    = errors.New("promotion changed concurrently")
	ErrCodeTaken   = errors.New("that code is already in use")
	ErrUnavailable = errors.New("promotion service unavailable")
	// ErrIssuerNotIssuing refuses a code in a suspended organization's name.
	// Codes already issued are untouched; this is only about new ones.
	ErrIssuerNotIssuing = errors.New("that organization is not issuing")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Promotion) error
	FindByCode(context.Context, string) (domain.Promotion, error)
	Append(ctx context.Context, promotion domain.Promotion, expected uint64, commandID string) error
	ListByIssuer(ctx context.Context, issuerID string, limit int) ([]domain.Promotion, error)
}

// Issuers answers whether an organization may have a code issued in its name.
// It takes the id and answers a bool, so this context never learns anything
// else about an organization.
type Issuers interface {
	Issuing(ctx context.Context, organizationID string) (bool, error)
}

// Keyer turns a member id into the digest a redemption is recorded under.
type Keyer interface {
	MemberKey(memberID string) (string, error)
}

type IDSource interface{ NewID() string }
