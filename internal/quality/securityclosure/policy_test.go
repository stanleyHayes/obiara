package securityclosure

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var assessmentNow = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

func repositoryPaths(t *testing.T) (string, string) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	return filepath.Join(root, "deploy", "security", "security-policy.yaml"),
		filepath.Join(root, "deploy", "security", "penetration-evidence.synthetic.json")
}

func loadRepositoryEvidence(t *testing.T) (Policy, Evidence, Decision) {
	t.Helper()
	policyPath, evidencePath := repositoryPaths(t)
	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	evidence, decision, err := LoadEvidence(evidencePath, policy, assessmentNow)
	if err != nil {
		t.Fatal(err)
	}
	return policy, evidence, decision
}

func TestRepositorySecurityPolicyIsPinnedAndLocalOnly(t *testing.T) {
	policy, _, _ := loadRepositoryEvidence(t)
	if policy.DAST.Target != localTarget || policy.DAST.AllowExternalTargets ||
		policy.DAST.AllowProductionTargets {
		t.Fatal("DAST escaped the repository-owned localhost boundary")
	}
}

func TestSyntheticEvidenceCanNeverApproveProduction(t *testing.T) {
	_, _, decision := loadRepositoryEvidence(t)
	if decision.ProductionEligible || len(decision.Blockers) != 5 {
		t.Fatalf("synthetic decision = %#v", decision)
	}
}

func TestIndependentCompleteClosureCanApprovePolicyGate(t *testing.T) {
	policy, evidence, _ := loadRepositoryEvidence(t)
	evidence.AssessmentKind = "independent-penetration-test"
	evidence.ScopeRefs = []string{"service:api", "service:web", "service:admin", "service:mobile"}
	decision, err := Evaluate(policy, evidence, assessmentNow)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.ProductionEligible || len(decision.Blockers) != 0 {
		t.Fatalf("complete decision = %#v", decision)
	}
}

func TestUnsafeEvidenceFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy, *Evidence)
	}{
		{"external target", func(p *Policy, _ *Evidence) { p.DAST.Target = "https://production.example" }},
		{"mutable image", func(p *Policy, _ *Evidence) { p.DAST.Image = "ghcr.io/zaproxy/zaproxy:stable" }},
		{"overlong scan", func(p *Policy, _ *Evidence) { p.DAST.MaximumMinutes = 30 }},
		{"unknown assessment kind", func(_ *Policy, e *Evidence) { e.AssessmentKind = "vendor-report" }},
		{"stale evidence", func(_ *Policy, e *Evidence) { e.ExpiresAt = e.CompletedAt.AddDate(0, 0, 91) }},
		{"duplicate scope", func(_ *Policy, e *Evidence) { e.ScopeRefs = append(e.ScopeRefs, e.ScopeRefs[0]) }},
		{"invalid severity", func(_ *Policy, e *Evidence) { e.Findings[0].Severity = "urgent" }},
		{"self verification", func(_ *Policy, e *Evidence) { e.Findings[0].VerifierRef = e.Findings[0].OwnerRef }},
		{"closure without retest", func(_ *Policy, e *Evidence) { e.Findings[0].RetestRef = "" }},
		{"open finding claims closure", func(_ *Policy, e *Evidence) { e.Findings[0].Status = "open" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy, evidence, _ := loadRepositoryEvidence(t)
			test.mutate(&policy, &evidence)
			if _, err := Evaluate(policy, evidence, assessmentNow); err == nil {
				t.Fatal("unsafe evidence accepted")
			}
		})
	}
}

func TestOpenFindingBlocksIndependentAssessment(t *testing.T) {
	policy, evidence, _ := loadRepositoryEvidence(t)
	evidence.AssessmentKind = "independent-penetration-test"
	evidence.ScopeRefs = []string{"service:api", "service:web", "service:admin", "service:mobile"}
	evidence.Findings[0] = Finding{
		FindingID: "finding:open:001",
		Severity:  "high",
		Status:    "open",
		OwnerRef:  "role:platform",
	}
	decision, err := Evaluate(policy, evidence, assessmentNow)
	if err != nil {
		t.Fatal(err)
	}
	if decision.ProductionEligible || len(decision.Blockers) != 1 {
		t.Fatalf("open finding decision = %#v", decision)
	}
}
