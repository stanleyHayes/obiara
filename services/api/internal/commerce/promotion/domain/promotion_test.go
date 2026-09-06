package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	opened = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	during = time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	closed = time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
)

func memberKey(n byte) string { return strings.Repeat(string(rune('a'+n%6)), 64) }

func issued(t *testing.T, shape Shape, amount int64, redemptionCap uint32) Promotion {
	t.Helper()
	promotion, err := Issue(
		"promo_1", "ashesi2026", "org_1", "sku_membership", shape, amount,
		opened, closed, redemptionCap, Command{ID: "cmd_1", At: opened},
	)
	if err != nil {
		t.Fatal(err)
	}
	return promotion
}

func TestACodeIsUpperCasedSoItCanBeTypedOffAPoster(t *testing.T) {
	// Mixed case turns a discount into a support ticket.
	if issued(t, ShapePercentage, 50, 10).Code() != "ASHESI2026" {
		t.Fatalf("code = %q", issued(t, ShapePercentage, 50, 10).Code())
	}
	for name, code := range map[string]string{
		"too short":   "ABC",
		"punctuation": "ASHESI-2026",
		"a space":     "ASHESI 26",
		"nothing":     "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Issue("promo_1", code, "org_1", "sku", ShapeFixed, 100,
				opened, closed, 10, Command{ID: "cmd_1", At: opened}); err == nil {
				t.Fatal("a code nobody could type was issued")
			}
		})
	}
}

func TestADiscountNeverExceedsThePrice(t *testing.T) {
	// A discount larger than what is owed would make the platform owe the
	// member, and the answer to "90% off a free pass" is zero.
	fixed := issued(t, ShapeFixed, 10_000, 10)
	if off := fixed.Discount(5_000); off != 5_000 {
		t.Fatalf("a 100 cedi discount took %d off a 50 cedi pass", off)
	}
	if off := fixed.Discount(0); off != 0 {
		t.Fatalf("discounted %d off nothing", off)
	}
	if off := issued(t, ShapePercentage, 100, 10).Discount(5_000); off != 5_000 {
		t.Fatalf("a full discount took %d", off)
	}
}

func TestPercentagesRoundTowardThePlatform(t *testing.T) {
	// A member is charged whole pesewas. A fraction nobody can pay has to go
	// somewhere, and it is not the member's to absorb in their favour by
	// accident — it is a decision, so it is stated and tested.
	if off := issued(t, ShapePercentage, 33, 10).Discount(1_000); off != 330 {
		t.Fatalf("33%% of 1000 took %d", off)
	}
	if off := issued(t, ShapePercentage, 33, 10).Discount(101); off != 33 {
		t.Fatalf("33%% of 101 took %d, want 33 rather than 33.33", off)
	}
}

func TestAPercentageOverAHundredIsNotADiscount(t *testing.T) {
	if _, err := Issue("promo_1", "OVER", "org_1", "sku", ShapePercentage, 101,
		opened, closed, 10, Command{ID: "cmd_1", At: opened}); !errors.Is(err, ErrInvalidPromotion) {
		t.Fatal("a code that owed the member money was issued")
	}
}

func TestOneMemberUsesACodeOnce(t *testing.T) {
	// One per member, so a code is a welcome and not an income.
	promotion := issued(t, ShapePercentage, 50, 10)
	used, err := promotion.Redeem(memberKey(0), Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatal(err)
	}
	if used.Redeemed() != 1 {
		t.Fatalf("redeemed = %d", used.Redeemed())
	}
	if err := used.Redeemable(memberKey(0), during); !errors.Is(err, ErrAlreadyRedeemed) {
		t.Fatalf("err = %v, want ErrAlreadyRedeemed", err)
	}
	// Somebody else still may.
	if err := used.Redeemable(memberKey(1), during); err != nil {
		t.Fatalf("a second member was refused: %v", err)
	}
}

func TestTheCapIsWhatBoundsALeakedCode(t *testing.T) {
	// A code is a bearer token by decision: whoever has it may use it until
	// the cap is reached. That bounds the damage rather than preventing it.
	promotion := issued(t, ShapeFixed, 100, 2)
	for member := range 2 {
		next, err := promotion.Redeem(memberKey(byte(member)), Command{
			ID: "cmd_" + string(rune('a'+member)), At: during,
		})
		if err != nil {
			t.Fatal(err)
		}
		promotion = next
	}
	if err := promotion.Redeemable(memberKey(5), during); !errors.Is(err, ErrExhausted) {
		t.Fatalf("err = %v, want ErrExhausted", err)
	}
}

