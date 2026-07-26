package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSharedDeviceIsDeniedUntilExplicitApproval(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	reviewCase, created, err := NewCase("case-key", KindSharedDevice, "signal-key", "subject-key", now)
	if err != nil {
		t.Fatal(err)
	}
	if reviewCase.Allowed() || created.Sequence != 1 || created.ReasonCode != "collision_detected" {
		t.Fatalf("unsafe initial case: %+v, %+v", reviewCase, created)
	}
	audit, err := reviewCase.Resolve(ResolutionApprove, "household_confirmed", "actor-key", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !reviewCase.Allowed() || reviewCase.Version() != 2 || audit.Sequence != 2 {
		t.Fatalf("approval not deterministic: %+v, %+v", reviewCase, audit)
	}
	if _, err := reviewCase.Resolve(ResolutionDeny, "duplicate_account", "actor-key", now); !errors.Is(err, ErrCaseClosed) {
		t.Fatalf("closed case resolved twice: %v", err)
	}
}

func TestKnownNameReviewDoesNotBlockWithoutDenial(t *testing.T) {
	reviewCase, _, err := NewCase("case-key", KindKnownName, "name-key", "subject-key", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !reviewCase.Allowed() {
		t.Fatal("known-name review unexpectedly blocked access")
	}
	if _, err := reviewCase.Resolve(ResolutionDeny, "impersonation_confirmed", "actor-key", time.Now()); err != nil {
		t.Fatal(err)
	}
	if reviewCase.Allowed() {
		t.Fatal("explicit denial did not block access")
	}
}
