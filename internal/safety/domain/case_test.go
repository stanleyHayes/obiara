package domain

import (
	"testing"
	"time"
)

var caseNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func newReport(tier Tier) Report {
	return ReconstituteReport("rep_1", "m-1", "m-2", CategoryFraud, tier, SurfaceRoom, "", "", StatusReceived, 1, caseNow)
}

func TestSLADeadlines(t *testing.T) {
	for tier, want := range map[Tier]time.Duration{
		TierA: 8 * time.Hour,
		TierB: 24 * time.Hour,
		TierC: 72 * time.Hour,
		TierD: 0,
	} {
		if got := SLADueAt(tier, caseNow).Sub(caseNow); got != want {
			t.Fatalf("SLA %s = %v, want %v", tier, got, want)
		}
	}
	if QueueFor(TierD) != QueueCare {
		t.Fatal("care tier must route to the care queue")
	}
	if QueueFor(TierA) != QueueTriage {
		t.Fatal("tier A routes to triage")
	}
}

func TestCaseFromReport(t *testing.T) {
	safetyCase, err := NewCaseFromReport("case_1", newReport(TierA), caseNow)
	if err != nil {
		t.Fatal(err)
	}
	if safetyCase.Queue() != QueueTriage || safetyCase.Status() != CaseQueued {
		t.Fatalf("case = %#v", safetyCase)
	}
	if safetyCase.SLADueAt() != caseNow.Add(8*time.Hour) {
		t.Fatalf("sla = %v", safetyCase.SLADueAt())
	}
	if _, err := NewCaseFromReport(" ", newReport(TierA), caseNow); err != ErrCaseIDRequired {
		t.Fatalf("missing id = %v", err)
	}
}

func TestCaseLifecycle(t *testing.T) {
	safetyCase, _ := NewCaseFromReport("case_1", newReport(TierB), caseNow)

	if err := safetyCase.Resolve("ban", "agent-1", caseNow); err != ErrCaseNotOpen {
		t.Fatalf("resolve before assign = %v", err)
	}
	if err := safetyCase.Assign(" ", caseNow); err != ErrAgentRequired {
		t.Fatalf("blank agent = %v", err)
	}
	if err := safetyCase.Assign("agent-1", caseNow); err != nil {
		t.Fatal(err)
	}
	if err := safetyCase.Resolve(" ", "agent-1", caseNow); err != ErrOutcomeRequired {
		t.Fatalf("blank outcome = %v", err)
	}
	if err := safetyCase.Resolve("warning issued", "agent-1", caseNow); err != nil {
		t.Fatal(err)
	}
	if safetyCase.Status() != CaseResolved || safetyCase.ResolvedAt() == nil {
		t.Fatalf("case = %#v", safetyCase)
	}
	if safetyCase.Breached(caseNow.Add(100 * 24 * time.Hour)) {
		t.Fatal("resolved case never breaches")
	}
}

func TestBreachClock(t *testing.T) {
	safetyCase, _ := NewCaseFromReport("case_1", newReport(TierA), caseNow)
	if safetyCase.Breached(caseNow.Add(7 * time.Hour)) {
		t.Fatal("not yet breached")
	}
	if !safetyCase.Breached(caseNow.Add(9 * time.Hour)) {
		t.Fatal("tier A breached after 8h")
	}
}
