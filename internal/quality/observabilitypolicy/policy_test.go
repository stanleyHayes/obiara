package observabilitypolicy

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func loadRepositoryPolicy(t *testing.T) Policy {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", "deploy", "observability", "slo.yaml")
	policy, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func TestRepositoryPolicyIsCompleteAndBounded(t *testing.T) {
	policy := loadRepositoryPolicy(t)
	if got := ErrorBudgetMinutes(policy.Objectives[0].Target, 30); got != 43 {
		t.Fatalf("99.9%% monthly error budget = %d minutes, want 43", got)
	}
	for _, objective := range policy.Objectives {
		if !objective.ReleaseBlocking {
			t.Fatalf("%s is not release blocking", objective.ID)
		}
	}
	for _, dimension := range policy.Telemetry.AllowedDimensions {
		lower := strings.ToLower(dimension)
		for _, private := range []string{"member", "session", "request.id", "correlation", "phone", "email", "content", "secret"} {
			if strings.Contains(lower, private) {
				t.Fatalf("private/high-cardinality dimension %q", dimension)
			}
		}
	}
}

func TestPolicyRejectsUnsafeAndIncompleteVariants(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy)
	}{
		{"plaintext exporter", func(p *Policy) { p.Telemetry.Transport = "http" }},
		{"forbidden dimension", func(p *Policy) { p.Telemetry.AllowedDimensions = append(p.Telemetry.AllowedDimensions, "member.id") }},
		{"unowned objective", func(p *Policy) { p.Objectives[0].Owner = "" }},
		{"nonblocking objective", func(p *Policy) { p.Objectives[0].ReleaseBlocking = false }},
		{"unknown alert objective", func(p *Policy) { p.Alerts[0].Objective = "unknown" }},
		{"invalid windows", func(p *Policy) { p.Alerts[0].ShortWindowMinutes = p.Alerts[0].LongWindowMinutes }},
		{"unowned alert", func(p *Policy) { p.Alerts[0].Owner = "" }},
		{"missing runbook", func(p *Policy) { p.Alerts[0].Runbook = "" }},
		{"vanity dashboard", func(p *Policy) { p.Dashboards[0].Panels = []string{"traffic"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := loadRepositoryPolicy(t)
			test.mutate(&policy)
			if err := Validate(policy); err == nil {
				t.Fatal("unsafe policy accepted")
			}
		})
	}
}

func TestErrorBudgetRejectsInvalidInput(t *testing.T) {
	for _, candidate := range []struct {
		target float64
		days   int
	}{{0, 30}, {1, 30}, {.999, 0}} {
		if got := ErrorBudgetMinutes(candidate.target, candidate.days); got != 0 {
			t.Fatalf("invalid budget = %d", got)
		}
	}
}
