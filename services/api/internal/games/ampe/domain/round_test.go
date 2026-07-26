package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestChoicesStayHiddenUntilAtomicSecondLockReveal(t *testing.T) {
	spec := validSpec()
	round := mustOpen(t, spec)
	round = mustApply(t, round, command("ready-a", spec.PlayerKeys[0], ActionReady, "", 0))
	round = mustApply(t, round, command("ready-b", spec.PlayerKeys[1], ActionReady, "", 1))
	round = mustApply(t, round, command("lock-a", spec.PlayerKeys[0], ActionLock, ChoiceTogether, 2))

	owner := mustView(t, round, spec.PlayerKeys[0])
	opponent := mustView(t, round, spec.PlayerKeys[1])
	if owner.OwnChoice == nil || *owner.OwnChoice != ChoiceTogether {
		t.Fatalf("owner choice = %v", owner.OwnChoice)
	}
	if opponent.OwnChoice != nil || opponent.Reveal != nil {
		t.Fatalf("opponent saw choice before reveal: %+v", opponent)
	}
	publicEvidence, err := json.Marshal(struct {
		View   View
		Events []Event
	}{opponent, round.Events()})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicEvidence), string(ChoiceTogether)) {
		t.Fatalf("public-safe state leaked choice: %s", publicEvidence)
	}

	round = mustApply(t, round, command("lock-b", spec.PlayerKeys[1], ActionLock, ChoiceApart, 3))
	revealed := mustView(t, round, spec.PlayerKeys[1])
	if revealed.Reveal == nil || revealed.Reveal.Sequence != 4 ||
		revealed.Reveal.Choices != [2]Choice{ChoiceTogether, ChoiceApart} {
		t.Fatalf("simultaneous reveal = %+v", revealed.Reveal)
	}
}

func TestDisconnectPausesWithoutForfeitOrExposure(t *testing.T) {
	spec := validSpec()
	round := readyRound(t, spec)
	round = mustApply(t, round, command("lock-a", spec.PlayerKeys[0], ActionLock, ChoiceApart, 2))
	round = mustApply(t, round, command("disconnect-b", spec.PlayerKeys[1], ActionDisconnect, "", 3))

	view := mustView(t, round, spec.PlayerKeys[1])
	if !view.Paused || view.Players[1].Connected || view.Reveal != nil || view.OwnChoice != nil {
		t.Fatalf("disconnect state = %+v", view)
	}
	if _, _, err := round.Apply(command("lock-while-paused", spec.PlayerKeys[1], ActionLock, ChoiceTogether, 4)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("lock while paused = %v", err)
	}

	round = mustApply(t, round, command("reconnect-b", spec.PlayerKeys[1], ActionReconnect, "", 4))
	if view = mustView(t, round, spec.PlayerKeys[0]); view.Paused || view.Reveal != nil {
		t.Fatalf("reconnect invented outcome = %+v", view)
	}
	round = mustApply(t, round, command("lock-b", spec.PlayerKeys[1], ActionLock, ChoiceTogether, 5))
	if mustView(t, round, spec.PlayerKeys[0]).Reveal == nil {
		t.Fatal("both locks did not reveal after safe reconnect")
	}
}

func TestServerSequenceAndIdempotency(t *testing.T) {
	spec := validSpec()
	round := mustOpen(t, spec)
	ready := command("ready-a", spec.PlayerKeys[0], ActionReady, "", 0)
	applied, replayed, err := round.Apply(ready)
	if err != nil || replayed {
		t.Fatalf("first apply replayed=%v err=%v", replayed, err)
	}
	replayedRound, replayed, err := applied.Apply(ready)
	if err != nil || !replayed || !reflect.DeepEqual(applied, replayedRound) {
		t.Fatalf("idempotent retry replayed=%v err=%v", replayed, err)
	}
	mismatch := ready
	mismatch.Action = ActionDisconnect
	if _, _, err = applied.Apply(mismatch); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("command mismatch = %v", err)
	}
	if _, _, err = applied.Apply(command("stale", spec.PlayerKeys[1], ActionReady, "", 0)); !errors.Is(err, ErrStaleSequence) {
		t.Fatalf("stale sequence = %v", err)
	}
}

func TestPrivateTranscriptReplaysAndRejectsTampering(t *testing.T) {
	spec := validSpec()
	round := readyRound(t, spec)
	round = mustApply(t, round, command("lock-a", spec.PlayerKeys[0], ActionLock, ChoiceTogether, 2))
	round = mustApply(t, round, command("lock-b", spec.PlayerKeys[1], ActionLock, ChoiceTogether, 3))
	transcript := round.PrivateTranscript()
	replayed, err := Replay(spec, transcript)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(round, replayed) {
		t.Fatal("deterministic replay changed round")
	}
	transcript.FinalDigest = strings.Repeat("0", 64)
	if _, err = Replay(spec, transcript); !errors.Is(err, ErrTranscript) {
		t.Fatalf("tampered transcript = %v", err)
	}
}

