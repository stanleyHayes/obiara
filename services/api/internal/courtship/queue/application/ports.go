package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
)

var (
	ErrNotFound         = errors.New("courtship queue room not found")
	ErrConcurrentChange = errors.New("courtship queue changed concurrently")
	ErrUnavailable      = errors.New("courtship queue unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	State(context.Context, string) (domain.State, error)
	EventByCommand(context.Context, string, string) (domain.Event, error)
	Append(context.Context, domain.State, domain.Event, uint64) (domain.Event, bool, error)
	Events(context.Context, string, uint64, int) ([]domain.Event, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
