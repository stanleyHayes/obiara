package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

func TestBoundaryMutualRelightAndArchive(t *testing.T) {
	now := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	pace, err := New("pace_room_123", []string{key(2), key(1)}, "command_start", key(1), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pace.Advance(Command{ID: "command_early", ExpectedRevision: 1, At: pace.ResponseAt().Add(-time.Nanosecond)}); !errorsIs(err, ErrDenied) {
		t.Fatalf("early transition = %v", err)
	}
	pace, err = pace.Advance(Command{ID: "command_rest", ExpectedRevision: 1, At: pace.ResponseAt()})
	if err != nil || pace.Status() != StatusResting {
		t.Fatalf("rest = %v, %s", err, pace.Status())
	}
	pace, err = pace.Relight(Command{ID: "command_one", ActorKey: key(1), ExpectedRevision: 2, At: now.Add(3 * 24 * time.Hour)})
	if err != nil || pace.Status() != StatusResting {
		t.Fatalf("one grant = %v, %s", err, pace.Status())
	}
	pace, err = pace.Relight(Command{ID: "command_two", ActorKey: key(2), ExpectedRevision: 3, At: now.Add(4 * 24 * time.Hour)})
	if err != nil || pace.Status() != StatusActive || len(pace.RelightGrants()) != 0 {
		t.Fatalf("mutual relight = %v, %s", err, pace.Status())
	}
	pace, _ = pace.Advance(Command{ID: "command_rest2", ExpectedRevision: 4, At: pace.ResponseAt()})
	pace, err = pace.Advance(Command{ID: "command_archive", ExpectedRevision: 5, At: pace.ArchiveAt()})
	if err != nil || pace.Status() != StatusArchived {
		t.Fatalf("archive = %v, %s", err, pace.Status())
	}
}

func TestReplayMismatchAndPrivacyShape(t *testing.T) {
	now := time.Now().UTC()
	pace, _ := New("pace_room_456", []string{key(1), key(2)}, "command_start", key(1), now)
	restAt := pace.ResponseAt()
	changed, _ := pace.Advance(Command{ID: "command_rest", ExpectedRevision: 1, At: restAt})
	replayed, err := changed.Advance(Command{ID: "command_rest", ExpectedRevision: 1, At: restAt})
	if err != nil || replayed.Revision() != changed.Revision() {
		t.Fatalf("replay = %v", err)
	}
	_, err = changed.Relight(Command{ID: "command_rest", ActorKey: key(1), ExpectedRevision: 2, At: restAt})
	if !errorsIs(err, ErrCommandMismatch) {
		t.Fatalf("mismatch = %v", err)
	}
	state := State{ID: changed.ID(), Members: changed.Members(), Status: changed.Status(), ResponseAt: changed.ResponseAt(),
		ArchiveAt: changed.ArchiveAt(), Revision: changed.Revision(), Events: changed.Events(), Commands: changed.Commands()}
	if _, err = Rehydrate(state); err != nil {
		t.Fatal(err)
	}
}

func TestTimeBoundariesAreDeterministic(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for trial := 0; trial < 300; trial++ {
		started := base.Add(time.Duration(trial) * 17 * time.Minute)
		pace, err := New(fmt.Sprintf("pace_trial_%03d", trial), []string{key(1), key(2)}, "command_start", key(1), started)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pace.Advance(Command{ID: "command_before", ExpectedRevision: 1, At: pace.ResponseAt().Add(-time.Nanosecond)}); err != ErrDenied {
			t.Fatalf("trial %d accepted before the response boundary: %v", trial, err)
		}
		pace, err = pace.Advance(Command{ID: "command_rest", ExpectedRevision: 1, At: pace.ResponseAt()})
		if err != nil {
			t.Fatalf("trial %d exact response boundary: %v", trial, err)
		}
		if _, err = pace.Advance(Command{ID: "command_archive_before", ExpectedRevision: 2, At: pace.ArchiveAt().Add(-time.Nanosecond)}); err != ErrDenied {
			t.Fatalf("trial %d accepted before archive boundary: %v", trial, err)
		}
		pace, err = pace.Advance(Command{ID: "command_archive", ExpectedRevision: 2, At: pace.ArchiveAt()})
		if err != nil || pace.Status() != StatusArchived {
			t.Fatalf("trial %d exact archive boundary: %v", trial, err)
		}
	}
}

func errorsIs(got, want error) bool { return got == want }
