// Package reviewdesk coordinates the two halves of a screening decision: the
// sow that is being held, and the review record that says who decided it.
//
// It exists as its own thing rather than inside a handler because the order
// of those two writes is the whole design, and an ordering rule that lives in
// a route is one nobody can test.
package reviewdesk

import (
	"context"
	"errors"
	"strings"

	screeningdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/domain"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

var (
	// ErrUnavailable reports a desk missing one of its two halves.
	ErrUnavailable = errors.New("review desk is not composed")
	// ErrNotFound reports a reference no held sow or review answers to.
	ErrNotFound = errors.New("no review for that reference")
)

// Sows settles the held sow: delivering it, or refusing it and returning the
// seed.
type Sows interface {
	Review(ctx context.Context, screeningRef string, approve bool, decisionRef string) (sowapplication.Result, error)
}

// Reviews records who decided, and when.
type Reviews interface {
	Decide(ctx context.Context, reference string, status screeningdomain.Status, actorID, commandID string) (screeningdomain.Review, error)
}

type Desk struct {
	sows    Sows
	reviews Reviews
}

func New(sows Sows, reviews Reviews) Desk { return Desk{sows: sows, reviews: reviews} }

// Decide settles the sow and then records the judgement.
//
// The sow goes first on purpose. If the sow settles and the review record
// then fails to write, the review stays pending: a reviewer sees it again,
// tries again, and the retry finds the sow already settled and finishes the
// record. The reverse order fails much worse — a review marked decided over a
// sow still held would strand that sow forever, with the member's seed inside
// it and nothing left pointing at the problem.
func (desk Desk) Decide(ctx context.Context, reference string, approve bool, actorID, commandID string) error {
	if desk.sows == nil || desk.reviews == nil {
		return ErrUnavailable
	}
	reference = strings.TrimSpace(reference)
	if reference == "" || strings.TrimSpace(actorID) == "" || strings.TrimSpace(commandID) == "" {
		return ErrNotFound
	}

	_, err := desk.sows.Review(ctx, reference, approve, commandID)
	switch {
	case err == nil:
	case errors.Is(err, sowdomain.ErrNotPending):
		// A previous attempt settled the sow and did not finish the record.
		// Carrying on is the recovery, not a second decision.
	case errors.Is(err, sowapplication.ErrSowNotFound):
		return ErrNotFound
	default:
		return err
	}

	status := screeningdomain.StatusRefused
	if approve {
		status = screeningdomain.StatusReleased
	}
	if _, err := desk.reviews.Decide(ctx, reference, status, actorID, commandID); err != nil {
		if errors.Is(err, screeningdomain.ErrNotPending) {
			// Both halves are already done. Saying so as success is right:
			// the reviewer's decision happened, and telling them it failed
			// would invite them to make it again.
			return nil
		}
		return err
	}
	return nil
}
