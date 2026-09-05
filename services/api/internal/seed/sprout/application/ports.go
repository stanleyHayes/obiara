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
	// ErrNotHeard refuses a sow toward someone the member has not listened
	// to (FR-202). Reaching for a person you have not heard is the thing
	// this product exists not to be.
	ErrNotHeard = errors.New("their voice has not been heard for long enough")
	// ErrNoSeeds refuses a sow the member cannot pay for (FR-201a). The
	// allowance is what makes a sow cost something, and a sow that costs
	// nothing is a swipe.
	ErrNoSeeds = errors.New("no seed allowance left this week")
)

// Allowance spends one seed for a sow.
//
// Implementations must be idempotent by command id: a retried sow must not
// charge twice.
type Allowance interface {
	Spend(ctx context.Context, memberID, commandID string) error
}

// ListenGate answers whether the actor has heard enough of the target's Voice
// of Introduction to reach toward them.
//
// It takes members rather than an asset id on purpose: the caller must not
// choose which recording is checked, or a member could satisfy the gate with
// a recording that is not the one they are reaching toward.
type ListenGate interface {
	Heard(ctx context.Context, listenerID, targetID string) (bool, error)
}

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
