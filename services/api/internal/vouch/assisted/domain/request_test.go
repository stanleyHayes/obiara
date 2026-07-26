package domain

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

func key(value string) string { return strings.Repeat(value, 64) }

func cmd(id, actor, reason string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: key(actor), ExpectedRevision: revision, ReasonCode: reason, At: at}
}

func newFixture(t *testing.T) (Request, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, err := NewRequest(
		"vouch-1", key("a"), key("b"), key("c"), now.Add(time.Hour),
		cmd("request-1", "b", "assisted_request", 0, now),
	)
	if err != nil {
		t.Fatal(err)
	}
	return request, now
}

func TestConsentPrecedesImmutableManualOutcome(t *testing.T) {
	request, now := newFixture(t)
	if _, err := request.Decide(DecisionApprove, cmd("early-1", "d", "operator_approved", 1, now)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("decision without consent = %v", err)
	}
	consented, err := request.Consent(cmd("consent-1", "c", "voucher_consented", 1, now.Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	approvedCommand := cmd("decision-1", "d", "identity_confirmed", 2, now.Add(2*time.Minute))
	approved, err := consented.Decide(DecisionApprove, approvedCommand)
	if err != nil {
		t.Fatal(err)
	}
	outcome := approved.Outcome()
	if outcome == nil || outcome.Decision != DecisionApprove || outcome.Provenance != "manual_assisted" ||
		outcome.OperatorKey != key("d") || !outcome.DecidedAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("outcome = %+v", outcome)
	}
	replayed, err := approved.Decide(DecisionApprove, approvedCommand)
	if err != nil || replayed.Revision() != 3 {
		t.Fatalf("replay revision=%d err=%v", replayed.Revision(), err)
	}
	if _, err := approved.Withdraw(cmd("withdraw-1", "b", "request_withdrawn", 3, now.Add(3*time.Minute))); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("decided request withdrawn: %v", err)
	}
}

func TestWithdrawalAndExpiryEndPendingRequest(t *testing.T) {
	request, now := newFixture(t)
	withdrawn, err := request.Withdraw(cmd("withdraw-1", "b", "request_withdrawn", 1, now.Add(time.Minute)))
	if err != nil || withdrawn.Status() != StatusWithdrawn {
		t.Fatalf("withdraw=%s err=%v", withdrawn.Status(), err)
	}
	if _, err := request.Consent(cmd("late-1", "c", "voucher_consented", 1, now.Add(time.Hour))); !errors.Is(err, ErrRequestExpired) {
		t.Fatalf("late consent = %v", err)
	}
	expired, err := request.Expire(cmd("expire-1", "d", "request_expired", 1, now.Add(time.Hour)))
	if err != nil || expired.Status() != StatusExpired {
		t.Fatalf("expire=%s err=%v", expired.Status(), err)
	}
}

func TestPropertyEveryAcceptedHistoryRehydrates(t *testing.T) {
	property := func(actions []byte) bool {
		if len(actions) > 60 {
			actions = actions[:60]
		}
		request, now := newFixture(t)
		for index, value := range actions {
			change := cmd(fmt.Sprintf("property-%d", index), "d", "property_action", request.Revision(), now.Add(time.Duration(index+1)*time.Second))
			var next Request
			var err error
			switch value % 5 {
			case 0:
				change.ActorKey = key("c")
				next, err = request.Consent(change)
			case 1:
				next, err = request.Decide(DecisionApprove, change)
			case 2:
				next, err = request.Decide(DecisionDecline, change)
			case 3:
				change.ActorKey = key("b")
				next, err = request.Withdraw(change)
			default:
				change.At = now.Add(time.Hour)
				next, err = request.Expire(change)
			}
			if err == nil {
				request = next
			}
			rehydrated, hydrateErr := Rehydrate(State{
				ID: request.ID(), SubjectKey: request.SubjectKey(), RequesterKey: request.RequesterKey(),
				VoucherKey: request.VoucherKey(), Status: request.Status(), ExpiresAt: request.ExpiresAt(),
				ConsentedAt: request.ConsentedAt(), Outcome: request.Outcome(), Revision: request.Revision(),
				Events: request.Events(), Commands: request.Commands(),
			})
			if hydrateErr != nil || rehydrated.Status() != request.Status() || rehydrated.Revision() != request.Revision() {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatal(err)
	}
}
