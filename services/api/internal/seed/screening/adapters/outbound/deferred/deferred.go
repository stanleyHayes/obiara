// Package deferred supplies the screening opinions for a deployment that
// reviews every sow by hand.
//
// The owner ruled that a person decides before a sow is delivered. That makes
// the vendor question moot rather than urgent: an advisor with no opinion and
// an adjudicator that never claims a human decision send every sow to the
// queue, which is the policy stated exactly. Neither pretends to screen.
//
// The safety property is structural. `validAdjudication` in this context
// refuses any adjudication that is not marked HumanReviewed, so nothing here
// can approve a sow even if a future advisor becomes decisive — the worst
// these can do is send work to a person.
package deferred

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

// ErrNoAdjudication reports that no human decision exists for this sow yet.
// It is not a failure: it is the reason the sow is going to the queue.
var ErrNoAdjudication = errors.New("no human screening decision has been made")

// Advisor gives no opinion, deliberately.
//
// It reports uncertain rather than erroring, because an error would be logged
// as a provider failure and read as an outage. There is no provider and
// nothing is broken; there is simply no machine opinion to offer.
type Advisor struct{}

func NewAdvisor() Advisor { return Advisor{} }

func (Advisor) Screen(context.Context, application.ScreeningInput) (application.Advisory, error) {
	return application.Advisory{
		Status:     application.StatusUncertain,
		Reasons:    []application.ReasonCode{application.ReasonUncertain},
		Confidence: 0,
	}, nil
}

// Adjudicator never has a decision to report.
//
// A human decision reaches the sow through the review store and the reviewer's
// own action, not through this port, so the honest answer here is always that
// there is not one yet.
type Adjudicator struct{}

func NewAdjudicator() Adjudicator { return Adjudicator{} }

func (Adjudicator) Decide(context.Context, application.ScreeningInput, application.Advisory) (application.Adjudication, error) {
	return application.Adjudication{}, ErrNoAdjudication
}
