package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var at = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

func command(id string) Command {
	return Command{ID: id, ReasonCode: "qualified_conversion", At: at}
}

func referral(n byte) string { return strings.Repeat(string(rune('a'+n%6)), 64) }

func registered(t *testing.T) Affiliate {
	t.Helper()
	affiliate, err := Register(
		"aff_1", "Campus Reps Ghana", "campus26", "reps@example.test", command("cmd_1"))
	if err != nil {
		t.Fatal(err)
	}
	return affiliate
}

func TestAnAffiliateEarnsNothingUntilSomethingConverts(t *testing.T) {
	affiliate := registered(t)
	if affiliate.Balance() != 0 || affiliate.Conversions() != 0 {
		t.Fatalf("a new affiliate is owed %d for %d", affiliate.Balance(), affiliate.Conversions())
	}
	if !affiliate.Earning() {
		t.Fatal("a new affiliate cannot earn")
	}
}

func TestOneReferralIsCountedOnce(t *testing.T) {
	// A referral converts once. Paying twice for it is the simplest way to
	// farm a scheme.
	affiliate, err := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	if err != nil {
		t.Fatal(err)
	}
	if affiliate.Balance() != 2000 || affiliate.Conversions() != 1 {
		t.Fatalf("balance %d over %d conversions", affiliate.Balance(), affiliate.Conversions())
	}
	if _, err := affiliate.Accrue(referral(0), 2000, command("cmd_3")); !errors.Is(err, ErrAlreadyCounted) {
		t.Fatalf("err = %v, want ErrAlreadyCounted", err)
	}
}

func TestAClawbackUnEarnsWhatWasEarned(t *testing.T) {
	// This is what makes the qualification rule mean something after the
	// fact rather than only when it was checked.
	earned, err := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := earned.ClawBack(referral(0), 2000, command("cmd_3"))
	if err != nil {
		t.Fatal(err)
	}
	if reversed.Balance() != 0 {
		t.Fatalf("balance = %d after a clawback", reversed.Balance())
	}
	// The referral stays counted: it converted once and was reversed once.
	// Letting it accrue again would make a clawback a way to double-count.
	if _, err := reversed.Accrue(referral(0), 2000, command("cmd_4")); !errors.Is(err, ErrAlreadyCounted) {
		t.Fatalf("err = %v, want ErrAlreadyCounted", err)
	}
}

func TestClawingBackSomethingThatNeverEarnedIsRefused(t *testing.T) {
	if _, err := registered(t).ClawBack(referral(3), 2000, command("cmd_2")); !errors.Is(err, ErrNotCounted) {
		t.Fatalf("err = %v, want ErrNotCounted", err)
	}
}

func TestABalanceCanGoNegativeWhenAPaidConversionReverses(t *testing.T) {
	// An affiliate paid for a conversion that was later reversed owes it
	// back. Clamping at zero would silently forgive it.
	earned, _ := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	paid, err := earned.RecordPayout(2000, command("cmd_3"))
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := paid.ClawBack(referral(0), 2000, command("cmd_4"))
	if err != nil {
		t.Fatal(err)
	}
	if reversed.Balance() != -2000 {
		t.Fatalf("balance = %d, want the debt to be visible", reversed.Balance())
	}
}

func TestNobodyIsPaidMoreThanTheyEarned(t *testing.T) {
	earned, _ := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	if _, err := earned.RecordPayout(2001, command("cmd_3")); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("err = %v, want ErrInsufficientBalance", err)
	}
	// And a second payout cannot re-spend the same balance.
	paid, err := earned.RecordPayout(2000, command("cmd_3"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := paid.RecordPayout(1, command("cmd_4")); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("err = %v, want ErrInsufficientBalance", err)
	}
}

func TestASuspendedAffiliateEarnsNothingButIsStillOwed(t *testing.T) {
	// Work done is owed for. Whether to pay it is a decision somebody makes,
	// not something a status change should silently settle.
	earned, _ := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	suspended, err := earned.Suspend(command("cmd_3"))
	if err != nil {
		t.Fatal(err)
	}
	if suspended.Balance() != 2000 {
		t.Fatalf("suspending wrote off %d", 2000-suspended.Balance())
	}
	if _, err := suspended.Accrue(referral(1), 2000, command("cmd_4")); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("err = %v, want ErrInvalidTransition", err)
	}
}

func TestTheStatementSaysHowManyAndNeverWho(t *testing.T) {
	// An affiliate is told how many converted. A statement naming the people
	// somebody recruited would hand an outside party a list of members.
	earned, _ := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	for _, key := range earned.CountedKeys() {
		if len(key) != 64 {
			t.Fatalf("a referral was recorded unkeyed: %q", key)
		}
	}
	// A raw member id cannot be accrued against at all.
	if _, err := earned.Accrue("member-2", 2000, command("cmd_3")); !errors.Is(err, ErrInvalidAffiliate) {
		t.Fatal("a raw member id was counted")
	}
}

func TestARetriedAccrualEarnsOnce(t *testing.T) {
	// Retries are normal on the path that decides this. A second one must not
	// pay again.
	earned, err := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := earned.Accrue(referral(0), 2000, command("cmd_2"))
	if err != nil {
		t.Fatalf("a retry was refused: %v", err)
	}
	if replayed.Balance() != 2000 || replayed.Revision() != earned.Revision() {
		t.Fatalf("a retry earned again: %d", replayed.Balance())
	}
}

func TestOneCommandIdCannotMeanTwoAmounts(t *testing.T) {
	// The fingerprint binds the id to what it did, amount included. Without
	// that a retry could quietly change what was earned.
	earned, _ := registered(t).Accrue(referral(0), 2000, command("cmd_2"))
	if _, err := earned.Accrue(referral(0), 9999, command("cmd_2")); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("err = %v, want ErrCommandMismatch", err)
	}
}

func TestAnAffiliateNeedsACodeSomebodyCanType(t *testing.T) {
	for name, code := range map[string]string{
		"too short":   "AB",
		"punctuation": "CAMPUS-26",
		"nothing":     "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Register("aff_1", "Reps", code, "reps@example.test", command("cmd_1")); err == nil {
				t.Fatal("a code nobody could type was registered")
			}
		})
	}
	// Lower case is accepted and normalised, because somebody will type it.
	affiliate, err := Register("aff_1", "Reps", "campus26", "reps@example.test", command("cmd_1"))
	if err != nil || affiliate.Code() != "CAMPUS26" {
		t.Fatalf("code = %q, err = %v", affiliate.Code(), err)
	}
}

func TestAnAffiliateNeedsSomewhereToBePaidAndReachedAt(t *testing.T) {
	if _, err := Register("aff_1", "Reps", "CAMPUS26", "", command("cmd_1")); err == nil {
		t.Fatal("an affiliate was registered with no contact at all")
	}
}
