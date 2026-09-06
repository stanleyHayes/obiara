package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"time"
)

// MemberKeyNamespace is the namespace every member id is keyed under in this
// context — owners, recipients and actors alike.
//
// It is a constant rather than a literal at each call site because anything
// outside this package that needs to match a stored recipient key has to key
// the same way, and a namespace that drifts by one character does not error:
// it silently matches nobody.
const MemberKeyNamespace = "seed-pod:member"

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Pod) error
	Find(context.Context, string) (domain.Pod, error)
	FindByCommand(context.Context, string) (domain.Pod, error)
	Append(context.Context, domain.Pod, uint64, string) error
	// ForRecipient lists what is resting for one member. Without it a member
	// can only open a pod whose id they already know, and nothing tells them
	// — which is a house front with no door.
	ForRecipient(ctx context.Context, recipientKey string, at time.Time, limit int) ([]domain.Pod, error)
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
