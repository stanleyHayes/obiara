package releasebundle

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var reviewTime = time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

func fixturePath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "deploy", "release", "examples", "staging.synthetic.json")
}

func validBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := Load(fixturePath(t), reviewTime)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestSyntheticStagingBundleIsExplicitlyBlocked(t *testing.T) {
	bundle := validBundle(t)
	if bundle.Environment != "staging" || bundle.Disposition != "blocked" ||
		bundle.Approvals.ProductionApproved {
		t.Fatalf("unexpected synthetic disposition: %#v", bundle)
	}
}

func TestBundleFailsClosedForUnsafeVariants(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
	}{
		{"vague candidate", func(b *Bundle) { b.Candidate.CommitSHA = "latest" }},
		{"over-consented", func(b *Bundle) { b.UAT.Consented = b.UAT.Invited + 1 }},
		{"untrained completion", func(b *Bundle) { b.UAT.Completed = b.UAT.Trained + 1 }},
		{"slow rollback", func(b *Bundle) { b.Rollback.RTOMinutes = 61 }},
		{"same approver", func(b *Bundle) { b.Approvals.ReviewedBy = b.Approvals.PreparedBy }},
		{"stale", func(b *Bundle) { b.ExpiresAt = reviewTime.Add(-time.Minute) }},
		{"hidden blockers", func(b *Bundle) { b.Disposition = "qualified-non-production" }},
		{"false production", func(b *Bundle) {
			b.Environment = "production"
			b.Disposition = "production-approved"
			b.Approvals.ProductionApproved = false
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := validBundle(t)
			test.mutate(&bundle)
			if err := Validate(bundle, reviewTime); err == nil {
				t.Fatal("unsafe release bundle accepted")
			}
		})
	}
}

func TestClosedNonProductionBundleCanQualify(t *testing.T) {
	bundle := validBundle(t)
	bundle.UAT.CriticalOpen = 0
	bundle.Hypercare.Blockers = nil
	bundle.Disposition = "qualified-non-production"
	if err := Validate(bundle, reviewTime); err != nil {
		t.Fatal(err)
	}
}

func TestProductionRemainsBlockedEvenWithBundleApproval(t *testing.T) {
	bundle := validBundle(t)
	bundle.Environment = "production"
	bundle.UAT.CriticalOpen = 0
	bundle.Hypercare.Blockers = nil
	bundle.Approvals.ProductionApproved = true
	bundle.Disposition = "production-approved"
	if err := Validate(bundle, reviewTime); err == nil {
		t.Fatal("bundle alone must not approve production")
	}
}
