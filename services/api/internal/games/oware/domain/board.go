// Package domain implements the Oware legality engine (E10-S01, FR-501:
// every move is validated server-side). The engine is deterministic and
// transport-free: given a board and a pit, it validates and applies a
// move. Persistence, async timers and ratings are separate stories
// (E10-S02/S03).
//
// Rules implemented (abapa variant):
//   - 12 houses, 4 seeds each, 48 total. Players alternate single moves.
//   - Sowing is counterclockwise from the next house; with 12+ seeds the
//     origin house is skipped on the way round.
//   - Capture: when the last seed lands in an opponent house holding 2 or
//     3 seeds after placement, that house and every immediately preceding
//     opponent house with 2 or 3 seeds is captured.
//   - Grand-slam captures are forbidden: a move that would capture all of
//     the opponent's seeds yields no capture at all.
//   - Feed rule: if the opponent has no seeds, the mover must choose a
//     move that sows into the opponent's side when such a move exists.
//   - The game ends when a player captures more than half (25+), or when
//     the opponent cannot be fed; leftover seeds are scored to the player
//     on whose side they remain.
package domain

import "errors"

// Player is a board side.
type Player int

const (
	South Player = iota
	North
)

// Opponent returns the other side.
func (player Player) Opponent() Player {
	if player == South {
		return North
	}
	return South
}

const (
	houses        = 12
	perSide       = 6
	seedsPerHouse = 4
	totalSeeds    = 48
	winThreshold  = 25
)

var (
	ErrInvalidPit = errors.New("pit must be one of the player's six houses")
	ErrEmptyPit   = errors.New("cannot move from an empty house")
	ErrMustFeed   = errors.New("must choose a move that feeds the opponent")
	ErrGameOver   = errors.New("game is already over")
)

// Board is an immutable game position. Moves return new boards.
type Board struct {
	houses   [houses]int
	captured [2]int
	gameOver bool
	winner   int // -1 none, 0 South, 1 North, 2 draw
}

// NewBoard returns the standard opening position.
func NewBoard() Board {
	return Board{houses: [houses]int{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}, winner: -1}
}

// ReconstituteBoard rebuilds a stored position without policy checks.
func ReconstituteBoard(housesArray [houses]int, captured [2]int, gameOver bool, winner int) Board {
	return Board{houses: housesArray, captured: captured, gameOver: gameOver, winner: winner}
}

func (board Board) Houses() [houses]int { return board.houses }
func (board Board) Captured() [2]int    { return board.captured }
func (board Board) GameOver() bool      { return board.gameOver }

// Winner returns -1 (none), 0 (South), 1 (North) or 2 (draw).
func (board Board) Winner() int { return board.winner }

// TotalSeeds counts every seed in play and captured (invariant: 48).
func (board Board) TotalSeeds() int {
	total := board.captured[0] + board.captured[1]
	for _, seeds := range board.houses {
		total += seeds
	}
	return total
}

func (board Board) sideTotal(player Player) int {
	total := 0
	for pit := range perSide {
		total += board.houses[board.pitIndex(player, pit)]
	}
	return total
}

func (board Board) pitIndex(player Player, pit int) int {
	if player == South {
		return pit
	}
	return houses - 1 - pit
}

// LegalMoves lists the pits the player may move from, applying the feed
// rule.
func (board Board) LegalMoves(player Player) []int {
	if board.gameOver {
		return nil
	}
	var moves []int
	for pit := 0; pit < perSide; pit++ {
		if err := board.ValidateMove(player, pit); err == nil {
			moves = append(moves, pit)
		}
	}
	return moves
}

