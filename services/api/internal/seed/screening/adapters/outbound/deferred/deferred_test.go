package deferred

import (
	"context"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

func TestTheAdvisorOffersNoOpinionWithoutClaimingAnOutage(t *testing.T) {
	// Erroring would be logged as a provider failure and read as something
	// being broken. Nothing is broken: there is no provider, and no machine
	// opinion to give.
	advisory, err := NewAdvisor().Screen(context.Background(), application.ScreeningInput{Text: "hello"})
	if err != nil {
		t.Fatalf("the advisor reported an outage: %v", err)
	}
	if advisory.Status != application.StatusUncertain {
		t.Fatalf("status = %q, want uncertain", advisory.Status)
	}
	if len(advisory.Reasons) != 1 || advisory.Reasons[0] != application.ReasonUncertain {
		t.Fatalf("reasons = %v", advisory.Reasons)
	}
}

func TestNothingHereCanApproveASow(t *testing.T) {
	// The safety property worth pinning: with these two in place the only
	// outcome available is "a person looks at it". The adjudicator never
	// claims a human decision, so even a future decisive advisor cannot
	// produce an approval through this path.
	adjudication, err := NewAdjudicator().Decide(
		context.Background(),
		application.ScreeningInput{Text: "hello"},
		application.Advisory{Status: application.StatusApproved},
	)
	if err == nil {
		t.Fatal("the adjudicator claimed a decision nobody made")
	}
	if adjudication.Status == application.StatusApproved || adjudication.HumanReviewed {
		t.Fatalf("adjudication = %#v", adjudication)
	}
}
