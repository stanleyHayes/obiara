package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

var (
	ErrUnavailable           = errors.New("sow service unavailable")
	ErrInsufficientAllowance = errors.New("insufficient weekly seed allowance")
	// ErrHumanReviewRequired is returned by a Screening implementation that
	// has routed the sow to a person. It is not a failure: the sow is held,
	// its seed is spent, and delivery waits on the review.
	ErrHumanReviewRequired = errors.New("sow requires human screening review")
	// ErrSowNotFound reports a screening reference no held sow answers to.
	ErrSowNotFound = errors.New("no sow for that screening reference")
	// ErrMediaNotOwned refuses a sow carrying a recording the sower does not
	// own. In a product where people meet through their voices, sending
	// somebody else's voice as your own is impersonation, not a mistake.
	ErrMediaNotOwned = errors.New("that recording does not belong to you")
	// ErrNotHeard refuses a sow toward somebody the member has not listened
	// to (FR-202).
	ErrNotHeard = errors.New("their voice has not been heard for long enough")
	// ErrReachNotAvailable covers a block and a decline alike. Telling them
	// apart would tell the member which one it was.
	ErrReachNotAvailable = errors.New("this reach is not available")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Screening interface {
	Screen(context.Context, string, []string) (ScreeningDecision, error)
}
type Acceptance interface {
	// Accept atomically stores the accepted sow and spends its allowance units.
	Accept(context.Context, domain.Sow) (domain.Sow, bool, error)
	// FindByScreening returns the sow a screening reference belongs to. It
	// does not filter on status: a second decision must be refused by the
	// aggregate with a reason, not disappear as "not found".
	FindByScreening(ctx context.Context, screeningRef string) (domain.Sow, error)
	// Settle stores a decided sow and, when the decision refused it, credits
	// the allowance back in the same transaction. Separating those two
	// writes would let a refusal keep a member's seed for a sow that was
	// never delivered.
	Settle(ctx context.Context, sow domain.Sow, refund bool) error
}

// ListenGate, BlockList and DeclineLock are the same three questions the
// sprout path asks before it lets anybody reach toward anybody. A sow is the
// same gesture carrying words, so it answers to the same rules — and until it
// had a target it could not even be asked them.
type ListenGate interface {
	Heard(ctx context.Context, listenerID, targetID string) (bool, error)
}

type BlockList interface {
	Blocked(ctx context.Context, memberID, otherID string) (bool, error)
}

type DeclineLock interface {
	Locked(ctx context.Context, sowerID, targetID string) (bool, error)
}

// MediaOwnership answers whether these recordings belong to the member
// sowing them. It takes the whole set rather than one reference at a time so
// an implementation can answer in a single read.
type MediaOwnership interface {
	OwnedBy(ctx context.Context, ownerID string, refs []string) (bool, error)
}

type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }

type ScreeningDecision struct {
	Approved  bool
	Reference string
}
