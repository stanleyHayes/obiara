package launchgates

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

var adversarialReviewTime = time.Date(2026, 7, 27, 8, 30, 0, 0, time.UTC)

func adversarialRegistry(t *testing.T) Registry {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve security review fixture")
	}
	path := filepath.Join(filepath.Dir(file), "../../../deploy/release/examples/production-gates.synthetic.json")
	registry, _, err := Load(path, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func independentlySatisfied(registry Registry) Registry {
	registry.Evidence = append([]Evidence(nil), registry.Evidence...)
	for index := range registry.Evidence {
		registry.Evidence[index].Synthetic = false
		registry.Evidence[index].Outcome = Satisfied
	}
	return registry
}

func evidenceIndex(t *testing.T, registry Registry, gateID string) int {
	t.Helper()
	for index := range registry.Evidence {
		if registry.Evidence[index].GateID == gateID {
			return index
		}
	}
	t.Fatalf("missing evidence for %s", gateID)
	return -1
}

func gateResult(t *testing.T, decision Decision, gateID string) GateResult {
	t.Helper()
	for _, gate := range decision.Gates {
		if gate.ID == gateID {
			return gate
		}
	}
	t.Fatalf("missing decision gate %s", gateID)
	return GateResult{}
}

func hasBlocker(blockers []string, expected string) bool {
	for _, blocker := range blockers {
		if blocker == expected {
			return true
		}
	}
	return false
}

func TestSecurityReviewRejectsEvidenceReplayAcrossDifferentGates(t *testing.T) {
	registry := independentlySatisfied(adversarialRegistry(t))
	registry.Evidence[1].EvidenceRef = registry.Evidence[0].EvidenceRef

	if _, err := Evaluate(registry, adversarialReviewTime); err == nil {
		t.Fatal("replayed evidence reference was accepted across gates")
	}
}

func TestSecurityReviewKeepsSyntheticAndStaleEvidenceFailClosed(t *testing.T) {
	synthetic := adversarialRegistry(t)
	decision, err := Evaluate(synthetic, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	for _, gate := range decision.Gates {
		if gate.Satisfied || !hasBlocker(gate.Blockers, "synthetic-evidence") {
			t.Fatalf("synthetic gate passed: %#v", gate)
		}
	}

	stale := independentlySatisfied(synthetic)
	residencyIndex := evidenceIndex(t, stale, "external.residency")
	stale.Evidence[residencyIndex].ExpiresAt = adversarialReviewTime.Add(-time.Second)
	decision, err = Evaluate(stale, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Ready ||
		!hasBlocker(gateResult(t, decision, "external.residency").Blockers, "evidence-stale") ||
		!hasBlocker(gateResult(t, decision, "external.production-topology").Blockers, "dependency-external.residency") ||
		!hasBlocker(gateResult(t, decision, "provider.atlas").Blockers, "dependency-external.residency") {
		t.Fatalf("stale external decision did not propagate: %#v", decision)
	}
}

func TestSecurityReviewRejectsWrongEnvironmentAndSelfApproval(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Evidence)
	}{
		{"staging evidence", func(evidence *Evidence) { evidence.Environment = "staging" }},
		{"self approved", func(evidence *Evidence) { evidence.ReviewerRef = evidence.IssuerRef }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := independentlySatisfied(adversarialRegistry(t))
			test.mutate(&registry.Evidence[0])
			if _, err := Evaluate(registry, adversarialReviewTime); err == nil {
				t.Fatal("unsafe evidence metadata was accepted")
			}
		})
	}
}

func TestSecurityReviewRejectsPrivateOrSecretShapedActorReferences(t *testing.T) {
	for _, unsafe := range []string{
		"role/secret-custodian",
		"role/token-owner",
		"role/private-reviewer",
		"person@example.com",
		"https://provider.invalid/decision",
		"role/2335550101",
	} {
		t.Run(unsafe, func(t *testing.T) {
			registry := independentlySatisfied(adversarialRegistry(t))
			registry.Evidence[0].IssuerRef = unsafe
			if _, err := Evaluate(registry, adversarialReviewTime); err == nil {
				t.Fatalf("unsafe actor reference %q was accepted", unsafe)
			}
		})
	}
}

func TestSecurityReviewPreventsRepositoryEvidenceFromSatisfyingExternalGate(t *testing.T) {
	registry := independentlySatisfied(adversarialRegistry(t))
	index := evidenceIndex(t, registry, "external.residency")
	registry.Evidence[index].Kind = Repository
	registry.Evidence[index].Provenance = RepositoryControl

	decision, err := Evaluate(registry, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	residency := gateResult(t, decision, "external.residency")
	if decision.Ready || residency.Satisfied ||
		!hasBlocker(residency.Blockers, "evidence-kind-mismatch") ||
		!hasBlocker(residency.Blockers, "provenance-mismatch") {
		t.Fatalf("repository evidence satisfied an external gate: %#v", residency)
	}
}

func TestSecurityReviewActivationEvidenceCannotBypassExternalDependencies(t *testing.T) {
	registry := independentlySatisfied(adversarialRegistry(t))
	residency := evidenceIndex(t, registry, "external.residency")
	registry.Evidence[residency].Outcome = Pending

	decision, err := Evaluate(registry, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	activation := gateResult(t, decision, "production.activation")
	if decision.Ready || activation.Satisfied ||
		!hasBlocker(activation.Blockers, "dependency-external.founder-decision") {
		t.Fatalf("activation metadata bypassed external decision: %#v", activation)
	}
}

func TestSecurityReviewEvidenceOrderCannotChangeDecision(t *testing.T) {
	registry := adversarialRegistry(t)
	baseline, err := Evaluate(registry, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	reordered := registry
	reordered.Evidence = append([]Evidence(nil), registry.Evidence...)
	for left, right := 0, len(reordered.Evidence)-1; left < right; left, right = left+1, right-1 {
		reordered.Evidence[left], reordered.Evidence[right] = reordered.Evidence[right], reordered.Evidence[left]
	}
	decision, err := Evaluate(reordered, adversarialReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decision, baseline) {
		t.Fatalf("evidence order changed decision:\nbaseline=%#v\nreordered=%#v", baseline, decision)
	}
}
