package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/domain"
)

var (
	ErrNotFound         = errors.New("cloth gate policy not found")
	ErrCommandApplied   = errors.New("cloth gate command applied")
	ErrConcurrentChange = errors.New("cloth gate changed concurrently")
	ErrUnavailable      = errors.New("cloth gate unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Policy) error
	Find(context.Context, string) (domain.Policy, error)
	FindByCommand(context.Context, string) (domain.Policy, error)
	Append(context.Context, domain.Policy, uint64, string) error
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }
