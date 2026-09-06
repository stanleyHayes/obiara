package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var at = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

func command(id, reason string) Command {
	return Command{ID: id, ActorKey: strings.Repeat("a", 64), ReasonCode: reason, At: at}
}

func opened(t *testing.T) Fund {
	t.Helper()
	fund, err := Open("fund_1", "org_1", command("cmd_1", "partner_agreement"))
	if err != nil {
		t.Fatal(err)
	}
	return fund
}

func funded(t *testing.T, pesewas int64) Fund {
	t.Helper()
	fund, err := opened(t).Deposit(pesewas, command("cmd_2", "bank_transfer_received"))
	if err != nil {
		t.Fatal(err)
	}
	return fund
}

func TestAFundHoldsNothingUntilSomebodyPays(t *testing.T) {
	fund := opened(t)
	if fund.Balance() != 0 || fund.Seats() != 0 {
		t.Fatalf("a new fund holds %d over %d seats", fund.Balance(), fund.Seats())
	}
}

func TestTheBalanceIsDepositedLessDrawn(t *testing.T) {
	// Derived rather than stored, so it cannot drift from the history an
	// organization is shown when it asks where its money went.
	fund := funded(t, 100_000)
	if fund.Balance() != 100_000 {
		t.Fatalf("balance = %d", fund.Balance())
	}
	drawn, err := fund.Draw("purchase_1", 5_000, command("cmd_3", "seat_taken"))
	if err != nil {
		t.Fatal(err)
	}
	if drawn.Balance() != 95_000 || drawn.Seats() != 1 {
		t.Fatalf("balance %d over %d seats", drawn.Balance(), drawn.Seats())
	}
}

func TestOneSeatIsDrawnOnce(t *testing.T) {
	// Drawing again would charge an organization twice for one member's seat.
	drawn, err := funded(t, 100_000).Draw("purchase_1", 5_000, command("cmd_3", "seat_taken"))
	if err != nil {
		t.Fatal(err)
	}
	again, err := drawn.Draw("purchase_1", 5_000, command("cmd_4", "seat_taken"))
	if err != nil {
		t.Fatal(err)
	}
	if again.Balance() != drawn.Balance() || again.Seats() != 1 {
		t.Fatalf("a seat was drawn twice: balance %d over %d", again.Balance(), again.Seats())
	}
}

func TestAFundCannotBeOverdrawn(t *testing.T) {
	// The whole point of a prefunded balance. Letting it go negative would be
	// the platform lending an organization money it never agreed to lend.
	if _, err := funded(t, 4_999).Draw(
		"purchase_1", 5_000, command("cmd_3", "seat_taken"),
	); !errors.Is(err, ErrInsufficient) {
		t.Fatalf("err = %v, want ErrInsufficient", err)
	}
}

func TestARefundedSeatComesBackToTheBalance(t *testing.T) {
	// The organization did not get what it paid for, so it gets the money
	// back rather than the platform keeping it.
	drawn, _ := funded(t, 100_000).Draw("purchase_1", 5_000, command("cmd_3", "seat_taken"))
	refunded, err := drawn.Refund("purchase_1", command("cmd_4", "membership_refunded"))
	if err != nil {
		t.Fatal(err)
	}
	if refunded.Balance() != 100_000 {
		t.Fatalf("balance = %d after a refund", refunded.Balance())
	}
	if refunded.Seats() != 0 {
		t.Fatalf("%d seats still counted after a refund", refunded.Seats())
	}
	// The seat is still known to have been drawn once, so it cannot be
	// refunded a second time.
	if _, err := refunded.Refund("purchase_1", command("cmd_5", "membership_refunded")); !errors.Is(err, ErrNotDrawn) {
		t.Fatalf("err = %v, want ErrNotDrawn", err)
	}
}

func TestRefundingSomethingNeverDrawnIsRefused(t *testing.T) {
	if _, err := funded(t, 100_000).Refund(
		"purchase_9", command("cmd_3", "membership_refunded"),
	); !errors.Is(err, ErrNotDrawn) {
		t.Fatalf("err = %v, want ErrNotDrawn", err)
	}
}

func TestAClosedFundIsNotDrawnFromAndKeepsWhatIsLeft(t *testing.T) {
	// Money an organization paid and did not use is still theirs. Returning
	// it is a decision somebody makes with an invoice in front of them, not
	// something a status change should silently settle.
	closed, err := funded(t, 100_000).Close(command("cmd_3", "agreement_ended"))
	if err != nil {
		t.Fatal(err)
	}
	if closed.Balance() != 100_000 {
		t.Fatalf("closing wrote off %d", 100_000-closed.Balance())
	}
	if _, err := closed.Draw("purchase_1", 5_000, command("cmd_4", "seat_taken")); !errors.Is(err, ErrFundClosed) {
		t.Fatalf("err = %v, want ErrFundClosed", err)
	}
	if _, err := closed.Deposit(1_000, command("cmd_5", "bank_transfer_received")); !errors.Is(err, ErrFundClosed) {
		t.Fatal("a closed fund took another deposit")
	}
}

func TestARetriedDepositIsCountedOnce(t *testing.T) {
	// An operator whose connection dropped will record the same transfer
	// again, and it must not double an organization's balance.
	fund := funded(t, 100_000)
	replayed, err := fund.Deposit(100_000, command("cmd_2", "bank_transfer_received"))
	if err != nil {
		t.Fatalf("a retry was refused: %v", err)
	}
	if replayed.Balance() != 100_000 || replayed.Revision() != fund.Revision() {
		t.Fatalf("a retry deposited again: %d", replayed.Balance())
	}
}

func TestOneCommandIdCannotMeanTwoAmounts(t *testing.T) {
	// Without this a retry could quietly change what an organization is
	// recorded as having paid.
	fund := funded(t, 100_000)
	if _, err := fund.Deposit(999, command("cmd_2", "bank_transfer_received")); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("err = %v, want ErrCommandMismatch", err)
	}
}

func TestEveryMovementNamesWhoAndWhy(t *testing.T) {
	// This is money somebody else paid. The trail is what an organization is
	// shown when it asks where it went.
	drawn, _ := funded(t, 100_000).Draw("purchase_1", 5_000, command("cmd_3", "seat_taken"))
	events := drawn.Events()
	if len(events) != 3 {
		t.Fatalf("%d events", len(events))
	}
	for _, event := range events {
		if len(event.ActorKey) != 64 {
			t.Fatalf("an event named a raw operator: %q", event.ActorKey)
		}
		if event.ReasonCode == "" {
			t.Fatalf("an event says what happened and not why: %#v", event)
		}
	}
	if events[2].Action != ActionDrawn || events[2].AmountPesewas != 5_000 {
		t.Fatalf("the draw was recorded as %#v", events[2])
	}
}

func TestAnUnkeyedOperatorMovesNothing(t *testing.T) {
	unkeyed := command("cmd_2", "bank_transfer_received")
	unkeyed.ActorKey = "operator-1"
	if _, err := opened(t).Deposit(100_000, unkeyed); !errors.Is(err, ErrInvalidFund) {
		t.Fatal("a deposit was recorded against a raw operator id")
	}
}

func TestAMistypedDepositIsRefused(t *testing.T) {
	// The realistic failure is not overflow — that would take nine
	// quintillion pesewas. It is an operator typing an extra six zeros, and
	// an organization's balance becoming a number nobody can explain.
	if _, err := opened(t).Deposit(
		MaxMovementPesewas+1, command("cmd_2", "bank_transfer_received"),
	); !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("err = %v, want ErrAmountOutOfRange", err)
	}
	// And a large but real deposit still works.
	if _, err := opened(t).Deposit(
		MaxMovementPesewas, command("cmd_2", "bank_transfer_received"),
	); err != nil {
		t.Fatalf("a large real deposit was refused: %v", err)
	}
}

func TestAnAbsurdDrawIsRefused(t *testing.T) {
	if _, err := funded(t, 100_000).Draw(
		"purchase_1", MaxMovementPesewas+1, command("cmd_3", "seat_taken"),
	); !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("err = %v, want ErrAmountOutOfRange", err)
	}
}
