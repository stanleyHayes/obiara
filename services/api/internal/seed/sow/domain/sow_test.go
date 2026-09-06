package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestAcceptCopiesMediaAndRejectsInvalidUnits(t *testing.T) {
	media := []Media{{Key: "media-key", ScreeningKey: "screen-key"}}
	sow, err := Accept("id", "actor", "target", "body", media, "command", "fingerprint", 1, StatusDelivered, "screen-1", sent(1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	media[0].Key = "mutated"
	if sow.Media[0].Key != "media-key" {
		t.Fatal("media aliases caller memory")
	}
	if _, err = Accept("id", "actor", "target", "body", nil, "command", "fingerprint", 0, StatusDelivered, "screen-1", sent(1), time.Now()); err == nil {
		t.Fatal("zero-unit sow accepted")
	}
}

func TestASowMustStartInAStateScreeningCanProduce(t *testing.T) {
	// Rejected is not a starting state. A refused sow is one that was held
	// and then refused, and it has to pass through pending so its seed can
	// be refunded (M4-ABUSE-01).
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	if _, err := Accept("id", "actor", "target", "body", media, "command", "fp", 1, StatusRejected, "ref", sent(1), time.Now()); err == nil {
		t.Fatal("a sow was created already rejected")
	}
	if _, err := Accept("id", "actor", "target", "body", media, "command", "fp", 1, "invented", "ref", sent(1), time.Now()); err == nil {
		t.Fatal("a sow was created in a status nothing produces")
	}
	// The screening reference is what ties the status to the judgement that
	// caused it, so a status with no reference is not a decision.
	if _, err := Accept("id", "actor", "target", "body", media, "command", "fp", 1, StatusDelivered, "  ", sent(1), time.Now()); err == nil {
		t.Fatal("a sow was created with no screening reference")
	}
}

func TestAHeldSowIsDecidedExactlyOnce(t *testing.T) {
	// Deciding twice would refund a seed twice, or deliver a sow that had
	// already been refused.
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	held, err := Accept("id", "actor", "target", "body", media, "command", "fp", 1, StatusPendingReview, "review-1", sent(1), time.Now())
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

func TestASowMustBeTowardSomebodyElse(t *testing.T) {
	// A sow with no target reaches nobody, and one aimed at yourself is not
	// a reach. Both were accepted until the sow carried a target at all,
	// which is how the sow path came to skip every rule the sprout path
	// applies to reaching a person.
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	if _, err := Accept("id", "actor", "  ", "body", media, "c", "fp", 1, StatusDelivered, "ref", sent(1), time.Now()); err == nil {
		t.Fatal("a sow was accepted toward nobody")
	}
	if _, err := Accept("id", "actor", "actor", "body", media, "c", "fp", 1, StatusDelivered, "ref", sent(1), time.Now()); err == nil {
		t.Fatal("a sow was accepted toward its own sower")
	}
}

// sent is a delivery carrying n recordings, which is what every sow here has:
// a sow is something a member says out loud.
func sent(n int) Delivery {
	refs := make([]string, 0, n)
	for i := range n {
		refs = append(refs, fmt.Sprintf("recording-%d", i))
	}
	return Delivery{SowerID: "sower", TargetID: "target", MediaRefs: refs}
}

func TestASowMustCarryTheRecordingItSaysItCarries(t *testing.T) {
	// Delivery places one pod per recording. A delivery naming a different
	// number of recordings than the sow holds would either place a pod for a
	// recording nothing screened, or lose one that was.
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	if _, err := Accept("id", "actor", "target", "body", media, "c", "fp", 1,
		StatusDelivered, "ref", sent(2), time.Now()); err == nil {
		t.Fatal("a sow was accepted carrying two recordings and holding one")
	}
}

func TestASowWithNothingToSayCannotArrive(t *testing.T) {
	// A pod is a recording resting at a house front. A sow with none has
	// nothing to place, so it is refused at the door rather than accepted,
	// charged a seed, and silently undeliverable.
	if _, err := Accept("id", "actor", "target", "body", nil, "c", "fp", 1,
		StatusDelivered, "ref", Delivery{SowerID: "sower", TargetID: "target"}, time.Now()); err == nil {
		t.Fatal("a sow with no recording was accepted")
	}
}

func TestDeliveryMustNameBothPeople(t *testing.T) {
	media := []Media{{Key: "k", ScreeningKey: "s"}}
	for name, delivery := range map[string]Delivery{
		"no sower":          {TargetID: "target", MediaRefs: []string{"r"}},
		"no target":         {SowerID: "sower", MediaRefs: []string{"r"}},
		"toward yourself":   {SowerID: "same", TargetID: "same", MediaRefs: []string{"r"}},
		"a blank recording": {SowerID: "sower", TargetID: "target", MediaRefs: []string{"  "}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Accept("id", "actor", "target", "body", media, "c", "fp", 1,
				StatusDelivered, "ref", delivery, time.Now()); err == nil {
				t.Fatal("accepted a sow that could never be delivered")
			}
		})
	}
}
