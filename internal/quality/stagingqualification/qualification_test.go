package stagingqualification

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const fixtureSHA = "242f2214d29b10cdb70e046822b5146ab56db9a7"

var fixtureTime = time.Date(2026, 7, 27, 1, 30, 0, 0, time.UTC)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve fixture root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func fixtureSources(t *testing.T) Sources {
	t.Helper()
	root := repositoryRoot(t)
	return Sources{
		ReleaseEvidencePath: filepath.Join(root, "deploy", "release", "evidence", "staging.synthetic.json"),
		ReleaseBundlePath:   filepath.Join(root, "deploy", "release", "examples", "staging.synthetic.json"),
		DRRehearsalPath:     filepath.Join(root, "deploy", "atlas", "evidence", "staging.synthetic.json"),
	}
}

func qualificationPath(t *testing.T) string {
	return filepath.Join(repositoryRoot(t), "deploy", "release", "staging-qualification", "staging.synthetic.json")
}

func TestCommittedSyntheticQualificationIsExactAndDeterministic(t *testing.T) {
	sources := fixtureSources(t)
	committed, err := Load(qualificationPath(t), fixtureSHA, fixtureTime, sources)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(fixtureSHA, committed.GeneratedAt, sources)
	if err != nil {
		t.Fatal(err)
	}
	left, _ := json.Marshal(committed)
	right, _ := json.Marshal(generated)
	if string(left) != string(right) {
		t.Fatalf("committed artifact drifted\ncommitted=%s\ngenerated=%s", left, right)
	}
}

func TestQualificationFailsClosedForUnsafeVariants(t *testing.T) {
	base, err := Load(qualificationPath(t), fixtureSHA, fixtureTime, fixtureSources(t))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		sha    string
		now    time.Time
		mutate func(*Package)
	}{
		{"wrong expected SHA", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", fixtureTime, func(*Package) {}},
		{"production environment", fixtureSHA, fixtureTime, func(q *Package) { q.Environment = "production" }},
		{"not synthetic", fixtureSHA, fixtureTime, func(q *Package) { q.SyntheticOnly = false }},
		{"production evidence claim", fixtureSHA, fixtureTime, func(q *Package) { q.Limitations.ProductionEvidence = true }},
		{"legal evidence claim", fixtureSHA, fixtureTime, func(q *Package) { q.Limitations.LegalEvidence = true }},
		{"provider evidence claim", fixtureSHA, fixtureTime, func(q *Package) { q.Limitations.ProviderEvidence = true }},
		{"cohort evidence claim", fixtureSHA, fixtureTime, func(q *Package) { q.Limitations.CohortEvidence = true }},
		{"stale", fixtureSHA, base.ExpiresAt, func(*Package) {}},
		{"digest tamper", fixtureSHA, fixtureTime, func(q *Package) {
			q.Inputs.DRRehearsal.SHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := base
			test.mutate(&candidate)
			if err := Validate(candidate, test.sha, test.now, fixtureSources(t)); err == nil {
				t.Fatal("unsafe qualification accepted")
			}
		})
	}
}

func TestGenerationRejectsStaleReleaseEvidence(t *testing.T) {
	sources := fixtureSources(t)
	temp := t.TempDir()
	for _, source := range []struct {
		from string
		to   *string
		name string
	}{
		{sources.ReleaseEvidencePath, &sources.ReleaseEvidencePath, "release.json"},
		{sources.ReleaseBundlePath, &sources.ReleaseBundlePath, "bundle.json"},
		{sources.DRRehearsalPath, &sources.DRRehearsalPath, "dr.json"},
	} {
		raw, err := os.ReadFile(source.from)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(temp, source.name)
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		*source.to = path
	}
	var release ReleaseEvidence
	if err := decodeStrictFile(sources.ReleaseEvidencePath, &release); err != nil {
		t.Fatal(err)
	}
	release.GeneratedAt = fixtureTime.Add(-MaxAge - time.Minute)
	raw, _ := json.Marshal(release)
	if err := os.WriteFile(sources.ReleaseEvidencePath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(fixtureSHA, fixtureTime, sources); err == nil {
		t.Fatal("stale source evidence accepted")
	}
}
