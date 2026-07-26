package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/domain"
	"time"
)

var (
	ErrUnavailable      = errors.New("recording unavailable")
	ErrCommandApplied   = errors.New("recording command applied")
	ErrConcurrentChange = errors.New("recording changed concurrently")
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
type Authority interface {
	Authorize(context.Context, domain.State, domain.Action, string) error
}
type Membership interface {
	Current(context.Context, string) ([]string, error)
}
type Recorder interface {
	Start(context.Context, string, domain.Purpose, time.Duration, string) (string, error)
	Stop(context.Context, string, string) error
}
