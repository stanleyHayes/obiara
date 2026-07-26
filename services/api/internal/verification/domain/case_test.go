package domain

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var testDOB = time.Date(1998, time.April, 12, 0, 0, 0, 0, time.UTC)

func openCase(t *testing.T) VerificationCase {
	t.Helper()
	verificationCase, err := NewCase("vc_1", "id_1", "GHA-000000000-1", testDOB, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return verificationCase
}

func TestNewCaseValidation(t *testing.T) {
	if _, err := NewCase("", "id_1", "GHA-1", testDOB, testNow); err != ErrCaseIDRequired {
		t.Fatalf("missing id = %v", err)
	}
	if _, err := NewCase("vc_1", " ", "GHA-1", testDOB, testNow); err != ErrAccountIDRequired {
		t.Fatalf("missing account = %v", err)
	}
	if _, err := NewCase("vc_1", "id_1", " ", testDOB, testNow); err != ErrCardNumberRequired {
		t.Fatalf("missing card = %v", err)
	}
}

func TestCaseDecisionFlow(t *testing.T) {
	verificationCase := openCase(t)
	if err := verificationCase.Approve("ref-1", "issuer match", testNow); err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != StatusApproved || verificationCase.DecidedAt() == nil {
		t.Fatalf("case = %#v", verificationCase)
	}
	if err := verificationCase.Reject("ref-2", "again", testNow); err != ErrCaseNotOpen {
		t.Fatalf("decided case must be closed: %v", err)
	}
}

func TestManualQueueRules(t *testing.T) {
	verificationCase := openCase(t)
	if err := verificationCase.QueueForManualReview(" ", testNow); err != ErrDecisionReasonEmpty {
		t.Fatalf("blank reason = %v", err)
	}
	if err := verificationCase.QueueForManualReview("provider unavailable", testNow); err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != StatusQueuedManual {
		t.Fatalf("status = %q", verificationCase.Status())
	}
	// A queued case can still be decided by the desk.
	if err := verificationCase.Approve("manual:agent-1", "documents verified in person", testNow); err != nil {
		t.Fatal(err)
	}
	if verificationCase.ProviderRef() != "manual:agent-1" {
		t.Fatalf("providerRef = %q", verificationCase.ProviderRef())
	}
}

func TestDecisionRequiresReason(t *testing.T) {
	verificationCase := openCase(t)
	if err := verificationCase.Approve("ref-1", " ", testNow); err != ErrDecisionReasonEmpty {
		t.Fatalf("approve without reason = %v", err)
	}
	if err := verificationCase.Reject("ref-1", "", testNow); err != ErrDecisionReasonEmpty {
		t.Fatalf("reject without reason = %v", err)
	}
}
