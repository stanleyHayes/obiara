package domain

import (
	"testing"
	"time"
)

var careNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestNewCareCaseValidation(t *testing.T) {
	if _, err := NewCareCase("c-1", " ", SignalDistressReport, careNow); err != ErrCareSubjectNeeded {
		t.Fatalf("blank subject = %v", err)
	}
	if _, err := NewCareCase("c-1", "m-1", Signal("annoyance"), careNow); err != ErrInvalidSignal {
		t.Fatalf("unknown signal = %v", err)
	}
	careCase, err := NewCareCase("c-1", "m-1", SignalSelfHarmIndication, careNow)
	if err != nil {
		t.Fatal(err)
	}
	if careCase.Status() != CareOpen || careCase.NeedsQuietening() {
		t.Fatalf("case = %#v", careCase)
	}
}

func TestClosureNeedsQuietening(t *testing.T) {
	careCase, _ := NewCareCase("c-1", "m-1", SignalClosure, careNow)
	if !careCase.NeedsQuietening() {
		t.Fatal("closure must quieten (Doc 09 §5)")
	}
}

func TestCareLifecycle(t *testing.T) {
	careCase, _ := NewCareCase("c-1", "m-1", SignalVictimReport, careNow)
	if err := careCase.Resolve([]ScriptKey{ScriptHelplineDirectory}, careNow); err != ErrCaseNotOpen {
		t.Fatalf("resolve before engage = %v", err)
	}
	if err := careCase.Engage(); err != nil {
		t.Fatal(err)
	}
	if err := careCase.Resolve(nil, careNow); err != ErrScriptRequired {
		t.Fatalf("no scripts = %v", err)
	}
	if err := careCase.Resolve([]ScriptKey{ScriptKey("diagnose_depression")}, careNow); err != ErrInvalidScript {
		t.Fatalf("diagnostic script = %v, want rejected (resource-first only)", err)
	}
	if err := careCase.Resolve([]ScriptKey{ScriptHelplineDirectory, ScriptCounselorReferral}, careNow); err != nil {
		t.Fatal(err)
	}
	if careCase.Status() != CareResolved || len(careCase.Scripts()) != 2 {
		t.Fatalf("case = %#v", careCase)
	}
}
