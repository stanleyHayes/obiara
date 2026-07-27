package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func evidence(now time.Time) (FamilyDensity, HostCoverage, LicenseCoverage) {
	return FamilyDensity{"GH", 100, 100, 5, 5, 3, true, now},
		HostCoverage{"GH", 10, 10, 10, 2, 4, now.Add(time.Hour), true, now},
		LicenseCoverage{"GH", "gh-accra", 4, 4, 6, now.Add(time.Hour), true, now}
}
func TestReadySnapshotAndImmutability(t *testing.T) {
	now := time.Date(2026, 7, 27, 5, 0, 0, 0, time.UTC)
	f, h, l := evidence(now)
	s, e := Project(key(1), key(2), "GH", "gh-accra", "review-1", f, h, l, now)
	if e != nil {
		t.Fatal(e)
	}
	if !s.Ready() || len(s.State().Blockers) != 0 {
		t.Fatal("complete current evidence did not pass")
	}
	state := s.State()
	state.Families.ConsentedFamilies = 0
	if s.State().Families.ConsentedFamilies != 100 {
		t.Fatal("snapshot mutated through copy")
	}
}
func TestIncompleteExpiredAndWrongJurisdictionFailClosed(t *testing.T) {
	now := time.Now().UTC()
	f, h, l := evidence(now)
	f.Complete = false
	h.CertifiedUntil = now
	l.Jurisdiction = "gh-kumasi"
	s, e := Project(key(1), key(2), "GH", "gh-accra", "review-1", f, h, l, now)
	if e != nil {
		t.Fatal(e)
	}
	if s.Ready() {
		t.Fatal("unsafe evidence passed")
	}
	want := map[Blocker]bool{FamilyEvidenceIncomplete: true, HostCertificationExpired: true, JurisdictionMismatch: true}
	for _, b := range s.State().Blockers {
		delete(want, b)
	}
	if len(want) != 0 {
		t.Fatalf("missing blockers %v", want)
	}
}
func TestFamilyBoundsProperty(t *testing.T) {
	now := time.Now().UTC()
	if e := quick.Check(func(consented, target uint32) bool {
		f := FamilyDensity{"GH", int(consented), int(target), 1, 1, 1, true, now}
		want := consented <= MaxFamilies && target > 0 && target <= MaxFamilies
		return validFamilies(f, "GH") == want
	}, nil); e != nil {
		t.Fatal(e)
	}
}
func FuzzFailClosedCounts(fz *testing.F) {
	fz.Add(100, 100, 5, 5)
	fz.Fuzz(func(t *testing.T, consented, target, dense, required int) {
		now := time.Now().UTC()
		f, h, l := evidence(now)
		f.ConsentedFamilies = consented
		f.TargetFamilies = target
		f.DenseCircles = dense
		f.RequiredDenseCircles = required
		s, e := Project(key(1), key(2), "GH", "gh-accra", "review-1", f, h, l, now)
		if e == nil && s.Ready() && (consented < target || dense < required) {
			t.Fatal("shortfall passed")
		}
	})
}
