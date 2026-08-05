package releasepolicy_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func read(t *testing.T, relative string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repositoryRoot(t), relative))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestWorkflowIsManualReadOnlyAndBindsExactCandidate(t *testing.T) {
	raw := read(t, ".github/workflows/release-evidence.yml")
	var document map[string]any
	if err := yaml.Unmarshal([]byte(raw), &document); err != nil {
		t.Fatalf("parse workflow: %v", err)
	}
	if !strings.Contains(raw, "workflow_dispatch:") ||
		!strings.Contains(raw, "permissions:\n  actions: read\n  checks: read\n  contents: read") {
		t.Fatal("release evidence must be manual and repository read-only")
	}
	for _, required := range []string{
		"ref: ${{ inputs.commit_sha }}",
		"fetch-depth: 0",
		"cancel-in-progress: false",
		"retention-days: 90",
	} {
		if !strings.Contains(raw, required) {
			t.Fatalf("workflow missing %q", required)
		}
	}
	for _, forbidden := range []string{"render deploy", "curl ", "mongodb://", "mongodb+srv://"} {
		if strings.Contains(strings.ToLower(raw), forbidden) {
			t.Fatalf("workflow contains forbidden mutation or secret shape %q", forbidden)
		}
	}
}

func TestQualificationRequiresAllChecksAndBlocksProduction(t *testing.T) {
	raw := read(t, "scripts/release/qualify.sh")
	for _, required := range []string{
		`git rev-parse origin/main`,
		`Lint, test, and build`,
		`CodeQL (go)`,
		`CodeQL (javascript-typescript)`,
		`Dependency vulnerabilities`,
		`"$target" == "production"`,
		`exit 5`,
	} {
		if !strings.Contains(raw, required) {
			t.Fatalf("qualification script missing %q", required)
		}
	}
	if strings.Index(raw, `"$target" == "production"`) > strings.Index(raw, `> "$output_path"`) {
		t.Fatal("production must fail before an evidence artifact is written")
	}
}

func TestBlueprintProductionTargetIsBackendOnlyAndManual(t *testing.T) {
	raw := read(t, "render.yaml")
	for _, required := range []string{"- name: production", "obiara-api-production", "obiara-worker-production", "autoDeployTrigger: off"} {
		if !strings.Contains(raw, required) {
			t.Fatalf("production backend blueprint missing %q", required)
		}
	}
	for _, forbidden := range []string{"obiara-web-", "obiara-admin-", "runtime: node"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("Render must not host frontend resource %q", forbidden)
		}
	}
}

func TestRunbookNamesPromotionAndRollbackEvidence(t *testing.T) {
	raw := read(t, "deploy/render/promotion-runbook.md")
	for _, required := range []string{
		"exact 40-character SHA",
		"synthetic staging",
		"last known-good SHA",
		"residency/DPIA/provider/recovery/cost gates",
		"Database changes must remain backward compatible",
	} {
		if !strings.Contains(raw, required) {
			t.Fatalf("runbook missing %q", required)
		}
	}
}
