package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

func TestLifecycleAndExpiry(t *testing.T) {
	now := time.Date(2026, 7, 26, 6, 0, 0, 0, time.UTC)
	item, err := New(key(1), key(2), now.Add(24*time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
	for revision, state := range []State{StateDelivered, StateHeard, StateSprouted} {
		item, err = item.Transition(state, uint64(revision+1), now.Add(time.Duration(revision+1)*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = item.Transition(StateExpired, item.Revision, now.Add(48*time.Hour)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("terminal expiry error = %v", err)
	}
}

func TestSummaryHasNoUrgencyOrReadReceipt(t *testing.T) {
	now := time.Now().UTC()
	a, _ := New(key(1), key(9), now.Add(time.Hour), now)
	b, _ := New(key(2), key(9), now.Add(time.Hour), now)
	b, _ = b.Transition(StateDelivered, 1, now.Add(time.Minute))
	b, _ = b.Transition(StateHeard, 2, now.Add(2*time.Minute))
	summary, err := Summarize([]Item{a, b}, now)
	if err != nil || summary.MovingQuietly != 2 || summary.Message != "Your seeds are moving quietly." {
		t.Fatalf("summary=%#v err=%v", summary, err)
	}
}
