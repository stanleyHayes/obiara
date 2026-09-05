package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var testTime = time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)

func command(id string, at time.Time) Command {
	return Command{ID: id, Fingerprint: strings.Repeat("a", 64), At: at}
}

func draft(t *testing.T, retention Retention) Introduction {
	t.Helper()
	consent, err := NewConsentSnapshot("voice.introduction", 3, testTime)
	if err != nil {
		t.Fatal(err)
	}
	media, err := NewMediaRef("asset:1", "audio/ogg", 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	introduction, err := New(
		"introduction:1", "member:1", PromptArrival, consent, media, retention,
		command("command:create", testTime),
	)
	if err != nil {
		t.Fatal(err)
	}
	return introduction
}

func uploaded(t *testing.T) Introduction {
	t.Helper()
	introduction := draft(t, NewRetention(time.Time{}, false))
	authorized, err := introduction.AuthorizeUpload(command("command:upload", testTime), 1)
	if err != nil {
		t.Fatal(err)
	}
	media, _ := NewMediaRef(
		"asset:1", "audio/ogg", 1024, 45*time.Second, strings.Repeat("b", 64),
	)
	complete, err := authorized.ConfirmUpload(media, command("command:confirm", testTime), 2)
	if err != nil {
		t.Fatal(err)
	}
	return complete
}

func TestVersionedConsentAndMediaMetadataAreRetainedWithoutBytes(t *testing.T) {
	introduction := uploaded(t)
	if introduction.Consent().PurposeID() != "voice.introduction" ||
		introduction.Consent().Version() != 3 ||
		introduction.Media().Size() != 1024 ||
		introduction.Media().Duration() != 45*time.Second {
		t.Fatalf("metadata missing: %+v", introduction)
	}
	if introduction.Status() != StatusUploaded || introduction.Version() != 3 {
		t.Fatalf("status=%q version=%d", introduction.Status(), introduction.Version())
	}
	for _, event := range introduction.Events() {
		if event.Fingerprint() == introduction.OwnerID() ||
			event.Fingerprint() == introduction.Media().AssetID() {
			t.Fatal("audit event leaked a raw identifier")
		}
	}
}

func TestTranscriptionLifecycleStoresReferenceNotText(t *testing.T) {
	introduction := uploaded(t)
	transcribing, err := introduction.StartTranscription(command("command:start", testTime), 3)
	if err != nil {
		t.Fatal(err)
	}
	transcript, _ := NewTranscriptRef("transcript:1", "tw", 87)
	ready, err := transcribing.CompleteTranscription(
		transcript, command("command:ready", testTime), 4,
	)
	if err != nil {
		t.Fatal(err)
	}
	if ready.Status() != StatusReady || ready.Transcript().ID() != "transcript:1" ||
		ready.Transcript().Language() != "tw" {
		t.Fatalf("ready=%+v", ready)
	}
}

func TestUncertainAndFailureNeverBecomeReady(t *testing.T) {
	for _, test := range []struct {
		name string
		run  func(Introduction) (Introduction, error)
		want Status
	}{
		{"uncertain", func(value Introduction) (Introduction, error) {
			return value.TranscriptionUncertain(command("command:uncertain", testTime), 4)
		}, StatusTranscriptionUncertain},
		{"failed", func(value Introduction) (Introduction, error) {
			return value.TranscriptionFailed(command("command:failed", testTime), 4)
		}, StatusTranscriptionFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			transcribing, _ := uploaded(t).StartTranscription(command("command:start", testTime), 3)
			result, err := test.run(transcribing)
			if err != nil || result.Status() != test.want || result.Transcript().ID() != "" {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestCommandReplayAndOptimisticVersion(t *testing.T) {
	introduction := draft(t, NewRetention(time.Time{}, false))
	change := command("command:upload", testTime)
	authorized, err := introduction.AuthorizeUpload(change, 1)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := authorized.AuthorizeUpload(change, 1)
	if !errors.Is(err, ErrInvalidTransition) {
		// Lifecycle checks deliberately precede replay for an action that is
		// no longer valid; application-level command replay returns stored state.
		t.Fatalf("expected closed transition, got %+v, %v", replayed, err)
	}
	if _, err := introduction.AuthorizeUpload(command("command:other", testTime), 2); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("expected stale version, got %v", err)
	}
	conflict := command("command:create", testTime)
	conflict.Fingerprint = strings.Repeat("c", 64)
	if _, err := introduction.transition(conflict, ActionCreated, StatusDraft, 1); !errors.Is(err, ErrCommandConflict) {
		t.Fatalf("expected command conflict, got %v", err)
	}
}

func TestRevocationAndCancellationScheduleRetentionAwarePurge(t *testing.T) {
	retainUntil := testTime.Add(24 * time.Hour)
	introduction := draft(t, NewRetention(retainUntil, false))
	revoked, err := introduction.Revoke(command("command:revoke", testTime.Add(time.Hour)), 1)
	if err != nil {
		t.Fatal(err)
	}
	if revoked.Status() != StatusRevoked || revoked.DataStatus() != DataPurgePending ||
		!revoked.DeletionDueAt().Equal(retainUntil) {
		t.Fatalf("revoked=%+v", revoked)
	}
	if _, err := revoked.MarkPurged(command("command:purge", retainUntil.Add(-time.Second)), 2); !errors.Is(err, ErrRetentionActive) {
		t.Fatalf("expected retention active, got %v", err)
	}
	purged, err := revoked.MarkPurged(command("command:purge", retainUntil), 2)
	if err != nil || purged.DataStatus() != DataPurged || purged.Media().AssetID() != "" {
		t.Fatalf("purged=%+v, err=%v", purged, err)
	}
}

func TestLegalHoldPreventsPurge(t *testing.T) {
	introduction := draft(t, NewRetention(time.Time{}, true))
	cancelled, err := introduction.Cancel(command("command:cancel", testTime), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled.DeletionDueAt().IsZero() {
		t.Fatal("legal hold should not receive an executable deletion deadline")
	}
	if _, err := cancelled.MarkPurged(command("command:purge", testTime), 2); !errors.Is(err, ErrLegalHold) {
		t.Fatalf("expected legal hold, got %v", err)
	}
}

func TestDuplicateTakesDoNotFinishAnIntroduction(t *testing.T) {
	// The rule that stops three takes of "what brought you here" earning the
	// sowing rung. Counting recordings rather than distinct questions is the
	// obvious wrong implementation, so it is the one worth pinning.
	if Complete([]Prompt{PromptArrival, PromptArrival, PromptArrival}) {
		t.Fatal("three takes of one question finished an introduction")
	}
	if Complete([]Prompt{PromptArrival, PromptOrdinary, PromptOrdinary, PromptWelcome, PromptWelcome}) != true {
		t.Fatal("all three questions answered, with repeats, did not finish")
	}
	if Complete([]Prompt{PromptArrival, PromptOrdinary}) {
		t.Fatal("two questions finished an introduction")
	}
	if Complete(nil) {
		t.Fatal("no recordings finished an introduction")
	}
	if !Complete([]Prompt{PromptWelcome, PromptArrival, PromptOrdinary}) {
		t.Fatal("all three questions, in any order, did not finish")
	}
}
