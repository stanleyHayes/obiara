package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/domain"
)

var (
	ErrNotAvailable     = errors.New("fire control unavailable")
	ErrCommandApplied   = errors.New("fire command applied")
	ErrConcurrentChange = errors.New("fire changed concurrently")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Fire) error
	Find(context.Context, string) (domain.Fire, error)
	FindByCommand(context.Context, string) (domain.Fire, error)
	Append(context.Context, domain.Fire, uint64, string) error
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }
type Revalidator interface {
	Authorize(context.Context, domain.State, domain.Action, string, string) error
}
type RealtimeControl interface {
	SetRole(context.Context, string, string, domain.Role, string) error
	Mute(context.Context, string, string, string) error
	EjectAndRevoke(context.Context, string, string, string) error
}
