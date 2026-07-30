package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC)

func newAttempt(t *testing.T) Attempt {
	t.Helper()
	attempt, err := NewAttempt(
		"liveness:1", "command:1", strings.Repeat("a", 64),
		strings.Repeat("b", 64), fixedTime,
	)
	if err != nil {
		t.Fatal(err)
	}
	return attempt
}

func TestOnlyExplicitLiveDecisionPasses(t *testing.T) {
	actor := strings.Repeat("c", 64)
	tests := []struct {
		name   string
		change func(Attempt) (Attempt, error)
		passed bool
	}{
		{"live", func(value Attempt) (Attempt, error) {
			return value.ProviderDecision(true, "provider:proof", actor, fixedTime.Add(time.Second), 1)
		}, true},
		{"not live", func(value Attempt) (Attempt, error) {
			return value.ProviderDecision(false, "provider:proof", actor, fixedTime.Add(time.Second), 1)
		}, false},
		{"uncertain", func(value Attempt) (Attempt, error) {
			return value.QueueManual(ReasonProviderUncertain, actor, fixedTime.Add(time.Second), 1)
		}, false},
		{"outage", func(value Attempt) (Attempt, error) {
			return value.QueueManual(ReasonProviderUnavailable, actor, fixedTime.Add(time.Second), 1)
		}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := newAttempt(t)
			updated, err := test.change(original)
			if err != nil || updated.Passed() != test.passed {
				t.Fatalf("passed=%v, err=%v", updated.Passed(), err)
			}
			if original.Status() != StatusPending || original.Version() != 1 {
				t.Fatal("transition mutated the original aggregate")
			}
		})
	}
}

func TestReconstituteRejectsHistoryDrift(t *testing.T) {
	attempt := newAttempt(t)
	decided, err := attempt.ProviderDecision(
		true, "provider:proof", strings.Repeat("c", 64),
		fixedTime.Add(time.Second), attempt.Version(),
	)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := Reconstitute(
		decided.ID(), decided.CommandID(), decided.SubjectKey(), decided.InputKey(),
		decided.Status(), decided.Reason(), decided.ProviderRef(),
		decided.CreatedAt(), decided.DecidedAt(), decided.Version(), decided.Events(),
	)
	if err != nil || !restored.Passed() || restored.Version() != decided.Version() {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
	bad := decided.Events()
	bad[0], _ = NewEvent(EventParams{
		Status: StatusFailed, Reason: ReasonProviderNotLive,
		ActorKey: strings.Repeat("c", 64), OccurredAt: fixedTime.Add(time.Second),
		Version: 2,
	})
	if _, err := Reconstitute(
		decided.ID(), decided.CommandID(), decided.SubjectKey(), decided.InputKey(),
		decided.Status(), decided.Reason(), decided.ProviderRef(),
		decided.CreatedAt(), decided.DecidedAt(), decided.Version(), bad,
	); err == nil {
		t.Fatal("accepted history whose terminal event disagrees with state")
	}
}

func TestManualDecisionRequiresQueuedAttempt(t *testing.T) {
	attempt := newAttempt(t)
	actor := strings.Repeat("c", 64)
	if _, err := attempt.ManualDecision(true, actor, fixedTime, 1); !errors.Is(err, ErrManualReviewOnly) {
		t.Fatalf("expected manual-review-only, got %v", err)
	}
	queued, _ := attempt.QueueManual(ReasonProviderUncertain, actor, fixedTime, 1)
	passed, err := queued.ManualDecision(true, strings.Repeat("d", 64), fixedTime.Add(time.Minute), 2)
	if err != nil || !passed.Passed() || len(passed.Events()) != 2 {
		t.Fatalf("manual pass = %+v, %v", passed, err)
	}
}

func TestStaleAndRepeatedDecisionsAreRejected(t *testing.T) {
	attempt := newAttempt(t)
	actor := strings.Repeat("c", 64)
	if _, err := attempt.ProviderDecision(true, "provider:1", actor, fixedTime, 2); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("expected stale version, got %v", err)
	}
	passed, _ := attempt.ProviderDecision(true, "provider:1", actor, fixedTime, 1)
	if _, err := passed.ProviderDecision(true, "provider:1", actor, fixedTime, 2); !errors.Is(err, ErrAttemptNotOpen) {
		t.Fatalf("expected closed attempt, got %v", err)
	}
}
