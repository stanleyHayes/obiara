package domain

import (
	"errors"
	"testing"
	"time"
)

func turn(actor string, base uint64, id string) Command {
	return Command{
		ID: id, DeviceKey: key(2), ActorKey: actor,
		PayloadKey: key(4), Fingerprint: key(5),
		BaseSequence: base, At: time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC),
	}
}

func TestAMemberCannotSpeakTwiceInARow(t *testing.T) {
	// FR-301 and IM-029: this is the rule that makes a room a courtship
	// rather than an inbox. Until it existed, the queue happily accepted any
	// number of consecutive turns from one member.
	state, err := Open(key(1))
	if err != nil {
		t.Fatal(err)
	}
	amma, kofi := key(3), key(6)

	// The first turn in an empty room belongs to whoever takes it.
	state, _, err = state.Accept(turn(amma, 0, "cmd_1"))
	if err != nil {
		t.Fatalf("the first turn was refused: %v", err)
	}
	if _, _, err := state.Accept(turn(amma, 1, "cmd_2")); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("a second consecutive turn returned %v, want ErrNotYourTurn", err)
	}
	// The other member answering is what reopens it.
	state, _, err = state.Accept(turn(kofi, 1, "cmd_3"))
	if err != nil {
		t.Fatalf("the answer was refused: %v", err)
	}
	if _, _, err := state.Accept(turn(amma, 2, "cmd_4")); err != nil {
		t.Fatalf("a turn after being answered was refused: %v", err)
	}
	if _, _, err := state.Accept(turn(kofi, 2, "cmd_5")); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("the other member also got two in a row: %v", err)
	}
}

func TestBeingBehindIsReportedAsBehindNotAsOutOfTurn(t *testing.T) {
	// A member whose partner has just replied is stale, not out of turn, and
	// telling them to catch up is both true and the thing that lets them
	// send. Reporting "not your turn" there would be a dead end.
	state, err := Open(key(1))
	if err != nil {
		t.Fatal(err)
	}
	amma := key(3)
	state, _, err = state.Accept(turn(amma, 0, "cmd_1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := state.Accept(turn(amma, 0, "cmd_2")); !errors.Is(err, ErrStaleDevice) {
		t.Fatalf("a stale cursor returned %v, want ErrStaleDevice", err)
	}
}

func TestARehydratedRoomStillRemembersWhoSpokeLast(t *testing.T) {
	// The rule has to survive the process that enforced it. If the last
	// actor were dropped on reload, one restart would hand out a free
	// consecutive turn.
	amma := key(3)
	state, err := Rehydrate(key(1), 4, 4, amma)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := state.Accept(turn(amma, 4, "cmd_1")); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("a reloaded room forgot who spoke last: %v", err)
	}
}

func TestAMalformedLastActorIsRefusedRatherThanIgnored(t *testing.T) {
	// A value that is present but not a real key would never equal any
	// actor, silently turning the rule off for that room.
	if _, err := Rehydrate(key(1), 1, 1, "not-a-key"); !errors.Is(err, ErrInvalid) {
		t.Fatal("a malformed last actor was accepted")
	}
	// Absent stays legitimate: an unopened room, or one written before turns
	// were recorded.
	if _, err := Rehydrate(key(1), 0, 0, ""); err != nil {
		t.Fatalf("an empty last actor was refused: %v", err)
	}
}
