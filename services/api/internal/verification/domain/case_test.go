package domain

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var testDOB = time.Date(1998, time.April, 12, 0, 0, 0, 0, time.UTC)

func openCase(t *testing.T) VerificationCase {
	t.Helper()
	verificationCase, err := NewCase("vc_1", "id_1", "key_1", "0001", testDOB, testAssurance, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return verificationCase
}

var testAssurance = AgeAssurance{
	AssuredAt:  testNow,
	Method:     AgeSelfDeclared,
	MinimumAge: 18,
}

func TestNewCaseValidation(t *testing.T) {
	if _, err := NewCase("", "id_1", "key_1", "0001", testDOB, testAssurance, testNow); err != ErrCaseIDRequired {
		t.Fatalf("missing id = %v", err)
	}
	if _, err := NewCase("vc_1", " ", "key_1", "0001", testDOB, testAssurance, testNow); err != ErrAccountIDRequired {
		t.Fatalf("missing account = %v", err)
	}
	if _, err := NewCase("vc_1", "id_1", " ", "0001", testDOB, testAssurance, testNow); err != ErrCardKeyRequired {
		t.Fatalf("missing card key = %v", err)
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

func TestACaseCannotExistWithoutRecordingItsAgeCheck(t *testing.T) {
	// The case is the audit artifact. If it can be created without saying the
	// age was assured, the audit trail has a hole exactly where the most
	// consequential decision was made.
	for name, assurance := range map[string]AgeAssurance{
		"nothing recorded": {},
		"no method":        {AssuredAt: testNow, MinimumAge: 18},
		"no threshold":     {AssuredAt: testNow, Method: AgeSelfDeclared},
		"no timestamp":     {Method: AgeSelfDeclared, MinimumAge: 18},
	} {
		if _, err := NewCase("vc_1", "id_1", "key_1", "0001", testDOB, assurance, testNow); err != ErrAgeAssuranceRequired {
			t.Fatalf("%s: err = %v, want ErrAgeAssuranceRequired", name, err)
		}
	}
}

func TestTheAgeRecordOutlivesTheBirthDate(t *testing.T) {
	// Retention strips dateOfBirth thirty days after a decision. What is left
	// must still prove the check happened, so the record must not be derived
	// from the date it outlives.
	verificationCase, err := NewCase("vc_1", "id_1", "key_1", "0001", testDOB, testAssurance, testNow)
	if err != nil {
		t.Fatal(err)
	}
	stripped := ReconstituteCase(
		verificationCase.ID(), verificationCase.AccountID(), verificationCase.CardKey(),
		verificationCase.CardMask(), verificationCase.Status(), "", "",
		time.Time{}, verificationCase.Version(), verificationCase.CreatedAt(), nil,
	).WithAgeAssurance(verificationCase.AgeAssurance())

	if !stripped.DateOfBirth().IsZero() {
		t.Fatal("the fixture did not actually strip the date")
	}
	if !stripped.AgeAssurance().Recorded() {
		t.Fatal("stripping the birth date destroyed the proof that the age was checked")
	}
	if stripped.AgeAssurance().MinimumAge != 18 {
		t.Fatalf("threshold = %d", stripped.AgeAssurance().MinimumAge)
	}
}

func TestCorroborationOnlyUpgradesASelfDeclaredDate(t *testing.T) {
	verificationCase, err := NewCase("vc_1", "id_1", "key_1", "0001", testDOB, testAssurance, testNow)
	if err != nil {
		t.Fatal(err)
	}
	verificationCase.CorroborateAge()
	if got := verificationCase.AgeAssurance().Method; got != AgeIssuerCorroborated {
		t.Fatalf("method = %q", got)
	}
	// Corroborating twice must not invent a third, stronger claim.
	verificationCase.CorroborateAge()
	if got := verificationCase.AgeAssurance().Method; got != AgeIssuerCorroborated {
		t.Fatalf("method after a second corroboration = %q", got)
	}
}
