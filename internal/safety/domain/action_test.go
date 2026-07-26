package domain

import (
	"testing"
	"time"
)

func TestSuspensionDurations(t *testing.T) {
	for action, want := range map[Action]time.Duration{
		ActionSuspend14d: 14 * 24 * time.Hour,
		ActionSuspend30d: 30 * 24 * time.Hour,
		ActionSuspend90d: 90 * 24 * time.Hour,
	} {
		duration, ok := SuspensionDuration(action)
		if !ok || duration != want {
			t.Fatalf("SuspensionDuration(%s) = %v, %v", action, duration, ok)
		}
	}
	if IsSuspension(ActionWarning) || IsSuspension(ActionBan) {
		t.Fatal("warning and ban are not suspensions")
	}
}

func TestLadderTierA(t *testing.T) {
	if err := CheckLadder(TierA, ActionBan, 0); err != nil {
		t.Fatalf("tier A ban = %v", err)
	}
	if err := CheckLadder(TierA, ActionSuspend30d, 0); err == nil {
		t.Fatal("tier A suspension must be rejected: immediate ban only")
	}
	if err := CheckLadder(TierA, ActionWarning, 3); err == nil {
		t.Fatal("tier A warning must be rejected")
	}
}

func TestLadderTierB(t *testing.T) {
	if err := CheckLadder(TierB, ActionSuspend14d, 0); err != nil {
		t.Fatalf("first tier-B suspension = %v", err)
	}
	if err := CheckLadder(TierB, ActionWarning, 0); err == nil {
		t.Fatal("first tier-B warning must be rejected")
	}
	if err := CheckLadder(TierB, ActionBan, 1); err != nil {
		t.Fatalf("repeat tier-B ban = %v", err)
	}
	if err := CheckLadder(TierB, ActionSuspend90d, 1); err == nil {
		t.Fatal("repeat tier-B suspension must be rejected")
	}
}

func TestLadderTierC(t *testing.T) {
	if err := CheckLadder(TierC, ActionWarning, 0); err != nil {
		t.Fatalf("first tier-C warning = %v", err)
	}
	if err := CheckLadder(TierC, ActionSuspend14d, 0); err == nil {
		t.Fatal("first tier-C suspension must be rejected")
	}
	if err := CheckLadder(TierC, ActionSuspend30d, 2); err != nil {
		t.Fatalf("repeat tier-C escalation = %v", err)
	}
	if err := CheckLadder(TierC, ActionWarning, 1); err == nil {
		t.Fatal("repeat tier-C warning must escalate")
	}
}

func TestLadderRejectsUnknownTier(t *testing.T) {
	if err := CheckLadder(Tier("Z"), ActionBan, 0); err == nil {
		t.Fatal("unknown tier must be rejected")
	}
}
