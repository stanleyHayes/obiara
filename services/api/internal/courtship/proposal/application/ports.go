package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
)

var (
	ErrNotAvailable     = errors.New("proposal unavailable")
	ErrCommandApplied   = errors.New("proposal command already applied")
	ErrConcurrentChange = errors.New("proposal changed concurrently")
	ErrUnavailable      = errors.New("proposal service unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Proposal) error
	Find(context.Context, string) (domain.Proposal, error)
	FindByCommand(context.Context, string) (domain.Proposal, error)
	Append(context.Context, domain.Proposal, uint64, string) error
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type DetailProtector interface {
	Protect(context.Context, string) (string, error)
}
type IDSource interface{ NewID() string }
