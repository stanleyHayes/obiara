package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/domain"
)

var (
	ErrCommandApplied   = errors.New("ceremony command applied")
	ErrConcurrentChange = errors.New("ceremony changed concurrently")
	ErrUnavailable      = errors.New("ceremony unavailable")
	ErrPublishDenied    = errors.New("announcement unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Ceremony) error
	Find(context.Context, string) (domain.Ceremony, error)
	FindByCommand(context.Context, string) (domain.Ceremony, error)
	Append(context.Context, domain.Ceremony, uint64, string) error
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }
type PublishRevalidator interface {
	Authorize(context.Context, [2]string, string) error
}
type CirclePublisher interface {
	Publish(context.Context, PublishRequest) error
}
type PublishRequest struct{ IdempotencyKey, DestinationKey, Kind string }
