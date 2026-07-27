package fieldtest

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"testing/quick"
	"time"
)

const fixtureSHA = "68bf7b18d7a2c872640265d5b6f58ba96b29561c"

var reviewTime = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

func fixturePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve fixture path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "deploy", "field-test", "examples", "staging.synthetic.blocked.json")
}

func validManifest(t *testing.T) Manifest {
	t.Helper()
	manifest, err := Load(fixturePath(t), fixtureSHA, reviewTime)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func TestSyntheticFixtureIsExplicitlyBlocked(t *testing.T) {
	manifest := validManifest(t)
	if !manifest.SyntheticOnly || manifest.Device.Physical || manifest.Disposition != "blocked" {
		t.Fatalf("unsafe synthetic fixture: %+v", manifest)
	}
}

func TestManifestFailsClosedForUnsafeVariants(t *testing.T) {
	tests := []struct {
		name   string
		sha    string
		now    time.Time
		mutate func(*Manifest)
	}{
		{"cross SHA", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", reviewTime, func(*Manifest) {}},
		{"production", fixtureSHA, reviewTime, func(m *Manifest) { m.Environment = "production" }},
		{"production data", fixtureSHA, reviewTime, func(m *Manifest) { m.Safety.ProductionData = true }},
		{"member data", fixtureSHA, reviewTime, func(m *Manifest) { m.Safety.MemberData = true }},
		{"external spend", fixtureSHA, reviewTime, func(m *Manifest) { m.Safety.ExternalSpend = true }},
		{"external traffic", fixtureSHA, reviewTime, func(m *Manifest) { m.Networks[0].ExternalTraffic = true }},
		{"wrong API", fixtureSHA, reviewTime, func(m *Manifest) { m.Device.APILevel = 36 }},
		{"wrong RAM", fixtureSHA, reviewTime, func(m *Manifest) { m.Device.RAMMiB = 4096 }},
		{"too few samples", fixtureSHA, reviewTime, func(m *Manifest) { m.Measurements[0].Samples = m.Measurements[0].Samples[:29] }},
		{"fabricated percentile", fixtureSHA, reviewTime, func(m *Manifest) { m.Measurements[0].P90-- }},
		{"mutable evidence ref", fixtureSHA, reviewTime, func(m *Manifest) { m.Measurements[0].EvidenceRef = "latest" }},
		{"same reviewer", fixtureSHA, reviewTime, func(m *Manifest) { m.ReviewerRef = m.OperatorRef }},
		{"hidden blocker", fixtureSHA, reviewTime, func(m *Manifest) { m.Blockers = []string{"synthetic-evidence"} }},
		{"false qualification", fixtureSHA, reviewTime, func(m *Manifest) { m.Disposition = "qualified-field-evidence" }},
		{"expired", fixtureSHA, time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC), func(*Manifest) {}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := validManifest(t)
			test.mutate(&manifest)
			if err := Validate(manifest, test.sha, test.now); err == nil {
				t.Fatal("unsafe field-test evidence accepted")
			}
		})
	}
}

func TestValidationDoesNotMutateSampleOrBlockerOrder(t *testing.T) {
	manifest := validManifest(t)
	manifest.Measurements[0].Samples[0], manifest.Measurements[0].Samples[29] =
		manifest.Measurements[0].Samples[29], manifest.Measurements[0].Samples[0]
	manifest.Blockers[0], manifest.Blockers[1] = manifest.Blockers[1], manifest.Blockers[0]
	beforeFirst := manifest.Measurements[0].Samples[0]
	beforeBlocker := manifest.Blockers[0]
	if err := Validate(manifest, fixtureSHA, reviewTime); err != nil {
		t.Fatal(err)
	}
	if manifest.Measurements[0].Samples[0] != beforeFirst || manifest.Blockers[0] != beforeBlocker {
		t.Fatal("validation mutated caller-owned evidence")
	}
}

func TestNearestRankPercentilesAreAlwaysObservedSamples(t *testing.T) {
	property := func(values []uint16) bool {
		if len(values) == 0 {
			return true
		}
		samples := make([]int64, len(values))
		for i, value := range values {
			samples[i] = int64(value)
		}
		slicesCopy := append([]int64(nil), samples...)
		slices.Sort(slicesCopy)
		for _, percent := range []int{50, 90, 95} {
			value := percentile(slicesCopy, percent)
			if !slices.Contains(samples, value) {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func FuzzPercentileNeverLeavesRange(f *testing.F) {
	f.Add(int64(0), int64(10), int64(20))
	f.Fuzz(func(t *testing.T, a, b, c int64) {
		values := []int64{a, b, c}
		slices.Sort(values)
		got := percentile(values, 90)
		if got < values[0] || got > values[len(values)-1] {
			t.Fatalf("percentile %d outside %v", got, values)
		}
	})
}
