package domain

import (
	"errors"
	"testing"
	"time"
)

func TestAcceptCopiesMediaAndRejectsInvalidUnits(t *testing.T) {
	media := []Media{{Key: "media-key", ScreeningKey: "screen-key"}}
	sow, err := Accept("id", "actor", "body", media, "command", "fingerprint", 1, StatusDelivered, "screen-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	media[0].Key = "mutated"
	if sow.Media[0].Key != "media-key" {
		t.Fatal("media aliases caller memory")
	}
	if _, err = Accept("id", "actor", "body", nil, "command", "fingerprint", 0, StatusDelivered, "screen-1", time.Now()); err == nil {
		t.Fatal("zero-unit sow accepted")
	}
}

func TestASowMustStartInAStateScreeningCanProduce(t *testing.T) {
	// Rejected is not a starting state. A refused sow is one that was held
	// and then refused, and it has to pass through pending so its seed can
	// be refunded (M4-ABUSE-01).
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	if _, err := Accept("id", "actor", "body", media, "command", "fp", 1, StatusRejected, "ref", time.Now()); err == nil {
		t.Fatal("a sow was created already rejected")
	}
	if _, err := Accept("id", "actor", "body", media, "command", "fp", 1, "invented", "ref", time.Now()); err == nil {
		t.Fatal("a sow was created in a status nothing produces")
	}
	// The screening reference is what ties the status to the judgement that
	// caused it, so a status with no reference is not a decision.
	if _, err := Accept("id", "actor", "body", media, "command", "fp", 1, StatusDelivered, "  ", time.Now()); err == nil {
		t.Fatal("a sow was created with no screening reference")
	}
}

func TestAHeldSowIsDecidedExactlyOnce(t *testing.T) {
	// Deciding twice would refund a seed twice, or deliver a sow that had
	// already been refused.
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	held, err := Accept("id", "actor", "body", media, "command", "fp", 1, StatusPendingReview, "review-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	released, err := held.Release("review-1", time.Now())
	if err != nil || released.Status != StatusDelivered || released.DecidedAt == nil {
		t.Fatalf("release = %#v, err = %v", released, err)
	}
	if _, err := released.Release("review-1", time.Now()); !errors.Is(err, ErrNotPending) {
		t.Fatal("a delivered sow was released again")
	}

	refused, err := held.Refuse("review-1", time.Now())
	if err != nil || refused.Status != StatusRejected {
		t.Fatalf("refuse = %#v, err = %v", refused, err)
	}
	if _, err := refused.Refuse("review-1", time.Now()); !errors.Is(err, ErrNotPending) {
		t.Fatal("a rejected sow was refused again")
	}
	// A delivered sow cannot be walked back into a refusal either.
	if _, err := released.Refuse("review-1", time.Now()); !errors.Is(err, ErrNotPending) {
		t.Fatal("a delivered sow was refused")
	}
}
