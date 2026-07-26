package domain

import (
	"math/rand"
	"testing"
)

func TestOpeningPosition(t *testing.T) {
	board := NewBoard()
	if board.TotalSeeds() != totalSeeds {
		t.Fatalf("total = %d, want 48", board.TotalSeeds())
	}
	if board.GameOver() || board.Winner() != -1 {
		t.Fatal("opening must be in play")
	}
	if moves := board.LegalMoves(South); len(moves) != perSide {
		t.Fatalf("opening legal moves = %v, want all 6", moves)
	}
}

func TestValidateMoveBasics(t *testing.T) {
	board := NewBoard()
	if err := board.ValidateMove(South, -1); err != ErrInvalidPit {
		t.Fatalf("negative pit = %v", err)
	}
	if err := board.ValidateMove(South, 6); err != ErrInvalidPit {
		t.Fatalf("pit 6 = %v", err)
	}
	empty := ReconstituteBoard([12]int{0, 0, 0, 0, 0, 0, 4, 4, 4, 4, 4, 4}, [2]int{0, 0}, false, -1)
	if err := empty.ValidateMove(South, 0); err != ErrEmptyPit {
		t.Fatalf("empty pit = %v", err)
	}
	over := ReconstituteBoard([12]int{}, [2]int{25, 23}, true, 0)
	if err := over.ValidateMove(South, 0); err != ErrGameOver {
		t.Fatalf("game over = %v", err)
	}
}

func TestBasicSow(t *testing.T) {
	board := NewBoard()
	move, err := board.ApplyMove(South, 2)
	if err != nil {
		t.Fatal(err)
	}
	houses := move.Board.Houses()
	if houses[2] != 0 || houses[3] != 5 || houses[4] != 5 || houses[5] != 5 || houses[6] != 5 {
		t.Fatalf("houses after sow = %v", houses)
	}
	if move.Captured != 0 {
		t.Fatalf("captured = %d, want 0", move.Captured)
	}
	if move.Board.TotalSeeds() != totalSeeds {
		t.Fatal("seed conservation violated")
	}
}

func TestCaptureChain(t *testing.T) {
	// South plays pit 5 (4 seeds): sows into North houses 6,7,8,9 (indices
	// 6..9 map from North pits 5,4,3,2). North houses 6..9 hold 2,3,1,2
	// seeds respectively after placement: 6→2, 7→3, 8→1, 9→2? Set up so
	// the landing chain captures.
	houses := [12]int{0, 0, 0, 0, 0, 4, 1, 2, 0, 1, 4, 4}
	board := ReconstituteBoard(houses, [2]int{0, 0}, false, -1)
	move, err := board.ApplyMove(South, 5)
	if err != nil {
		t.Fatal(err)
	}
	// After sow: house6=2, house7=3, house8=1, house9=2. Landing at 9:
	// chain walks 9 (2 ✓), 8 (1 ✗ stop). Capture = 2.
	if move.Captured != 2 {
		t.Fatalf("captured = %d, want 2", move.Captured)
	}
	if move.Board.Houses()[9] != 0 || move.Board.Houses()[7] != 3 {
		t.Fatalf("houses = %v", move.Board.Houses())
	}
	if move.Board.TotalSeeds() != 16 {
		t.Fatalf("total = %d, want 16 (captured seeds stay in the count)", move.Board.TotalSeeds())
	}
}

func TestGrandSlamCapturesNothing(t *testing.T) {
	// South's sow would land on a full chain of 2s across North's side:
	// capturing all would empty North — forbidden, so nothing is captured.
	houses := [12]int{0, 0, 0, 0, 0, 6, 1, 1, 1, 1, 1, 1}
	board := ReconstituteBoard(houses, [2]int{0, 0}, false, -1)
	move, err := board.ApplyMove(South, 5)
	if err != nil {
		t.Fatal(err)
	}
	// After sow: North houses 6..11 become 2,2,2,2,2,2 — all capturable,
	// which would be a grand slam. Expect zero capture.
	if move.Captured != 0 {
		t.Fatalf("grand slam captured = %d, want 0", move.Captured)
	}
	northHouses := move.Board.Houses()
	for _, seeds := range northHouses[6:12] {
		if seeds != 2 {
			t.Fatalf("north houses = %v", northHouses)
		}
	}
}

func TestFeedRuleEnforced(t *testing.T) {
	// North is starving; South must play a move that reaches North's side.
	houses := [12]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	board := ReconstituteBoard(houses, [2]int{20, 20}, false, -1)
	// Pit 0 holds 1 seed: sows to house 1 (South) — does not feed.
	if err := board.ValidateMove(South, 0); err != ErrMustFeed {
		t.Fatalf("non-feeding move = %v, want ErrMustFeed", err)
	}
	// Give South a house that reaches North.
	houses[0] = 7
	board = ReconstituteBoard(houses, [2]int{20, 20}, false, -1)
	if err := board.ValidateMove(South, 0); err != nil {
		t.Fatalf("feeding move = %v", err)
	}
}

func TestWinByThreshold(t *testing.T) {
	// South pit 5 holds 1 seed: sows to house 6 making it 3 → capture.
	// House 7 holds 1 uncapturable seed, so this is not a grand slam.
	houses := [12]int{0, 0, 0, 0, 0, 1, 2, 1, 0, 0, 0, 0}
	board := ReconstituteBoard(houses, [2]int{24, 0}, false, -1)
	move, err := board.ApplyMove(South, 5)
	if err != nil {
		t.Fatal(err)
	}
	// 24 + 3 = 27 >= 25 → South wins, leftovers scored to their owners.
	if !move.Board.GameOver() || move.Board.Winner() != 0 {
		t.Fatalf("game = over %v winner %d", move.Board.GameOver(), move.Board.Winner())
	}
	if move.Captured != 3 {
		t.Fatalf("captured = %d, want 3", move.Captured)
	}
	if move.Board.TotalSeeds() != 28 {
		t.Fatalf("total = %d, want 28 (conserved from the synthetic position)", move.Board.TotalSeeds())
	}
	if move.Board.Captured()[0] != 27 {
		t.Fatalf("captured = %v, want 27", move.Board.Captured())
	}
}

// TestSeedConservationAndLegalAlternation is the FR-501 property suite:
// any sequence of legal alternating moves conserves exactly 48 seeds,
// keeps house counts non-negative, and never leaves a finished game
// without a winner decision.
func TestSeedConservationAndLegalAlternation(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 100; trial++ {
		board := NewBoard()
		player := South
		for ply := 0; ply < 200 && !board.GameOver(); ply++ {
			moves := board.LegalMoves(player)
			if len(moves) == 0 {
				break
			}
			pit := moves[rng.Intn(len(moves))]
			move, err := board.ApplyMove(player, pit)
			if err != nil {
				t.Fatalf("trial %d ply %d: legal move rejected: %v", trial, ply, err)
			}
			board = move.Board
			if board.TotalSeeds() != totalSeeds {
				t.Fatalf("trial %d ply %d: total = %d", trial, ply, board.TotalSeeds())
			}
			for _, seeds := range board.Houses() {
				if seeds < 0 {
					t.Fatalf("trial %d ply %d: negative house", trial, ply)
				}
			}
			player = player.Opponent()
		}
		if board.GameOver() && board.Winner() == -1 {
			t.Fatalf("trial %d: game over without winner decision", trial)
		}
	}
}

// TestGrandSlamProperty: no legal move ever empties the opponent via capture.
func TestGrandSlamProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for trial := 0; trial < 200; trial++ {
		var houses [12]int
		for i := range houses {
			houses[i] = rng.Intn(9)
		}
		board := ReconstituteBoard(houses, [2]int{0, 0}, false, -1)
		for _, player := range []Player{South, North} {
			for pit := 0; pit < perSide; pit++ {
				if err := board.ValidateMove(player, pit); err != nil {
					continue
				}
				move, err := board.ApplyMove(player, pit)
				if err != nil {
					continue
				}
				if move.Captured > 0 && move.Board.sideTotal(player.Opponent()) == 0 && !move.Board.GameOver() {
					t.Fatalf("trial %d: capture emptied opponent (grand slam) captured=%d houses=%v pit=%d",
						trial, move.Captured, houses, pit)
				}
			}
		}
	}
}
