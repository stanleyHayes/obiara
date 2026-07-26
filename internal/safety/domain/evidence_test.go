package domain

import (
	"strings"
	"testing"
	"time"
)

var evidenceNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestAccessRecordValidation(t *testing.T) {
	if _, err := NewAccessRecord("a-1", "c-1", "agent-1", Purpose("curiosity"), evidenceNow); err != ErrInvalidPurpose {
		t.Fatalf("curiosity access = %v, want rejected", err)
	}
	if _, err := NewAccessRecord("a-1", "c-1", " ", PurposeTriage, evidenceNow); err != ErrAgentRequired {
		t.Fatalf("blank agent = %v", err)
	}
	record, err := NewAccessRecord("a-1", "c-1", "agent-1", PurposeLegal, evidenceNow)
	if err != nil || record.Purpose != PurposeLegal {
		t.Fatalf("record = %#v, %v", record, err)
	}
}

func TestRedactMasksIdentifiers(t *testing.T) {
	cases := map[string]string{
		"call her on +233 55 000 0101":    redactionMask,
		"email him at friend@example.com": redactionMask,
		"find her at @ama_serwaa":         redactionMask,
		"local number 0550000101 works":   redactionMask,
	}
	for input, want := range cases {
		if got := Redact(input); !strings.Contains(got, want) {
			t.Errorf("Redact(%q) = %q, want masked", input, got)
		}
	}
}

func TestRedactKeepsInnocentText(t *testing.T) {
	input := "He kept messaging after I said no."
	if got := Redact(input); got != input {
		t.Fatalf("innocent text mangled: %q", got)
	}
}

func TestBuildBundleRedactsDescription(t *testing.T) {
	report, err := NewReport("rep_1", "m-1", "m-2", CategoryHarassment, SurfaceRoom, "room_1", "kept pushing, said reach him on +233550000101", evidenceNow)
	if err != nil {
		t.Fatal(err)
	}
	bundle := BuildBundle(report)
	if strings.Contains(bundle.Description, "+233550000101") {
		t.Fatalf("bundle leaked phone: %q", bundle.Description)
	}
	if bundle.SubjectID != "m-2" || bundle.Tier != TierB || bundle.CaseID != "rep_1" {
		t.Fatalf("bundle = %#v", bundle)
	}
}
