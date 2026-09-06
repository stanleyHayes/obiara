package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func operator() string { return strings.Repeat("b", 64) }

func requested(t *testing.T, gross int64) Payout {
	t.Helper()
	payout, err := RequestPayout("pay_1", "aff_1", gross, 20_000, 750, at)
	if err != nil {
		t.Fatal(err)
	}
	return payout
}

func TestWithholdingIsDeductedInIntegerArithmetic(t *testing.T) {
	// A tax rate multiplied by money must not be a floating-point operation,
	// and 7.5% written as 0.075 is a number no computer stores exactly.
	payout := requested(t, 100_000)
	if payout.WithheldPesewas() != 7_500 {
		t.Fatalf("withheld %d, want 7.5%% of 100000", payout.WithheldPesewas())
	}
	if payout.NetPesewas() != 92_500 {
		t.Fatalf("net %d", payout.NetPesewas())
	}
	// Gross always equals what was withheld plus what was paid. A payout
	// where those do not add up is one a filing cannot explain.
	if payout.WithheldPesewas()+payout.NetPesewas() != payout.GrossPesewas() {
		t.Fatal("the parts do not add up to the whole")
	}
}

func TestWithholdingRoundsInTheAffiliatesFavour(t *testing.T) {
	// Rounded down, so nobody is over-withheld by a decision nobody made.
	payout, err := RequestPayout("pay_1", "aff_1", 20_001, 20_000, 750, at)
	if err != nil {
		t.Fatal(err)
	}
	// 7.5% of 20001 is 1500.075
	if payout.WithheldPesewas() != 1_500 {
		t.Fatalf("withheld %d, want the fraction dropped", payout.WithheldPesewas())
	}
	if payout.WithheldPesewas()+payout.NetPesewas() != 20_001 {
		t.Fatal("rounding lost a pesewa")
	}
}

func TestNothingIsPaidUntilAWithholdingRateIsConfigured(t *testing.T) {
	// The difference between a deliberate zero and an unset config is the
	// difference between a decision and a mistake, and this is a tax
	// question.
	for _, basis := range []int64{0, -1, 10_000, 20_000} {
		if _, err := RequestPayout("pay_1", "aff_1", 100_000, 0, basis, at); !errors.Is(err, ErrWithholdingUnset) {
			t.Fatalf("basis %d gave %v, want ErrWithholdingUnset", basis, err)
		}
	}
}

func TestAPayoutBelowTheMinimumIsRefused(t *testing.T) {
	if _, err := RequestPayout("pay_1", "aff_1", 19_999, 20_000, 750, at); !errors.Is(err, ErrBelowMinimum) {
		t.Fatalf("err = %v, want ErrBelowMinimum", err)
	}
}

func TestAPayoutIsApprovedBySomebodyBeforeItMoves(t *testing.T) {
	// The first outbound money this product sends. A payout has a name
	// against it, and that name is a digest like every other actor here.
	payout := requested(t, 100_000)
	if _, err := payout.Sent("ref-1", "TRF_1"); !errors.Is(err, ErrPayoutTransition) {
		t.Fatal("a payout was sent without anybody approving it")
	}
	approved, err := payout.Approve(operator(), at)
	if err != nil {
		t.Fatal(err)
	}
	if approved.ApproverKey() != operator() || approved.DecidedAt() == nil {
		t.Fatalf("approved = %#v", approved)
	}
	sent, err := approved.Sent("ref-1", "TRF_1")
	if err != nil {
		t.Fatal(err)
	}
	if sent.Status() != PayoutSent || sent.TransferCode() != "TRF_1" {
		t.Fatalf("sent = %#v", sent)
	}
}

func TestAnUnkeyedApproverCannotApprove(t *testing.T) {
	if _, err := requested(t, 100_000).Approve("operator-1", at); !errors.Is(err, ErrInvalidPayout) {
		t.Fatal("an approval was recorded against a raw operator id")
	}
}

func TestApprovingTwiceIsRefused(t *testing.T) {
	// Two approvals would be two chances to send the same money.
	approved, err := requested(t, 100_000).Approve(operator(), at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := approved.Approve(operator(), at); !errors.Is(err, ErrPayoutTransition) {
		t.Fatalf("err = %v, want ErrPayoutTransition", err)
	}
	if _, err := approved.Refuse(operator(), at); !errors.Is(err, ErrPayoutTransition) {
		t.Fatal("an approved payout was refused afterwards")
	}
}

func TestAFailedTransferStaysFailed(t *testing.T) {
	// Whether to try again is a decision somebody makes, not a state
	// machine's to assume.
	approved, _ := requested(t, 100_000).Approve(operator(), at)
	failed, err := approved.Failed()
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status() != PayoutFailed {
		t.Fatalf("status = %q", failed.Status())
	}
	if _, err := failed.Sent("ref-1", "TRF_1"); !errors.Is(err, ErrPayoutTransition) {
		t.Fatal("a failed payout sent itself")
	}
}

func TestARefusedPayoutSendsNothing(t *testing.T) {
	refused, err := requested(t, 100_000).Refuse(operator(), at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := refused.Sent("ref-1", "TRF_1"); !errors.Is(err, ErrPayoutTransition) {
		t.Fatal("a refused payout was sent")
	}
}

func TestRehydrationRefusesAPayoutWhoseNumbersDoNotAddUp(t *testing.T) {
	// A stored payout where the parts do not sum to the whole is one a
	// filing cannot explain, and rebuilding it would carry the error forward.
	decided := time.Time{}
	_ = decided
	if _, err := RehydratePayout(
		"pay_1", "aff_1", 100_000, 750, 7_500, 90_000,
		PayoutSent, operator(), "ref-1", "TRF_1", at, nil,
	); !errors.Is(err, ErrInvalidPayout) {
		t.Fatal("a payout that does not add up was rehydrated")
	}
}
