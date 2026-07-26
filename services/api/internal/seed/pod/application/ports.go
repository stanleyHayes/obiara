package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Pod) error
	Find(context.Context, string) (domain.Pod, error)
	FindByCommand(context.Context, string) (domain.Pod, error)
	Append(context.Context, domain.Pod, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type PlaybackEligibility interface {
	Revalidate(context.Context, string, string) error
}
type MediaIssuer interface {
	Issue(context.Context, string, string, time.Duration) (string, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
