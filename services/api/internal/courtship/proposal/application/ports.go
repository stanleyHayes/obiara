package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	"time"
)

var (
	ErrNotAvailable     = errors.New("proposal unavailable")
	ErrCommandApplied   = errors.New("proposal command already applied")
	ErrConcurrentChange = errors.New("proposal changed concurrently")
	ErrUnavailable      = errors.New("proposal service unavailable")
)

// Summary is the read model a member sees in their own list.
//
// It is deliberately not the aggregate: rehydrating one requires its whole
// event log, and loading that per row would be a query per result for
// information a list never shows. The detail stays keyed and is not carried
// here at all.
//
//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Summary struct {
	ID        string
	Kind      domain.Type
	Status    domain.Status
	ExpiresAt time.Time
	Revision  uint64
	// Outgoing reports that this member sent the proposal, which is what
	// decides whether they may withdraw it or decide it.
	Outgoing bool
}

type Repository interface {
	Create(context.Context, domain.Proposal) error
	Find(context.Context, string) (domain.Proposal, error)
	FindByCommand(context.Context, string) (domain.Proposal, error)
	Append(context.Context, domain.Proposal, uint64, string) error
	// ListForMember returns proposals the keyed member sent or received.
	ListForMember(ctx context.Context, memberKey string, limit int) ([]Summary, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type DetailProtector interface {
	Protect(context.Context, string) (string, error)
}
type IDSource interface{ NewID() string }
