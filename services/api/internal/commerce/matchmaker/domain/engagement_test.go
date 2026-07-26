package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func fixture(t *testing.T) Engagement {
	t.Helper()
	l := License{ID: "license.gh", MatchmakerKey: key(2), Jurisdiction: "ghana", Version: 3, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(24 * time.Hour), MinimumFeePesewas: 1000, MaximumFeePesewas: 5000}
	terms := Terms{ID: "terms.1", Version: 1, TotalFeePesewas: 3000, Milestones: []Milestone{{ID: "consult", FeePesewas: 1000}, {ID: "complete", FeePesewas: 2000, DueAfter: 24 * time.Hour}}}
	e, err := Book(key(1), key(3), l, terms, "book-1", now)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func TestDualConsentBeforeExposureAndCompletedReview(t *testing.T) {
	e := fixture(t)
	e, _ = e.Curate("curate-1", key(4), now)
	if _, ok := e.ProposalRef(); ok {
		t.Fatal("proposal exposed")
	}
	if _, x := e.Expose("expose-1", now); x == nil {
		t.Fatal("single/no consent exposure")
	}
	e, _ = e.Consent(ConsentMember, "member-1", now)
	if _, x := e.Expose("expose-1", now); x == nil {
		t.Fatal("single consent exposure")
	}
	e, _ = e.Consent(ConsentCandidate, "candidate-1", now)
	e, x := e.Expose("expose-1", now)
	if x != nil {
		t.Fatal(x)
	}
	if ref, ok := e.ProposalRef(); !ok || ref != key(4) {
		t.Fatal("missing proposal")
	}
	if e.ReviewEligible() {
		t.Fatal("pre-completion review")
	}
	e, _ = e.Complete("complete-1", now)
	if !e.ReviewEligible() {
		t.Fatal("completed review unavailable")
	}
}
func TestCurrentLicenseAndFeeBand(t *testing.T) {
	l := License{ID: "license.gh", MatchmakerKey: key(2), Jurisdiction: "ghana", Version: 1, ValidFrom: now.Add(time.Hour), ValidUntil: now.Add(2 * time.Hour), MinimumFeePesewas: 1000, MaximumFeePesewas: 2000}
	terms := Terms{ID: "terms.1", Version: 1, TotalFeePesewas: 3000, Milestones: []Milestone{{ID: "one", FeePesewas: 3000}}}
	if _, e := Book(key(1), key(3), l, terms, "book-1", now); e == nil {
		t.Fatal("invalid license/fee accepted")
	}
}
func FuzzFeeMustEqualMilestones(f *testing.F) {
	f.Add(uint64(1000), uint64(2000))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		if a > 2500 || b > 2500 {
			t.Skip()
		}
		l := License{ID: "license.gh", MatchmakerKey: key(2), Jurisdiction: "ghana", Version: 1, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), MinimumFeePesewas: 1, MaximumFeePesewas: 5000}
		terms := Terms{ID: "terms.1", Version: 1, TotalFeePesewas: a + b, Milestones: []Milestone{{ID: "one", FeePesewas: a}, {ID: "two", FeePesewas: b}}}
		_, e := Book(key(1), key(3), l, terms, "book-1", now)
		valid := a > 0 && b > 0 && a+b > 0 && a+b <= 5000
		if valid != (e == nil) {
			t.Fatalf("%d %d %v", a, b, e)
		}
	})
}
