package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var reviewAt = time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

func reviewID() string { return strings.Repeat("a", 64) }

func TestARoutedReviewMustBeFindableAgain(t *testing.T) {
	// The id is the reference screening hands back, and the adapter validates
	// it as a 64-hex digest. A review routed under anything else could be
	// created and then never found — the sow would be held forever and the
	// member's seed with it.
	for name, id := range map[string]string{
		"not hex":   strings.Repeat("z", 64),
		"too short": strings.Repeat("a", 63),
		"empty":     "",
	} {
		if _, err := Route(id, "sexual_content", reviewAt); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s id was accepted", name)
		}
	}
	if _, err := Route(reviewID(), "  ", reviewAt); !errors.Is(err, ErrInvalid) {
		t.Fatal("a review was routed with no reason")
	}
	review, err := Route(reviewID(), "sexual_content", reviewAt)
	if err != nil || review.Status() != StatusPending {
		t.Fatalf("review = %#v, err = %v", review, err)
	}
}

func TestAReviewIsDecidedOnceAndBySomebody(t *testing.T) {
	// Deciding twice would release a sow that was refused, or refund its seed
	// a second time. An anonymous decision is not a review.
	review, err := Route(reviewID(), "threat", reviewAt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := review.Decide(StatusReleased, "", "cmd-1", reviewAt); !errors.Is(err, ErrInvalid) {
		t.Fatal("a review was decided by nobody")
	}
	if _, err := review.Decide(StatusReleased, "agent-1", "", reviewAt); !errors.Is(err, ErrInvalid) {
		t.Fatal("a review was decided with no request id")
	}
	if _, err := review.Decide(StatusPending, "agent-1", "cmd-1", reviewAt); !errors.Is(err, ErrInvalid) {
		t.Fatal("a review was decided as still pending")
	}

	released, err := review.Decide(StatusReleased, "agent-1", "cmd-1", reviewAt)
	if err != nil || !released.Released() || released.DecidedBy() != "agent-1" {
		t.Fatalf("released = %#v, err = %v", released, err)
	}
	if _, err := released.Decide(StatusRefused, "agent-2", "cmd-2", reviewAt); !errors.Is(err, ErrNotPending) {
		t.Fatal("a decided review was decided again")
	}
}

func TestOnlyAReleasedReviewLetsASowThrough(t *testing.T) {
	review, err := Route(reviewID(), "uncertain", reviewAt)
	if err != nil {
		t.Fatal(err)
	}
	if review.Released() {
		t.Fatal("a pending review let a sow through")
	}
	refused, err := review.Decide(StatusRefused, "agent-1", "cmd-1", reviewAt)
	if err != nil {
		t.Fatal(err)
	}
	if refused.Released() {
		t.Fatal("a refused review let a sow through")
	}
}
