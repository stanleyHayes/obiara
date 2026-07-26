package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

func TestThrottleResetsOnBoundedWindowAndSignalsAfterRepeatedDenial(t *testing.T) {
	now := time.Date(2026, 7, 26, 10, 1, 0, 0, time.UTC)
	bucket, _ := New(key(1), now)
	for range 6 {
		var decision Decision
		bucket, decision, _ = bucket.Evaluate(OperationSow, bucket.Revision, now)
		if !decision.Allowed {
			t.Fatal("premature throttle")
		}
	}
	for denial := 1; denial <= 3; denial++ {
		var decision Decision
		bucket, decision, _ = bucket.Evaluate(OperationSow, bucket.Revision, now)
		if decision.Allowed || decision.CareSignal != (denial == 3) {
			t.Fatalf("denial %d: %#v", denial, decision)
		}
	}
	bucket, decision, err := bucket.Evaluate(OperationSow, bucket.Revision, now.Add(10*time.Minute))
	if err != nil || !decision.Allowed || bucket.Denials != 0 {
		t.Fatalf("reset bucket=%#v decision=%#v err=%v", bucket, decision, err)
	}
}

func TestStaleWriteFails(t *testing.T) {
	bucket, _ := New(key(1), time.Now())
	_, _, err := bucket.Evaluate(OperationSow, 9, time.Now())
	if !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("err=%v", err)
	}
}