// ValidateMove checks a move without applying it.
func (board Board) ValidateMove(player Player, pit int) error {
	if board.gameOver {
		return ErrGameOver
	}
	if pit < 0 || pit >= perSide {
		return ErrInvalidPit
	}
	if board.houses[board.pitIndex(player, pit)] == 0 {
		return ErrEmptyPit
	}
	// Feed rule: when the opponent is empty, at least one legal move must
	// feed them, and this move must be one of those.
	if board.sideTotal(player.Opponent()) == 0 && !board.feeds(player, pit) {
		return ErrMustFeed
	}
	return nil
}

// feeds reports whether moving from pit sows at least one seed onto the
// opponent's side.
func (board Board) feeds(player Player, pit int) bool {
	seeds := board.houses[board.pitIndex(player, pit)]
	if seeds == 0 {
		return false
	}
	// Walk the sowing path and note any house on the opponent's side.
	index := board.pitIndex(player, pit)
	origin := index
	for sown := 0; sown < seeds; sown++ {
		index = (index + 1) % houses
		if index == origin {
			index = (index + 1) % houses // skip origin on long sowings
		}
		if board.onSide(player.Opponent(), index) {
			return true
		}
	}
	return false
}

func (board Board) onSide(player Player, index int) bool {
	if player == South {
		return index >= 0 && index < perSide
	}
	return index >= perSide && index < houses
}

// Move is the outcome of applying a legal move.
type Move struct {
	Board    Board
	Captured int
}

// ApplyMove validates and applies the move, returning the new position.
// Seed conservation, grand-slam suppression and end-of-game scoring are
// enforced here.
func (board Board) ApplyMove(player Player, pit int) (Move, error) {
	if err := board.ValidateMove(player, pit); err != nil {
		return Move{}, err
	}

	next := board
	origin := next.pitIndex(player, pit)
	seeds := next.houses[origin]
	next.houses[origin] = 0

	index := origin
	for sown := 0; sown < seeds; sown++ {
		index = (index + 1) % houses
		if index == origin {
			index = (index + 1) % houses
		}
		next.houses[index]++
	}

	// Capture chain, with grand-slam suppression.
	captured := 0
	var capturedIndexes []int
	if next.onSide(player.Opponent(), index) {
		for cursor := index; next.onSide(player.Opponent(), cursor); cursor-- {
			seeds := next.houses[cursor]
			if seeds == 2 || seeds == 3 {
				captured += seeds
				capturedIndexes = append(capturedIndexes, cursor)
			} else {
				break
			}
			if cursor == perSide {
				break
			}
		}
	}
	// A grand slam would leave the opponent with nothing; it captures
	// nothing instead.
	if captured > 0 && next.sideTotal(player.Opponent())-captured > 0 {
		for _, cursor := range capturedIndexes {
			next.houses[cursor] = 0
		}
		next.captured[player] += captured
	} else {
		captured = 0
	}

	next = next.scoreEnd(player)
	return Move{Board: next, Captured: captured}, nil
}

// scoreEnd applies end-of-game rules after a move.
func (board Board) scoreEnd(mover Player) Board {
	opponent := mover.Opponent()
	south := board.captured[South]
	north := board.captured[North]

	ended := south >= winThreshold || north >= winThreshold
	// The mover feeds or the opponent starves: if the opponent has no
	// seeds and cannot be fed by any legal reply, the game ends and each
	// side scores the seeds on its own side.
	if !ended && board.sideTotal(opponent) == 0 && len(board.LegalMoves(opponent)) == 0 {
		ended = true
	}
	if !ended {
		return board
	}

	board.gameOver = true
	board.captured[South] += board.sideTotal(South)
	for pit := 0; pit < perSide; pit++ {
		board.houses[board.pitIndex(South, pit)] = 0
	}
	board.captured[North] += board.sideTotal(North)
	for pit := 0; pit < perSide; pit++ {
		board.houses[board.pitIndex(North, pit)] = 0
	}

	switch {
	case board.captured[South] > board.captured[North]:
		board.winner = 0
	case board.captured[North] > board.captured[South]:
		board.winner = 1
	default:
		board.winner = 2
	}
	return board
}
