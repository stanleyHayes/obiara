package domain

import (
	"testing"
	"time"
)

var subanNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func event(kind Kind, age time.Duration) Event {
	return Event{ID: "e", SubjectID: "m-1", Kind: kind, OccurredAt: subanNow.Add(-age)}
}

func TestMarksFromFreshCredits(t *testing.T) {
	events := []Event{
		event(KindMeetingFollowThrough, time.Hour),
		event(KindMeetingFollowThrough, 24*time.Hour),
		event(KindMeetingFollowThrough, 48*time.Hour),
	}
	marks := ComputeMarks(events, subanNow)
	if len(marks) != 1 || marks[0] != MarkKeepsWord {
		t.Fatalf("marks = %v, want keeps_word", marks)
	}
}

func TestBelowThresholdNoMark(t *testing.T) {
	events := []Event{
		event(KindMeetingFollowThrough, time.Hour),
		event(KindMeetingFollowThrough, 24*time.Hour),
	}
	if marks := ComputeMarks(events, subanNow); len(marks) != 0 {
		t.Fatalf("marks = %v, want none below threshold", marks)
	}
}

func TestDecayAgesCreditsOut(t *testing.T) {
	// Three standard credits from 2.5 half-lives ago: 3 × 0.5^2.5 ≈ 0.53
	// — far below the 3.0 threshold.
	old := 913 * 24 * time.Hour
	events := []Event{
		event(KindMeetingFollowThrough, old),
		event(KindMeetingFollowThrough, old),
		event(KindMeetingFollowThrough, old),
	}
	if marks := ComputeMarks(events, subanNow); len(marks) != 0 {
		t.Fatalf("marks = %v, want decayed away", marks)
	}
}

func TestHarassmentSuppressesWithinWindow(t *testing.T) {
	events := []Event{
		event(KindMeetingFollowThrough, time.Hour),
		event(KindMeetingFollowThrough, 2*time.Hour),
		event(KindMeetingFollowThrough, 3*time.Hour),
		event(KindHarassmentFinding, 30*24*time.Hour),
	}
	if marks := ComputeMarks(events, subanNow); len(marks) != 0 {
		t.Fatalf("marks = %v, want suppressed by recent finding", marks)
	}
}

func TestOldFindingDoesNotSuppress(t *testing.T) {
	events := []Event{
		event(KindMeetingFollowThrough, time.Hour),
		event(KindMeetingFollowThrough, 2*time.Hour),
		event(KindMeetingFollowThrough, 3*time.Hour),
		event(KindHarassmentFinding, 19*30*24*time.Hour),
	}
	marks := ComputeMarks(events, subanNow)
	if len(marks) != 1 || marks[0] != MarkKeepsWord {
		t.Fatalf("marks = %v, want keeps_word after rehabilitation window", marks)
	}
}

func TestFraudSuppressesPermanently(t *testing.T) {
	events := []Event{
		event(KindMeetingFollowThrough, time.Hour),
		event(KindMeetingFollowThrough, 2*time.Hour),
		event(KindMeetingFollowThrough, 3*time.Hour),
		event(KindFraudFinding, 5*365*24*time.Hour),
	}
	if marks := ComputeMarks(events, subanNow); len(marks) != 0 {
		t.Fatalf("marks = %v, want permanently suppressed by fraud", marks)
	}
}

func TestVoucherMark(t *testing.T) {
	events := []Event{
		event(KindCleanVouch, time.Hour),
		event(KindCleanVouch, 2*time.Hour),
		event(KindCleanVouch, 3*time.Hour),
	}
	marks := ComputeMarks(events, subanNow)
	if len(marks) != 1 || marks[0] != MarkTrustedVoucher {
		t.Fatalf("marks = %v, want trusted_voucher", marks)
	}
}

func TestGraciousMarkWeights(t *testing.T) {
	// Six tiny gracious declines ≈ 1.5 — below threshold. Add three small
	// kind closures: 1.5 + 1.5 = 3.0 — threshold met.
	var events []Event
	for i := 0; i < 6; i++ {
		events = append(events, event(KindGraciousDecline, time.Duration(i)*time.Hour))
	}
	for i := 0; i < 3; i++ {
		events = append(events, event(KindKindClosure, time.Duration(i)*time.Hour))
	}
	marks := ComputeMarks(events, subanNow)
	if len(marks) != 1 || marks[0] != MarkGracious {
		t.Fatalf("marks = %v, want gracious from combined small credits", marks)
	}
}

func TestKindValidation(t *testing.T) {
	if !valid(KindMeetingFollowThrough) || valid(Kind("made_up")) {
		t.Fatal("kind validation broken")
	}
}
