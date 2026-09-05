package reviewdesk

import (
	"context"
	"errors"
	"testing"

	screeningdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/domain"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

type sowsStub struct {
	approve  bool
	called   bool
	err      error
	sequence *[]string
}

func (s *sowsStub) Review(_ context.Context, _ string, approve bool, _ string) (sowapplication.Result, error) {
	s.called, s.approve = true, approve
	if s.sequence != nil {
		*s.sequence = append(*s.sequence, "sow")
	}
	return sowapplication.Result{}, s.err
}

type reviewsStub struct {
	status   screeningdomain.Status
	called   bool
	err      error
	sequence *[]string
}

func (r *reviewsStub) Decide(_ context.Context, _ string, status screeningdomain.Status, _, _ string) (screeningdomain.Review, error) {
	r.called, r.status = true, status
	if r.sequence != nil {
		*r.sequence = append(*r.sequence, "review")
	}
	return screeningdomain.Review{}, r.err
}

func TestTheSowIsSettledBeforeTheRecordIsWritten(t *testing.T) {
	// The ordering is the whole design. A review marked decided over a sow
	// still held would strand that sow forever with the member's seed inside
	// it and nothing left pointing at the problem.
	var order []string
	sows := &sowsStub{sequence: &order}
	reviews := &reviewsStub{sequence: &order}

	if err := New(sows, reviews).Decide(context.Background(), "review-1", true, "agent-1", "cmd-1"); err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "sow" || order[1] != "review" {
		t.Fatalf("order = %v, want the sow settled first", order)
	}
	if !sows.approve || reviews.status != screeningdomain.StatusReleased {
		t.Fatalf("approve = %v, status = %q", sows.approve, reviews.status)
	}
}

func TestARefusalRecordsARefusal(t *testing.T) {
	sows, reviews := &sowsStub{}, &reviewsStub{}
	if err := New(sows, reviews).Decide(context.Background(), "review-1", false, "agent-1", "cmd-1"); err != nil {
		t.Fatal(err)
	}
	if sows.approve || reviews.status != screeningdomain.StatusRefused {
		t.Fatalf("approve = %v, status = %q", sows.approve, reviews.status)
	}
}

func TestAHalfFinishedDecisionIsFinishedByARetry(t *testing.T) {
	// The recovery path the ordering buys. A previous attempt settled the sow
	// and did not write the record; the retry must carry on and finish it
	// rather than refuse because the sow is no longer pending.
	sows := &sowsStub{err: sowdomain.ErrNotPending}
	reviews := &reviewsStub{}

	if err := New(sows, reviews).Decide(context.Background(), "review-1", true, "agent-1", "cmd-1"); err != nil {
		t.Fatalf("a retry over a settled sow failed: %v", err)
	}
	if !reviews.called {
		t.Fatal("the retry did not finish the record")
	}
}

func TestADecisionAlreadyRecordedIsSuccessNotFailure(t *testing.T) {
	// Both halves done. Telling the reviewer it failed would invite them to
	// make the decision again.
	sows := &sowsStub{err: sowdomain.ErrNotPending}
	reviews := &reviewsStub{err: screeningdomain.ErrNotPending}

	if err := New(sows, reviews).Decide(context.Background(), "review-1", true, "agent-1", "cmd-1"); err != nil {
		t.Fatalf("a fully settled decision reported %v", err)
	}
}

func TestAFailedSowSettlementNeverWritesTheRecord(t *testing.T) {
	// If the sow could not be settled, recording that somebody decided it
	// would be a lie, and would hide a sow that is still held.
	sows := &sowsStub{err: errors.New("mongo unavailable")}
	reviews := &reviewsStub{}

	if err := New(sows, reviews).Decide(context.Background(), "review-1", true, "agent-1", "cmd-1"); err == nil {
		t.Fatal("a failed settlement reported success")
	}
	if reviews.called {
		t.Fatal("the record was written over a sow that was never settled")
	}
}

func TestAnUnknownReferenceIsNotFound(t *testing.T) {
	sows := &sowsStub{err: sowapplication.ErrSowNotFound}
	if err := New(sows, &reviewsStub{}).Decide(context.Background(), "nope", true, "agent-1", "cmd-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	// A decision with no actor or no request id is not a decision.
	desk := New(&sowsStub{}, &reviewsStub{})
	for name, args := range map[string][2]string{
		"no actor":      {"", "cmd-1"},
		"no request id": {"agent-1", ""},
	} {
		if err := desk.Decide(context.Background(), "review-1", true, args[0], args[1]); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}
