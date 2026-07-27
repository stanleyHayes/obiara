package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const fixtureSHA = "68bf7b18d7a2c872640265d5b6f58ba96b29561c"

func fixturePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve fixture")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "deploy", "field-test", "examples", "staging.synthetic.blocked.json"))
}

func validArgs(t *testing.T) []string {
	return []string{
		"-manifest", fixturePath(t),
		"-candidate-sha", fixtureSHA,
		"-at", "2026-07-27T12:00:00Z",
	}
}

func TestBlockedManifestEmitsDecisionAndExitTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(validArgs(t), &stdout, &stderr)
	if code != 2 || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var got decision
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.Valid || got.Qualified || got.Disposition != "blocked" ||
		len(got.Blockers) != 2 || got.CandidateSHA != fixtureSHA {
		t.Fatalf("unsafe decision: %+v", got)
	}
}

func TestInvalidManifestEmitsInvalidDecisionAndExitOne(t *testing.T) {
	args := validArgs(t)
	args[3] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var got decision
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Valid || got.Qualified || got.Disposition != "invalid" || got.Error == "" {
		t.Fatalf("unexpected invalid decision: %+v", got)
	}
}

func TestBadUsageExits64WithoutDecision(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-manifest", fixturePath(t)}, &stdout, &stderr); code != 64 {
		t.Fatalf("code=%d", code)
	}
	if stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestOnlyQualifiedFieldEvidenceExitsZero(t *testing.T) {
	raw, err := os.ReadFile(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["syntheticOnly"] = false
	manifest["disposition"] = "qualified-field-evidence"
	manifest["blockers"] = []string{}
	manifest["device"].(map[string]any)["physical"] = true
	raw, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "qualified.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	args := validArgs(t)
	args[1] = path
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var got decision
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.Valid || !got.Qualified || got.Disposition != "qualified-field-evidence" {
		t.Fatalf("unexpected qualified decision: %+v", got)
	}
}

func TestBlockedManifestProcessExitCode(t *testing.T) {
	command := exec.Command(os.Args[0], append([]string{"-test.run=TestCLIProcessHelper", "--"}, validArgs(t)...)...)
	command.Env = append(os.Environ(), "OBIARA_FIELDTEST_HELPER=1")
	output, err := command.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 2 {
		t.Fatalf("expected process exit 2, err=%v output=%s", err, output)
	}
	var got decision
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatal(err)
	}
	if got.Disposition != "blocked" || got.Qualified {
		t.Fatalf("unexpected process decision: %+v", got)
	}
}

func TestCLIProcessHelper(t *testing.T) {
	if os.Getenv("OBIARA_FIELDTEST_HELPER") != "1" {
		return
	}
	separator := 0
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i + 1
			break
		}
	}
	os.Exit(run(os.Args[separator:], os.Stdout, os.Stderr))
}
