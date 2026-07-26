package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDoorwayAlternatesAndSealsAtExactlyThree(t *testing.T) {
	now := time.Now().UTC()
	doorway, err := Open("d", "alice", "bob", now)
	if err != nil {
		t.Fatal(err)
	}
	actors := []string{"alice", "bob", "alice"}
	for i, actor := range actors {
		doorway, _, err = doorway.Exchange(actor, "message", "command-"+string(rune('a'+i)), "fp-"+string(rune('a'+i)), now.Add(time.Duration(i)*time.Second))
		if err != nil {
			t.Fatal(err)
		}
	}
	if !doorway.Sealed() || len(doorway.Exchanges()) != 3 {
		t.Fatalf("doorway=%#v", doorway)
	}
	if _, _, err = doorway.Exchange("bob", "message", "fourth", "fp", now); !errors.Is(err, ErrDoorwaySealed) {
		t.Fatalf("fourth=%v", err)
	}
}

func TestDoorwayRejectsUnrelatedAndRepeatedTurns(t *testing.T) {
	now := time.Now()
	doorway, _ := Open("d", "alice", "bob", now)
	if _, _, err := doorway.Exchange("mallory", "m", "c", "f", now); !errors.Is(err, ErrNotParticipant) {
		t.Fatalf("got %v", err)
	}
	doorway, _, _ = doorway.Exchange("alice", "m", "c1", "f1", now)
	if _, _, err := doorway.Exchange("alice", "m", "c2", "f2", now); !errors.Is(err, ErrOutOfTurn) {
		t.Fatalf("got %v", err)
	}
}

func FuzzAlternationNeverExceedsThree(f *testing.F) {
	f.Add(uint8(7))
	f.Fuzz(func(t *testing.T, bits uint8) {
		now := time.Unix(0, 0)
		doorway, _ := Open("d", "a", "b", now)
		for i := 0; i < 8; i++ {
			actor := "a"
			if bits&(1<<i) != 0 {
				actor = "b"
			}
			next, changed, _ := doorway.Exchange(actor, "m", string(rune('a'+i)), string(rune('A'+i)), now)
			if changed {
				doorway = next
			}
		}
		exchanges := doorway.Exchanges()
		if len(exchanges) > 3 {
			t.Fatalf("count=%d", len(exchanges))
		}
		for i := 1; i < len(exchanges); i++ {
			if exchanges[i].ActorKey == exchanges[i-1].ActorKey {
				t.Fatal("did not alternate")
			}
		}
	})
}
