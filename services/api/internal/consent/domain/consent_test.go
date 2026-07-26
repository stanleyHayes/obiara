package domain

import (
	"errors"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func testPurpose(t *testing.T, kind PurposeKind, version uint64) Purpose {
	t.Helper()
	purpose, err := NewPurpose(NewPurposeParams{
		ID:   "promise.community",
		Kind: kind, Version: version, ContentRef: "content.promise.v1",
		Status: PurposeActive, EffectiveSince: fixedTime.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	return purpose
}

func testEvidence(t *testing.T, kind EvidenceKind, version uint64) Evidence {
	t.Helper()
	evidence, err := NewEvidence(kind, version, "evidence.policy.v1")
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func testChange(t *testing.T, purpose Purpose, command string, revision uint64) Change {
	t.Helper()
	return Change{
		CommandID: command, ExpectedRevision: revision, Purpose: purpose,
		ActorID: "member:123", ActorKind: ActorSubject, Source: SourceWeb,
		Evidence:   testEvidence(t, EvidenceAcknowledgement, purpose.Version()),
		RecordedAt: fixedTime,
	}
}

func TestPurposeValidationAndVersioning(t *testing.T) {
	tests := []struct {
		name string
		edit func(*NewPurposeParams)
	}{
		{"zero version", func(params *NewPurposeParams) { params.Version = 0 }},
		{"free form id", func(params *NewPurposeParams) { params.ID = "terms and email@example.com" }},
		{"raw content", func(params *NewPurposeParams) { params.ContentRef = "I accept these terms" }},
		{"zero effective time", func(params *NewPurposeParams) { params.EffectiveSince = time.Time{} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := NewPurposeParams{
				ID: "terms.service", Kind: PurposeTerms, Version: 1,
				ContentRef: "content.terms.v1", Status: PurposeActive, EffectiveSince: fixedTime,
			}
			test.edit(&params)
			if _, err := NewPurpose(params); !errors.Is(err, ErrInvalidPurpose) {
				t.Fatalf("expected invalid purpose, got %v", err)
			}
		})
	}

	first := testPurpose(t, PurposePromise, 1)
	second := testPurpose(t, PurposePromise, 2)
	if first.Version() != 1 || second.Version() != 2 || first.Version() == second.Version() {
		t.Fatal("purpose versions must remain immutable and distinct")
	}
}

func TestConsentLifecycleAndImmutableHistory(t *testing.T) {
	purpose := testPurpose(t, PurposePromise, 1)
	record, err := NewRecord("member:123", purpose.ID())
	if err != nil {
		t.Fatal(err)
	}
	granted, err := record.Grant(testChange(t, purpose, "command:grant", 0))
	if err != nil {
		t.Fatal(err)
	}
	if record.Revision() != 0 || len(record.History()) != 0 {
		t.Fatal("grant mutated the original record")
	}
	if !granted.Effective(purpose, fixedTime) {
		t.Fatal("current grant should be effective")
	}

	change := testChange(t, purpose, "command:withdraw", 1)
	withdrawn, err := granted.Withdraw(change)
	if err != nil {
		t.Fatal(err)
	}
	if withdrawn.Effective(purpose, fixedTime) || len(withdrawn.History()) != 2 {
		t.Fatal("withdrawal must revoke consent and retain history")
	}
	history := withdrawn.History()
	history[0] = Event{}
	if withdrawn.History()[0].Action() != ActionGranted {
		t.Fatal("history getter leaked mutable aggregate storage")
	}
	if _, err := withdrawn.Withdraw(testChange(t, purpose, "command:again", 2)); !errors.Is(err, ErrAlreadyWithdrawn) {
		t.Fatalf("expected already withdrawn, got %v", err)
	}
}

func TestConsentDeniesByDefaultAndOnPurposeChange(t *testing.T) {
	current := testPurpose(t, PurposePromise, 1)
	record, _ := NewRecord("member:123", current.ID())
	if record.Effective(current, fixedTime) {
		t.Fatal("empty record must deny")
	}
	granted, err := record.Grant(testChange(t, current, "command:grant", 0))
	if err != nil {
		t.Fatal(err)
	}
	nextVersion := testPurpose(t, PurposePromise, 2)
	if granted.Effective(nextVersion, fixedTime) {
		t.Fatal("new purpose version must require a new grant")
	}
	retired, _ := NewPurpose(NewPurposeParams{
		ID: current.ID(), Kind: current.Kind(), Version: current.Version(),
		ContentRef: current.ContentRef(), Status: PurposeRetired, EffectiveSince: current.EffectiveSince(),
	})
	if granted.Effective(retired, fixedTime) {
		t.Fatal("retired purpose must deny")
	}
}

func TestEveryNewPurposeVersionInvalidatesPreviousGrant(t *testing.T) {
	first := testPurpose(t, PurposePromise, 1)
	record, _ := NewRecord("member:123", first.ID())
	granted, err := record.Grant(testChange(t, first, "command:grant", 0))
	if err != nil {
		t.Fatal(err)
	}
	for version := uint64(2); version <= 64; version++ {
		next := testPurpose(t, PurposePromise, version)
		if granted.Effective(next, fixedTime) {
			t.Fatalf("grant for version 1 was effective for version %d", version)
		}
	}
}

func TestConsentRejectsStaleReplayAndPIIMetadata(t *testing.T) {
	purpose := testPurpose(t, PurposePromise, 1)
	record, _ := NewRecord("member:123", purpose.ID())
	granted, err := record.Grant(testChange(t, purpose, "command:grant", 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = granted.Grant(testChange(t, purpose, "command:stale", 0)); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("expected stale revision, got %v", err)
	}
	original := testChange(t, purpose, "command:grant", 1)
	replayed, err := granted.Grant(original)
	if err != nil || replayed.Revision() != granted.Revision() {
		t.Fatalf("same command must be replay safe: revision=%d err=%v", replayed.Revision(), err)
	}
	if !granted.ReplayMatches(original, ActionGranted) ||
		granted.ReplayMatches(original, ActionWithdrawn) {
		t.Fatal("command replay must preserve the original action and payload")
	}
	for _, invalid := range []string{"person@example.com", "Member Name", "10.1.2.3", ""} {
		if _, err := NewRecord(invalid, purpose.ID()); !errors.Is(err, ErrInvalidIdentity) {
			t.Errorf("identity %q should be rejected, got %v", invalid, err)
		}
	}
}

func TestAgeAffirmationIsEvidenceNotVerification(t *testing.T) {
	purpose, err := NewPurpose(NewPurposeParams{
		ID: "age.minimum", Kind: PurposeAge, Version: 3,
		ContentRef: "content.age.v3", Status: PurposeActive, EffectiveSince: fixedTime.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	record, _ := NewRecord("member:123", purpose.ID())
	change := testChange(t, purpose, "command:age", 0)
	if _, err = record.Grant(change); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("age consent must require affirmation evidence, got %v", err)
	}
	change.Evidence = testEvidence(t, EvidenceAgeAffirmation, 3)
	granted, err := record.Grant(change)
	if err != nil || !granted.Effective(purpose, fixedTime) {
		t.Fatalf("age affirmation should be retained as evidence: %v", err)
	}
	event := granted.History()[0]
	if event.Evidence().Kind() != EvidenceAgeAffirmation || event.Evidence().PolicyVersion() != 3 {
		t.Fatal("age affirmation evidence was not retained")
	}
}

func TestRehydrateRejectsNonContiguousHistory(t *testing.T) {
	evidence := testEvidence(t, EvidenceAcknowledgement, 1)
	event, err := NewEvent(EventParams{
		Revision: 2, CommandID: "command:grant", Action: ActionGranted,
		PurposeVersion: 1, ActorID: "member:123", ActorKind: ActorSubject,
		Source: SourceWeb, Evidence: evidence, RecordedAt: fixedTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Rehydrate("member:123", "promise.community", []Event{event}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected invalid history, got %v", err)
	}
}

func TestRehydrateRejectsDuplicateCommandHistory(t *testing.T) {
	evidence := testEvidence(t, EvidenceAcknowledgement, 1)
	first, _ := NewEvent(EventParams{
		Revision: 1, CommandID: "command:same", Action: ActionGranted,
		PurposeVersion: 1, ActorID: "member:123", ActorKind: ActorSubject,
		Source: SourceWeb, Evidence: evidence, RecordedAt: fixedTime,
	})
	second, _ := NewEvent(EventParams{
		Revision: 2, CommandID: "command:same", Action: ActionWithdrawn,
		PurposeVersion: 1, ActorID: "member:123", ActorKind: ActorSubject,
		Source: SourceWeb, Evidence: evidence, RecordedAt: fixedTime.Add(time.Minute),
	})
	if _, err := Rehydrate("member:123", "promise.community", []Event{first, second}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected duplicate history rejection, got %v", err)
	}
}
