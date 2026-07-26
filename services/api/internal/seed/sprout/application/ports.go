package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
)

var (
	ErrNotFound         = errors.New("sprout doorway not found")
	ErrConcurrentChange = errors.New("sprout doorway changed concurrently")
	ErrUnavailable      = errors.New("sprout service unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	RecordIntent(context.Context, domain.Intent) (*domain.Doorway, bool, error)
	FindDoorway(context.Context, string) (domain.Doorway, error)
	AppendExchange(context.Context, domain.Doorway, uint64) (domain.Doorway, bool, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }
