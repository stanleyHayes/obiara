package domain

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"testing"
)

func TestPrivateDuelAlternatesBoundedTurnsAndReplays(t *testing.T) {
	spec := duelSpec(t, 6)
	duel, err := NewDuel(spec)
	if err != nil {
		t.Fatal(err)
	}
	for turn := range 6 {
		answer := "wrong"
		if turn%2 == 0 {
			answer = "test answer"
		}
		duel, err = duel.Answer(spec.PlayerKeys[turn%2], answer, duel.Revision())
		if err != nil {
			t.Fatalf("turn %d: %v", turn, err)
		}
	}
	if !duel.Complete() || duel.Revision() != 6 || duel.Scores() != [2]uint8{3, 0} {
		t.Fatalf("finished duel = complete %v revision %d scores %v", duel.Complete(), duel.Revision(), duel.Scores())
	}
	replayed, err := Replay(spec, duel.Turns())
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Scores() != duel.Scores() || replayed.PromptVersions()[0] != duel.PromptVersions()[0] {
		t.Fatal("replay changed the duel")
	}
}

func TestDuelRejectsWrongPlayerStaleRevisionAndFurtherTurns(t *testing.T) {
	spec := duelSpec(t, 1)
	duel, err := NewDuel(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = duel.Answer(spec.PlayerKeys[1], "test answer", 0); !errors.Is(err, ErrUnexpectedPlayer) {
		t.Fatalf("wrong player = %v", err)
	}
	if _, err = duel.Answer(spec.PlayerKeys[0], "test answer", 1); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("stale revision = %v", err)
	}
	duel, err = duel.Answer(spec.PlayerKeys[0], "test answer", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = duel.Answer(spec.PlayerKeys[1], "test answer", duel.Revision()); !errors.Is(err, ErrDuelComplete) {
		t.Fatalf("post-completion answer = %v", err)
	}
}

func TestDuelRejectsTamperedReplayAndUnreviewedSnapshot(t *testing.T) {
	spec := duelSpec(t, 2)
	duel, err := NewDuel(spec)
	if err != nil {
		t.Fatal(err)
	}
	duel, err = duel.Answer(spec.PlayerKeys[0], "test answer", 0)
	if err != nil {
		t.Fatal(err)
	}
	tampered := duel.Turns()
	tampered[0].Correct = false
	if _, err = Replay(spec, tampered); !errors.Is(err, ErrReplayMismatch) {
		t.Fatalf("tampered correctness = %v", err)
	}
	spec.Prompts[0] = Prompt{}
	if _, err = NewDuel(spec); !errors.Is(err, ErrPromptUnapproved) {
		t.Fatalf("unreviewed snapshot = %v", err)
	}
}

func TestDuelBoundsAndCopiesCallerData(t *testing.T) {
	if _, err := NewDuel(duelSpec(t, MaxTurns+1)); !errors.Is(err, ErrDuelInvalid) {
		t.Fatalf("too many turns = %v", err)
	}
	spec := duelSpec(t, 2)
	duel, err := NewDuel(spec)
	if err != nil {
		t.Fatal(err)
	}
	spec.Prompts[0] = Prompt{}
	if duel.Prompts()[0].Digest() == "" {
		t.Fatal("caller mutated duel prompt snapshot")
	}
}

func TestDeterministicDuelProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 1000; trial++ {
		count := rng.Intn(MaxTurns) + 1
		spec := duelSpec(t, count)
		duel, err := NewDuel(spec)
		if err != nil {
			t.Fatal(err)
		}
		correct := [2]uint8{}
		for turn := range count {
			answer := "wrong"
			if rng.Intn(2) == 1 {
				answer = "reviewed variant"
				correct[turn%2]++
			}
			duel, err = duel.Answer(spec.PlayerKeys[turn%2], answer, duel.Revision())
			if err != nil {
				t.Fatalf("trial %d turn %d: %v", trial, turn, err)
			}
		}
		if duel.Revision() > MaxTurns || duel.Scores() != correct || !duel.Complete() {
			t.Fatalf("trial %d invalid result revision=%d score=%v want=%v", trial, duel.Revision(), duel.Scores(), correct)
		}
		if _, err = Replay(spec, duel.Turns()); err != nil {
			t.Fatalf("trial %d replay: %v", trial, err)
		}
	}
}

func TestPrivateDuelReplayIsRaceSafe(t *testing.T) {
	spec := duelSpec(t, 8)
	duel, err := NewDuel(spec)
	if err != nil {
		t.Fatal(err)
	}
	for turn := range 8 {
		duel, err = duel.Answer(spec.PlayerKeys[turn%2], "test answer", duel.Revision())
		if err != nil {
			t.Fatal(err)
		}
	}
	transcript := duel.Turns()
	var wait sync.WaitGroup
	for range 24 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				replayed, replayErr := Replay(spec, transcript)
				if replayErr != nil || replayed.Scores() != duel.Scores() {
					t.Errorf("concurrent replay: %v", replayErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzPrivateDuelReplay(f *testing.F) {
	f.Add([]byte{0, 1, 1, 0, 1})
	f.Add([]byte("private-reviewed-duel"))
	f.Fuzz(func(t *testing.T, choices []byte) {
		if len(choices) == 0 {
			return
		}
		count := len(choices)
		if count > MaxTurns {
			count = MaxTurns
		}
		spec := duelSpec(t, count)
		duel, err := NewDuel(spec)
		if err != nil {
			t.Fatal(err)
		}
		for turn, choice := range choices[:count] {
			answer := "wrong"
			if choice%2 == 0 {
				answer = "test answer"
			}
			duel, err = duel.Answer(spec.PlayerKeys[turn%2], answer, duel.Revision())
			if err != nil {
				t.Fatal(err)
			}
		}
		replayed, err := Replay(spec, duel.Turns())
		if err != nil {
			t.Fatal(err)
		}
		if replayed.Scores() != duel.Scores() || replayed.Revision() != duel.Revision() {
			t.Fatal("replay changed private duel")
		}
	})
}

func duelSpec(t testingT, count int) DuelSpec {
	t.Helper()
	prompts := make([]Prompt, count)
	for index := range prompts {
		prompts[index] = mustPrompt(t, approvedPromptSpec(
			"prompt-"+strings.Repeat("x", index+1),
			uint64(index+1),
			"ak",
			"Synthetic reviewed cue",
		))
	}
	return DuelSpec{
		ID: "duel-one",
		PlayerKeys: [2]string{
			strings.Repeat("b", 64),
			strings.Repeat("c", 64),
		},
		Prompts: prompts,
	}
}
