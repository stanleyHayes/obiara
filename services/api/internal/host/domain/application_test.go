package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)

func hostApplication(t *testing.T, proofExpiry time.Time) Application {
	t.Helper()
	proof, err := NewProof(
		"evidence:1", InstitutionUniversity, strings.Repeat("a", 64),
		testNow.Add(-24*time.Hour), proofExpiry,
	)
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewApplication(
		"host:1", "submission:1", strings.Repeat("b", 64), proof,
		Command{ID: "command:submitted", ActorKey: "system", At: testNow},
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestOnlyReasonedVerificationApproves(t *testing.T) {
	value := hostApplication(t, testNow.Add(2*MaxApprovalTerm))
	approved, err := value.ProviderDecision(
		true, "provider:proof",
		Command{ID: "command:provider", ActorKey: "system", At: testNow}, 1,
	)
	if err != nil || !approved.Eligible(testNow) ||
		approved.Reason() != ReasonProviderVerified ||
		approved.ApprovedUntil() != testNow.Add(MaxApprovalTerm) {
		t.Fatalf("approved=%+v, err=%v", approved, err)
	}
	rejected, err := value.ProviderDecision(
		false, "provider:proof",
		Command{ID: "command:reject", ActorKey: "system", At: testNow}, 1,
	)
	if err != nil || rejected.Eligible(testNow) || rejected.Reason() != ReasonProviderRejected {
		t.Fatalf("rejected=%+v, err=%v", rejected, err)
	}
}

func TestUncertaintyNeverPassesAndManualReviewIsExplicit(t *testing.T) {
	value := hostApplication(t, testNow.Add(MaxApprovalTerm))
	queued, err := value.QueueManual(
		ReasonProviderUncertain, "",
		Command{ID: "command:queue", ActorKey: "system", At: testNow}, 1,
	)
	if err != nil || queued.Eligible(testNow) || queued.Status() != StatusQueuedManual {
		t.Fatalf("queued=%+v, err=%v", queued, err)
	}
	reviewer := strings.Repeat("c", 64)
	approved, err := queued.ManualDecision(
		true, Command{ID: "command:manual", ActorKey: reviewer, At: testNow}, 2,
	)
	if err != nil || !approved.Eligible(testNow) || approved.Reason() != ReasonManualApproved {
		t.Fatalf("manual approval=%+v, err=%v", approved, err)
	}
	if approved.Audit()[2].ActorKey() != reviewer {
		t.Fatal("privacy-safe reviewer key missing from audit")
	}
}

func TestProofExpiryCapsApprovalAndRequiresRecheck(t *testing.T) {
	expires := testNow.Add(60 * 24 * time.Hour)
	value := hostApplication(t, expires)
	approved, err := value.ProviderDecision(
		true, "provider:proof",
		Command{ID: "command:provider", ActorKey: "system", At: testNow}, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if approved.ApprovedUntil() != expires ||
		approved.RecheckDueAt() != expires.Add(-RecheckWindow) {
		t.Fatalf("expiry=%v recheck=%v", approved.ApprovedUntil(), approved.RecheckDueAt())
	}
	expired, err := approved.Expire(
		Command{ID: "command:expire", ActorKey: "system", At: expires}, 2,
	)
	if err != nil || expired.Eligible(expires) || expired.Status() != StatusExpired {
		t.Fatalf("expired=%+v, err=%v", expired, err)
	}
}

func TestExpiredProofAndStaleCommandsCannotApprove(t *testing.T) {
	value := hostApplication(t, testNow.Add(time.Hour))
	if _, err := value.ProviderDecision(
		true, "provider:proof",
		Command{ID: "command:late", ActorKey: "system", At: testNow.Add(2 * time.Hour)}, 1,
	); !errors.Is(err, ErrProofExpired) {
		t.Fatalf("expected proof expired, got %v", err)
	}
	if _, err := value.ProviderDecision(
		true, "provider:proof",
		Command{ID: "command:stale", ActorKey: "system", At: testNow}, 2,
	); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("expected stale version, got %v", err)
	}
}
