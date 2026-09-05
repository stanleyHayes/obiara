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

// MediaIssuer mints a short-lived grant to hear one recording.
//
// It takes the listener because a grant that names nobody is a grant to
// anybody who obtains it: the media context authorizes reads per subject, and
// a token issued without one would have to be issued as somebody else.
type MediaIssuer interface {
	Issue(ctx context.Context, listenerID, mediaRef, commandID string, ttl time.Duration) (string, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
