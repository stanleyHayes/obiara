package launchgates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

var reviewTime = time.Date(2026, 7, 27, 8, 30, 0, 0, time.UTC)

func fixturePath(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "../../../deploy/release/examples/production-gates.synthetic.json")
}
func fixture(t *testing.T) (Registry, Decision) {
	t.Helper()
	registry, decision, e := Load(fixturePath(t), reviewTime)
	if e != nil {
		t.Fatal(e)
	}
	return registry, decision
}
func TestSyntheticRegistryDistinguishesKindsAndFailsProductionClosed(t *testing.T) {
	_, decision := fixture(t)
	if decision.Ready {
		t.Fatal("synthetic production evidence passed")
	}
	if len(decision.Gates) != len(Requirements()) || len(decision.Blockers) != len(Requirements()) {
		t.Fatalf("incomplete decision %#v", decision)
	}
	kinds := map[EvidenceKind]bool{}
	for _, gate := range decision.Gates {
		kinds[gate.Kind] = true
	}
	for _, kind := range []EvidenceKind{Repository, ExternalDecision, Provider, Credential, Cohort, Store, ProductionAction} {
		if !kinds[kind] {
			t.Fatalf("missing %s", kind)
		}
	}
}
func TestCompleteIndependentCurrentEvidenceCanProduceReadinessMetadata(t *testing.T) {
	registry, _ := fixture(t)
	for i := range registry.Evidence {
		registry.Evidence[i].Synthetic = false
		registry.Evidence[i].Outcome = Satisfied
	}
	decision, e := Evaluate(registry, reviewTime)
	if e != nil {
		t.Fatal(e)
	}
	if !decision.Ready || len(decision.Blockers) != 0 {
		t.Fatalf("complete metadata blocked: %#v", decision)
	}
}
func TestRegistryFailsClosedForUnsafeVariants(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Registry)
	}{
		{"wrong environment", func(r *Registry) { r.Environment = "staging" }},
		{"wrong market", func(r *Registry) { r.Market = "US" }},
		{"stale registry", func(r *Registry) { r.ExpiresAt = reviewTime.Add(-time.Second) }},
		{"missing gate", func(r *Registry) { r.Evidence = r.Evidence[1:] }},
		{"duplicate gate", func(r *Registry) { r.Evidence = append(r.Evidence, r.Evidence[0]) }},
		{"duplicate evidence ref", func(r *Registry) { r.Evidence[1].EvidenceRef = r.Evidence[0].EvidenceRef }},
		{"wrong candidate", func(r *Registry) { r.Evidence[0].CandidateSHA = "2222222222222222222222222222222222222222" }},
		{"self review", func(r *Registry) { r.Evidence[0].ReviewerRef = r.Evidence[0].IssuerRef }},
		{"repository substitutes external", func(r *Registry) {
			for i := range r.Evidence {
				if r.Evidence[i].GateID == "external.residency" {
					r.Evidence[i].Kind = Repository
					r.Evidence[i].Provenance = RepositoryControl
				}
			}
		}},
		{"overlong evidence validity", func(r *Registry) { r.Evidence[0].ExpiresAt = r.Evidence[0].CollectedAt.Add(25 * time.Hour) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry, _ := fixture(t)
			for i := range registry.Evidence {
				registry.Evidence[i].Synthetic = false
				registry.Evidence[i].Outcome = Satisfied
			}
			test.mutate(&registry)
			decision, e := Evaluate(registry, reviewTime)
			if e == nil && decision.Ready {
				t.Fatal("unsafe registry passed")
			}
		})
	}
}
func TestEvidenceOrderDoesNotChangeDecision(t *testing.T) {
	registry, baseline := fixture(t)
	if e := quick.Check(func(seed uint64) bool {
		candidate := registry
		candidate.Evidence = append([]Evidence(nil), registry.Evidence...)
		for i := len(candidate.Evidence) - 1; i > 0; i-- {
			j := int(seed % uint64(i+1))
			candidate.Evidence[i], candidate.Evidence[j] = candidate.Evidence[j], candidate.Evidence[i]
			seed = seed*6364136223846793005 + 1
		}
		decision, e := Evaluate(candidate, reviewTime)
		return e == nil && reflect.DeepEqual(decision, baseline)
	}, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
func TestStrictFixtureRejectsPrivateOrUnknownFields(t *testing.T) {
	raw, e := os.ReadFile(fixturePath(t))
	if e != nil {
		t.Fatal(e)
	}
	var document map[string]any
	if e = json.Unmarshal(raw, &document); e != nil {
		t.Fatal(e)
	}
	document["secretValue"] = "forbidden"
	mutated, _ := json.Marshal(document)
	path := filepath.Join(t.TempDir(), "unsafe.json")
	if e = os.WriteFile(path, mutated, 0o600); e != nil {
		t.Fatal(e)
	}
	if _, _, e = Load(path, reviewTime); e == nil {
		t.Fatal("unknown private field accepted")
	}
}
func TestProductionActionIsMetadataOnly(t *testing.T) {
	typ := reflect.TypeOf(Evidence{})
	for _, forbidden := range []string{"URL", "Command", "Payload", "Credential", "Secret", "Token", "Execute", "Deploy"} {
		if _, ok := typ.FieldByName(forbidden); ok {
			t.Fatalf("evidence gained action/private field %s", forbidden)
		}
	}
}

func TestSchemaCoversCanonicalGatesAndEvidenceKinds(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../../deploy/release/launch-gates.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err = json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, requirement := range Requirements() {
		if !strings.Contains(text, `"`+requirement.ID+`"`) {
			t.Fatalf("schema missing gate %s", requirement.ID)
		}
	}
	for _, kind := range []EvidenceKind{Repository, ExternalDecision, Provider, Credential, Cohort, Store, ProductionAction} {
		if !strings.Contains(text, `"`+string(kind)+`"`) {
			t.Fatalf("schema missing kind %s", kind)
		}
	}
}
