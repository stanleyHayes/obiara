package readinesshandoff

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 27, 14, 0, 0, 0, time.UTC)

func valid() Registry {
	r := Registry{SchemaVersion: SchemaVersion, GeneratedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Environment: "production", Market: "GH"}
	for i, requirement := range Requirements() {
		r.Evidence = append(r.Evidence, Evidence{
			RequirementID: requirement.ID, Kind: requirement.Kind, Provenance: requirement.Provenance,
			EvidenceRef: fmt.Sprintf("%064x", i+1), IssuerRef: "role/external-owner",
			ReviewerRef: "role/independent-reviewer", Outcome: "satisfied",
			CollectedAt: now.Add(-time.Hour), ExpiresAt: now.Add(24 * time.Hour),
		})
	}
	return r
}

func TestCompleteHandoffReady(t *testing.T) {
	d, err := Evaluate(valid(), now)
	if err != nil || !d.Ready {
		t.Fatalf("expected ready: %#v %v", d, err)
	}
}

func TestStrictSyntheticTemplateBlocked(t *testing.T) {
	_, d, err := Load("../../../deploy/release/examples/readiness-handoff.synthetic.json", now)
	if err != nil || d.Ready || len(d.Blockers) != len(Requirements()) {
		t.Fatalf("template must be valid and fully blocked: %#v %v", d, err)
	}
}

func TestLoadRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	path := t.TempDir() + "/bad.json"
	raw := `{"schemaVersion":"obiara.readiness-handoff.v1","generatedAt":"2026-07-27T13:00:00Z","expiresAt":"2026-07-27T15:00:00Z","environment":"production","market":"GH","evidence":[],"secretValue":"no"}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, now); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field must fail: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"schemaVersion":"obiara.readiness-handoff.v1","generatedAt":"2026-07-27T13:00:00Z","expiresAt":"2026-07-27T15:00:00Z","environment":"production","market":"GH","evidence":[]} {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, now); err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Fatalf("trailing data must fail: %v", err)
	}
}

func TestFailsClosed(t *testing.T) {
	tests := map[string]func(*Registry){
		"missing":   func(r *Registry) { r.Evidence = r.Evidence[1:] },
		"synthetic": func(r *Registry) { r.Evidence[0].Synthetic = true },
		"pending":   func(r *Registry) { r.Evidence[0].Outcome = "pending" },
		"stale":     func(r *Registry) { r.Evidence[0].ExpiresAt = now.Add(-time.Second) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			r := valid()
			mutate(&r)
			d, err := Evaluate(r, now)
			if err != nil || d.Ready {
				t.Fatalf("must be a blocked decision: %#v %v", d, err)
			}
		})
	}
}

func TestRejectsUnsafeMetadata(t *testing.T) {
	tests := map[string]func(*Registry){
		"unknown":                 func(r *Registry) { r.Evidence[0].RequirementID = "credential.other" },
		"duplicate requirement":   func(r *Registry) { r.Evidence = append(r.Evidence, r.Evidence[0]) },
		"replayed ref":            func(r *Registry) { r.Evidence[1].EvidenceRef = r.Evidence[0].EvidenceRef },
		"kind substitution":       func(r *Registry) { r.Evidence[0].Kind = "repository" },
		"provenance substitution": func(r *Registry) { r.Evidence[0].Provenance = "repository-control" },
		"self review":             func(r *Registry) { r.Evidence[0].ReviewerRef = r.Evidence[0].IssuerRef },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			r := valid()
			mutate(&r)
			if _, err := Evaluate(r, now); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestOrderInvariant(t *testing.T) {
	base, err := Evaluate(valid(), now)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(29))
	for i := 0; i < 1000; i++ {
		r := valid()
		random.Shuffle(len(r.Evidence), func(i, j int) { r.Evidence[i], r.Evidence[j] = r.Evidence[j], r.Evidence[i] })
		got, err := Evaluate(r, now)
		if err != nil || !reflect.DeepEqual(got, base) {
			t.Fatalf("round %d changed decision: %#v %v", i, got, err)
		}
	}
}