func TestACodeOutsideItsWindowIsNotRedeemable(t *testing.T) {
	promotion := issued(t, ShapeFixed, 100, 10)
	if err := promotion.Redeemable(memberKey(0), opened.Add(-time.Hour)); !errors.Is(err, ErrNotRedeemable) {
		t.Fatal("a code was used before it opened")
	}
	// The end is exclusive: a code that runs "until 1 October" does not work
	// on 1 October.
	if err := promotion.Redeemable(memberKey(0), closed); !errors.Is(err, ErrNotRedeemable) {
		t.Fatal("a code was used on the day it closed")
	}
	if err := promotion.Redeemable(memberKey(0), closed.Add(-time.Second)); err != nil {
		t.Fatalf("a code was refused a second before it closed: %v", err)
	}
}

func TestAWithdrawnCodeStopsWorking(t *testing.T) {
	withdrawn, err := issued(t, ShapeFixed, 100, 10).Withdraw(Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatal(err)
	}
	if err := withdrawn.Redeemable(memberKey(0), during); !errors.Is(err, ErrNotRedeemable) {
		t.Fatalf("err = %v, want ErrNotRedeemable", err)
	}
}

func TestARetriedRedemptionIsTheSameRedemption(t *testing.T) {
	// Retries are normal on a payment path. A second one must not spend
	// another of the cap.
	promotion := issued(t, ShapePercentage, 50, 10)
	used, err := promotion.Redeem(memberKey(0), Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := used.Redeem(memberKey(0), Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatalf("a retry was refused: %v", err)
	}
	if replayed.Redeemed() != 1 || replayed.Revision() != used.Revision() {
		t.Fatalf("a retry spent a second redemption: %d used", replayed.Redeemed())
	}
}

func TestOneCommandIdCannotMeanTwoThings(t *testing.T) {
	used, err := issued(t, ShapeFixed, 100, 10).Redeem(memberKey(0), Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := used.Withdraw(Command{ID: "cmd_2", At: during}); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("err = %v, want ErrCommandMismatch", err)
	}
}

func TestTheRowSaysHowManyAndNeverWho(t *testing.T) {
	// Counting is the whole reporting story: an organization sees how many of
	// its codes are in use and never which members used them.
	used, err := issued(t, ShapeFixed, 100, 10).Redeem(memberKey(0), Command{ID: "cmd_2", At: during})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range used.RedeemedKeys() {
		if len(key) != 64 {
			t.Fatalf("a redemption was recorded unkeyed: %q", key)
		}
	}
	// An unkeyed member cannot redeem at all, so a raw id can never get in.
	if err := used.Redeemable("member-1", during); !errors.Is(err, ErrNotRedeemable) {
		t.Fatal("a raw member id was accepted")
	}
}

func TestAWindowThatEndsBeforeItStartsIsRefused(t *testing.T) {
	if _, err := Issue("promo_1", "BACKWARD", "org_1", "sku", ShapeFixed, 100,
		closed, opened, 10, Command{ID: "cmd_1", At: opened}); !errors.Is(err, ErrInvalidPromotion) {
		t.Fatal("a code was issued for a window that never opens")
	}
}

func TestACodeWithNoCapIsRefused(t *testing.T) {
	// A cap is the only thing standing between a leaked code and every
	// membership being free, so there is no such thing as an uncapped one.
	if _, err := Issue("promo_1", "UNCAPPED", "org_1", "sku", ShapeFixed, 100,
		opened, closed, 0, Command{ID: "cmd_1", At: opened}); !errors.Is(err, ErrInvalidPromotion) {
		t.Fatal("an uncapped code was issued")
	}
}

func TestASponsorshipCoversTheWholePriceAndSaysSo(t *testing.T) {
	// A discount and a sponsorship take the same number off the price and are
	// completely different money: one is revenue the platform never earns,
	// the other is revenue it earns from somebody else's deposit.
	sponsored, err := Issue("promo_1", "ASHESISEATS", "org_1", "sku_membership",
		ShapeSponsored, 0, opened, closed, 50, Command{ID: "cmd_1", At: opened})
	if err != nil {
		t.Fatal(err)
	}
	if !sponsored.Sponsored() {
		t.Fatal("a sponsorship did not report itself as one")
	}
	if off := sponsored.Discount(5_000); off != 5_000 {
		t.Fatalf("covered %d of a 5000 pass", off)
	}
	// And an ordinary discount is not a sponsorship, however large.
	full := issued(t, ShapePercentage, 100, 10)
	if full.Sponsored() {
		t.Fatal("a hundred percent discount reported itself as sponsored")
	}
}

func TestASponsorshipCarriesNoAmount(t *testing.T) {
	// It covers the whole price, so an amount would be a number nothing
	// reads. Refusing one means a code cannot be issued that looks like it
	// means something it does not.
	if _, err := Issue("promo_1", "SEATS", "org_1", "sku", ShapeSponsored, 2_500,
		opened, closed, 50, Command{ID: "cmd_1", At: opened}); !errors.Is(err, ErrInvalidPromotion) {
		t.Fatalf("err = %v, want ErrInvalidPromotion", err)
	}
}
