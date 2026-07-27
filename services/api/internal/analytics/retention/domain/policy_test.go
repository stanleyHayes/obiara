package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 27, 2, 0, 0, 0, time.UTC)

func policy(t *testing.T) Policy {
	t.Helper()
	p, e := NewPolicy(PolicySpec{ID: "analytics.retention", ReviewID: "privacy.review", ReviewerKey: key(1), Version: 2, PseudonymKeyVersion: 3, ReviewedAt: now, BatchSize: 100})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func candidate(at time.Time) Candidate {
	return Candidate{ID: "507f1f77bcf86cd799439011", Name: "epono.pod_heard", SubjectRef: key(2), OccurredAt: at}
}
func TestExactAgeBoundaries(t *testing.T) {
	tests := []struct {
		age    time.Duration
		action Action
	}{{PseudonymizeAfter - time.Second, ActionKeep}, {PseudonymizeAfter, ActionPseudonymize}}
	for _, test := range tests {
		d, e := Decide(candidate(now.Add(-test.age)), policy(t), now)
		if e != nil || d.Action != test.action {
			t.Fatalf("%v %+v %v", test.age, d, e)
		}
	}
	thirteen := candidate(now.AddDate(0, -13, 0))
	d, e := Decide(thirteen, policy(t), now)
	if e != nil || d.Action != ActionAggregateErase {
		t.Fatal(d, e)
	}
}
func TestAlreadyPseudonymizedWaitsForAggregation(t *testing.T) {
	c := candidate(now.Add(-100 * 24 * time.Hour))
	c.PseudonymizedAt = now.Add(-time.Hour)
	d, e := Decide(c, policy(t), now)
	if e != nil || d.Action != ActionKeep {
		t.Fatal(d, e)
	}
}
func FuzzNeverActsBeforeNinetyDays(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(PseudonymizeAfter - time.Second))
	f.Fuzz(func(t *testing.T, nanos int64) {
		if nanos < 0 || nanos >= int64(PseudonymizeAfter) {
			t.Skip()
		}
		d, e := Decide(candidate(now.Add(-time.Duration(nanos))), policy(t), now)
		if e != nil || d.Action != ActionKeep {
			t.Fatalf("%v %+v %v", time.Duration(nanos), d, e)
		}
	})
}