func TestInvalidActorsChoicesAndRoundShapeAreRejected(t *testing.T) {
	spec := validSpec()
	duplicate := spec
	duplicate.PlayerKeys[1] = duplicate.PlayerKeys[0]
	if _, err := Open(duplicate); !errors.Is(err, ErrInvalidRound) {
		t.Fatalf("duplicate players = %v", err)
	}
	round := readyRound(t, spec)
	if _, _, err := round.Apply(command("bad-choice", spec.PlayerKeys[0], ActionLock, "camera_pose", 2)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("unbounded choice = %v", err)
	}
	if _, err := round.View(strings.Repeat("d", 64)); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("non-player view = %v", err)
	}
}

func TestRoundInvariantsAcrossDeterministicScenarios(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 2000; trial++ {
		spec := validSpec()
		round := readyRound(t, spec)
		first := ChoiceTogether
		second := ChoiceTogether
		if rng.Intn(2) == 1 {
			first = ChoiceApart
		}
		if rng.Intn(2) == 1 {
			second = ChoiceApart
		}
		round = mustApply(t, round, command("lock-a", spec.PlayerKeys[0], ActionLock, first, 2))
		if rng.Intn(2) == 1 {
			round = mustApply(t, round, command("disconnect-b", spec.PlayerKeys[1], ActionDisconnect, "", 3))
			if mustView(t, round, spec.PlayerKeys[1]).Reveal != nil {
				t.Fatalf("trial %d disconnect revealed", trial)
			}
			round = mustApply(t, round, command("reconnect-b", spec.PlayerKeys[1], ActionReconnect, "", 4))
		}
		round = mustApply(t, round, command("lock-b", spec.PlayerKeys[1], ActionLock, second, round.Sequence()))
		view := mustView(t, round, spec.PlayerKeys[0])
		if view.Reveal == nil || view.Reveal.Choices != [2]Choice{first, second} || view.Paused {
			t.Fatalf("trial %d final view = %+v", trial, view)
		}
	}
}

func TestImmutableRoundCallsAreRaceSafe(t *testing.T) {
	spec := validSpec()
	base := readyRound(t, spec)
	command := command("lock-a", spec.PlayerKeys[0], ActionLock, ChoiceTogether, 2)
	want := mustApply(t, base, command)
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				got, replayed, err := base.Apply(command)
				if err != nil || replayed || !reflect.DeepEqual(got, want) {
					t.Errorf("concurrent apply replayed=%v err=%v", replayed, err)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzRoundNeverRevealsOneChoice(f *testing.F) {
	f.Add(byte(0), byte(1), false)
	f.Add(byte(1), byte(1), true)
	f.Fuzz(func(t *testing.T, firstRaw, secondRaw byte, disconnect bool) {
		spec := validSpec()
		round := readyRound(t, spec)
		choices := [2]Choice{ChoiceTogether, ChoiceApart}
		first := choices[firstRaw%2]
		second := choices[secondRaw%2]
		round = mustApply(t, round, command("lock-a", spec.PlayerKeys[0], ActionLock, first, 2))
		if view := mustView(t, round, spec.PlayerKeys[1]); view.Reveal != nil || view.OwnChoice != nil {
			t.Fatal("first choice exposed")
		}
		if disconnect {
			round = mustApply(t, round, command("disconnect-b", spec.PlayerKeys[1], ActionDisconnect, "", 3))
			round = mustApply(t, round, command("reconnect-b", spec.PlayerKeys[1], ActionReconnect, "", 4))
		}
		round = mustApply(t, round, command("lock-b", spec.PlayerKeys[1], ActionLock, second, round.Sequence()))
		if reveal := mustView(t, round, spec.PlayerKeys[0]).Reveal; reveal == nil || reveal.Choices != [2]Choice{first, second} {
			t.Fatalf("bad reveal %+v", reveal)
		}
	})
}

func validSpec() Spec {
	return Spec{
		ID:      "ampe-round-one",
		RoomKey: strings.Repeat("a", 64),
		PlayerKeys: [2]string{
			strings.Repeat("b", 64),
			strings.Repeat("c", 64),
		},
	}
}

func command(id, actor string, action Action, choice Choice, sequence uint64) Command {
	return Command{ID: id, ActorKey: actor, Action: action, Choice: choice, ExpectedSequence: sequence}
}

func mustOpen(t *testing.T, spec Spec) Round {
	t.Helper()
	round, err := Open(spec)
	if err != nil {
		t.Fatal(err)
	}
	return round
}

func readyRound(t *testing.T, spec Spec) Round {
	t.Helper()
	round := mustOpen(t, spec)
	round = mustApply(t, round, command("ready-a", spec.PlayerKeys[0], ActionReady, "", 0))
	return mustApply(t, round, command("ready-b", spec.PlayerKeys[1], ActionReady, "", 1))
}

func mustApply(t *testing.T, round Round, command Command) Round {
	t.Helper()
	next, replayed, err := round.Apply(command)
	if err != nil {
		t.Fatal(err)
	}
	if replayed {
		t.Fatal("unexpected replay")
	}
	return next
}

func mustView(t *testing.T, round Round, playerKey string) View {
	t.Helper()
	view, err := round.View(playerKey)
	if err != nil {
		t.Fatal(err)
	}
	return view
}
