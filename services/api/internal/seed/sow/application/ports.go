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
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }

type ScreeningDecision struct {
	Approved  bool
	Reference string
}
