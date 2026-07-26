package rating

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	oware "github.com/stanleyHayes/obiara/services/api/internal/games/oware/domain"
)

func TestNotationReplaysExistingDomainMoves(t *testing.T) {
	notation := NewNotation()
	for _, turn := range []struct {
		player oware.Player
		pit    int
	}{
		{oware.South, 2},
		{oware.North, 1},
		{oware.South, 4},
		{oware.North, 3},
	} {
		var err error
		notation, err = notation.Play(turn.player, turn.pit)
		if err != nil {
			t.Fatal(err)
		}
	}
	encoded, err := notation.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ParseNotation(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := replayed.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) || replayed.Board() != notation.Board() {
		t.Fatalf("notation did not replay canonically:\n%s", encoded)
	}
}

func TestNotationRejectsTamperedEvidenceAndTurnOrder(t *testing.T) {
	notation, err := NewNotation().Play(oware.South, 2)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := notation.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(encoded), " 0 ", " 3 ", 1)
	if _, err = ParseNotation([]byte(tampered)); err != ErrNotationReplay {
		t.Fatalf("tampered capture = %v", err)
	}
	if _, err = notation.Play(oware.South, 1); err != ErrNotationInvalid {
		t.Fatalf("consecutive player = %v", err)
	}
}

func TestNotationIsBoundedAndImmutable(t *testing.T) {
	tooLarge := append([]byte(notationHeader+"\n"), bytes.Repeat([]byte("x"), MaxNotationBytes)...)
	if _, err := ParseNotation(tooLarge); err != ErrNotationBound {
		t.Fatalf("oversized notation = %v", err)
	}
	notation, err := NewNotation().Play(oware.South, 2)
	if err != nil {
		t.Fatal(err)
	}
	copyOfPlies := notation.Plies()
	copyOfPlies[0].Pit = 5
	if notation.Plies()[0].Pit != 2 {
		t.Fatal("caller mutated notation")
	}
}

func TestNotationReplayIsRaceSafe(t *testing.T) {
	notation := buildLegalNotation(t, []byte{2, 1, 4, 3, 0, 5, 1, 2, 3, 4})
	encoded, err := notation.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for range 24 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 50 {
				replayed, replayErr := ParseNotation(encoded)
				if replayErr != nil || replayed.Board() != notation.Board() {
					t.Errorf("concurrent replay failed: %v", replayErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzNotationRoundTrip(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte("oware-abapa-replay"))
	f.Fuzz(func(t *testing.T, choices []byte) {
		if len(choices) > MaxPlies {
			choices = choices[:MaxPlies]
		}
		notation := buildLegalNotation(t, choices)
		encoded, err := notation.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		replayed, err := ParseNotation(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if replayed.Board() != notation.Board() || len(replayed.Plies()) != len(notation.Plies()) {
			t.Fatal("round trip changed game")
		}
	})
}

func buildLegalNotation(t *testing.T, choices []byte) Notation {
	t.Helper()
	notation := NewNotation()
	player := oware.South
	for _, choice := range choices {
		moves := notation.Board().LegalMoves(player)
		if len(moves) == 0 || notation.Board().GameOver() {
			break
		}
		var err error
		notation, err = notation.Play(player, moves[int(choice)%len(moves)])
		if err != nil {
			t.Fatal(err)
		}
		player = player.Opponent()
	}
	return notation
}
