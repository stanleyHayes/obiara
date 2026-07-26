package domain

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestNewRequestStatutoryClocks(t *testing.T) {
	export, err := NewRequest("pr_1", "id_1", KindExport, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if export.DueAt() != testNow.Add(ExportDueWithin) {
		t.Fatalf("export due = %v, want 72h", export.DueAt())
	}
	deletion, err := NewRequest("pr_2", "id_1", KindDeletion, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if deletion.DueAt() != testNow.Add(DeletionDueWithin) {
		t.Fatalf("deletion due = %v, want 30d", deletion.DueAt())
	}
	if _, err := NewRequest("pr_3", "id_1", Kind("unknown"), testNow); err == nil {
		t.Fatal("unknown kind must fail")
	}
	if _, err := NewRequest("", "id_1", KindExport, testNow); err != ErrRequestIDRequired {
		t.Fatalf("missing id = %v", err)
	}
}

func TestRequestLifecycle(t *testing.T) {
	request, _ := NewRequest("pr_1", "id_1", KindExport, testNow)

	if err := request.Complete(testNow); err != ErrRequestNotOpen {
		t.Fatalf("complete before processing = %v", err)
	}
	if err := request.StartProcessing(); err != nil {
		t.Fatal(err)
	}
	if err := request.Complete(testNow.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if request.Status() != StatusCompleted || request.CompletedAt() == nil {
		t.Fatalf("request = %#v", request)
	}
	if request.Overdue(testNow.Add(365 * 24 * time.Hour)) {
		t.Fatal("completed request is never overdue")
	}
}

func TestOverdueClock(t *testing.T) {
	request, _ := NewRequest("pr_1", "id_1", KindExport, testNow)
	if request.Overdue(testNow.Add(71 * time.Hour)) {
		t.Fatal("not yet overdue")
	}
	if !request.Overdue(testNow.Add(73 * time.Hour)) {
		t.Fatal("export overdue after 72h")
	}
}

func TestLegalHoldBlockAndUnblock(t *testing.T) {
	request, _ := NewRequest("pr_1", "id_1", KindDeletion, testNow)
	if err := request.Block(testNow); err != nil {
		t.Fatal(err)
	}
	if request.Status() != StatusBlocked {
		t.Fatalf("status = %q", request.Status())
	}
	if err := request.StartProcessing(); err != ErrRequestNotOpen {
		t.Fatalf("held request must not process: %v", err)
	}
	if err := request.Unblock(); err != nil {
		t.Fatal(err)
	}
	if err := request.StartProcessing(); err != nil {
		t.Fatal(err)
	}
}

func TestLegalHoldValidation(t *testing.T) {
	if _, err := NewLegalHold("id_1", " ", "agent-1", testNow); err != ErrHoldReasonRequired {
		t.Fatalf("blank reason = %v", err)
	}
	if _, err := NewLegalHold("id_1", "court order", " ", testNow); err != ErrHoldActorRequired {
		t.Fatalf("blank actor = %v", err)
	}
	hold, err := NewLegalHold("id_1", "court order", "agent-1", testNow)
	if err != nil || !hold.Active() {
		t.Fatalf("hold = %#v, %v", hold, err)
	}
}
