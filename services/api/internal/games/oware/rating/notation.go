package rating

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	oware "github.com/stanleyHayes/obiara/services/api/internal/games/oware/domain"
)

const (
	MaxPlies         = 512
	MaxNotationBytes = 64 * 1024
	notationHeader   = "OWARE-ABAPA/1"
)

var (
	ErrNotationInvalid = errors.New("invalid Oware notation")
	ErrNotationBound   = errors.New("Oware notation exceeds bound")
	ErrNotationReplay  = errors.New("Oware notation does not replay")
)

type Ply struct {
	Number       int
	Player       oware.Player
	Pit          int
	Captured     int
	PositionHash string
}

// Notation is an immutable, bounded transcript of legal domain moves.
type Notation struct {
	board oware.Board
	plies []Ply
}

func NewNotation() Notation {
	return Notation{board: oware.NewBoard()}
}

func (notation Notation) Board() oware.Board { return notation.board }

func (notation Notation) Plies() []Ply {
	return append([]Ply(nil), notation.plies...)
}

// Play applies the move through domain.Board.ApplyMove and records its
// resulting capture count and position digest.
func (notation Notation) Play(player oware.Player, pit int) (Notation, error) {
	if len(notation.plies) >= MaxPlies {
		return notation, ErrNotationBound
	}
	expected := oware.South
	if len(notation.plies)%2 == 1 {
		expected = oware.North
	}
	if player != expected || notation.board.GameOver() {
		return notation, ErrNotationInvalid
	}
	move, err := notation.board.ApplyMove(player, pit)
	if err != nil {
		return notation, err
	}
	ply := Ply{
		Number:       len(notation.plies) + 1,
		Player:       player,
		Pit:          pit,
		Captured:     move.Captured,
		PositionHash: positionHash(move.Board),
	}
	return Notation{
		board: move.Board,
		plies: append(append([]Ply(nil), notation.plies...), ply),
	}, nil
}

func (notation Notation) MarshalText() ([]byte, error) {
	var builder strings.Builder
	builder.WriteString(notationHeader)
	builder.WriteByte('\n')
	for _, ply := range notation.plies {
		player := "S"
		if ply.Player == oware.North {
			player = "N"
		}
		fmt.Fprintf(&builder, "%d %s %d %d %s\n", ply.Number, player, ply.Pit, ply.Captured, ply.PositionHash)
		if builder.Len() > MaxNotationBytes {
			return nil, ErrNotationBound
		}
	}
	return []byte(builder.String()), nil
}

// ParseNotation strictly parses and replays every ply. Captures and position
// hashes are evidence, never trusted inputs.
func ParseNotation(encoded []byte) (Notation, error) {
	if len(encoded) > MaxNotationBytes {
		return Notation{}, ErrNotationBound
	}
	scanner := bufio.NewScanner(strings.NewReader(string(encoded)))
	if !scanner.Scan() || scanner.Text() != notationHeader {
		return Notation{}, ErrNotationInvalid
	}
	notation := NewNotation()
	for scanner.Scan() {
		if len(notation.plies) >= MaxPlies {
			return Notation{}, ErrNotationBound
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) != 5 {
			return Notation{}, ErrNotationInvalid
		}
		number, err := strconv.Atoi(fields[0])
		if err != nil || number != len(notation.plies)+1 {
			return Notation{}, ErrNotationInvalid
		}
		var player oware.Player
		switch fields[1] {
		case "S":
			player = oware.South
		case "N":
			player = oware.North
		default:
			return Notation{}, ErrNotationInvalid
		}
		pit, pitErr := strconv.Atoi(fields[2])
		captured, capturedErr := strconv.Atoi(fields[3])
		if pitErr != nil || capturedErr != nil || len(fields[4]) != sha256.Size*2 {
			return Notation{}, ErrNotationInvalid
		}
		next, playErr := notation.Play(player, pit)
		if playErr != nil {
			return Notation{}, errors.Join(ErrNotationReplay, playErr)
		}
		played := next.plies[len(next.plies)-1]
		if played.Captured != captured || played.PositionHash != fields[4] {
			return Notation{}, ErrNotationReplay
		}
		notation = next
	}
	if err := scanner.Err(); err != nil {
		return Notation{}, ErrNotationInvalid
	}
	return notation, nil
}

func positionHash(board oware.Board) string {
	var payload [72]byte
	offset := 0
	for _, seeds := range board.Houses() {
		binary.BigEndian.PutUint32(payload[offset:], uint32(seeds))
		offset += 4
	}
	for _, captured := range board.Captured() {
		binary.BigEndian.PutUint32(payload[offset:], uint32(captured))
		offset += 4
	}
	if board.GameOver() {
		payload[offset] = 1
	}
	offset++
	binary.BigEndian.PutUint32(payload[offset:], uint32(board.Winner()+1))
	sum := sha256.Sum256(payload[:offset+4])
	return hex.EncodeToString(sum[:])
}
