package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	oware "github.com/stanleyHayes/obiara/services/api/internal/games/oware/domain"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusExpired   Status = "expired"
)
const (
	MinMoveWindow = time.Minute
	MaxMoveWindow = 24 * time.Hour
)

var (
	ErrInvalid    = errors.New("invalid oware session")
	ErrNotTurn    = errors.New("not player's turn")
	ErrExpired    = errors.New("oware session expired")
	ErrTransition = errors.New("invalid oware session transition")
	ErrStale      = errors.New("stale oware session")
	ErrMismatch   = errors.New("oware session command mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence  uint64    `bson:"sequence"`
	CommandID string    `bson:"commandId"`
	Action    string    `bson:"action"`
	ActorKey  string    `bson:"actorKey,omitempty"`
	Pit       int       `bson:"pit"`
	At        time.Time `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type Projection struct {
	ID, RoomRef              string
	Players                  []string
	Turn                     oware.Player
	YourPlayer               oware.Player
	Houses                   [12]int
	Captured                 [2]int
	Status                   Status
	Winner                   int
	Revision                 uint64
	MoveDeadline, ServerTime time.Time
}
type Session struct {
	id, roomRef string
	players     []string
	board       oware.Board
	turn        oware.Player
	moveWindow  time.Duration
	deadline    time.Time
	status      Status
	revision    uint64
	events      []Event
	commands    []Applied
}
type State struct {
	ID, RoomRef string
	Players     []string
	Houses      [12]int
	Captured    [2]int
	GameOver    bool
	Winner      int
	Turn        oware.Player
	MoveWindow  time.Duration
	Deadline    time.Time
	Status      Status
	Revision    uint64
	Events      []Event
	Commands    []Applied
}

func Create(id, room string, players []string, window time.Duration, now time.Time, c Command) (Session, error) {
	p := append([]string(nil), players...)
	slices.Sort(p)
	s := Session{id: id, roomRef: room, players: p, board: oware.NewBoard(), turn: oware.South, moveWindow: window, deadline: now.UTC().Add(window), status: StatusActive}
	if !opaque.MatchString(id) || !key.MatchString(room) || !validPlayers(p) || window < MinMoveWindow || window > MaxMoveWindow || now.IsZero() || c.ExpectedRevision != 0 {
		return Session{}, ErrInvalid
	}
	if e := s.apply(c, "create", "", -1); e != nil {
		return Session{}, e
	}
	return s, nil
}
func Rehydrate(st State) (Session, error) {
	s := Session{id: st.ID, roomRef: st.RoomRef, players: append([]string(nil), st.Players...), board: oware.ReconstituteBoard(st.Houses, st.Captured, st.GameOver, st.Winner), turn: st.Turn, moveWindow: st.MoveWindow, deadline: st.Deadline.UTC(), status: st.Status, revision: st.Revision, events: append([]Event(nil), st.Events...), commands: append([]Applied(nil), st.Commands...)}
	if !opaque.MatchString(s.id) || !key.MatchString(s.roomRef) || !validPlayers(s.players) || !slices.IsSorted(s.players) || s.moveWindow < MinMoveWindow || s.moveWindow > MaxMoveWindow || s.deadline.IsZero() || len(s.events) != int(s.revision) || len(s.commands) != int(s.revision) || s.revision == 0 {
		return Session{}, ErrInvalid
	}
	return s, nil
}
func (s Session) Move(actor string, pit int, now time.Time, c Command) (Session, error) {
	if replay, e := s.replay(c, "move", actor, pit); replay || e != nil {
		return s, e
	}
	if s.status != StatusActive {
		return Session{}, ErrTransition
	}
	if !now.Before(s.deadline) {
		return Session{}, ErrExpired
	}
	expected := s.players[int(s.turn)]
	if actor != expected {
		return Session{}, ErrNotTurn
	}
	move, e := s.board.ApplyMove(s.turn, pit)
	if e != nil {
		return Session{}, e
	}
	s.board = move.Board
	if s.board.GameOver() {
		s.status = StatusCompleted
	} else {
		s.turn = s.turn.Opponent()
		s.deadline = now.UTC().Add(s.moveWindow)
	}
	return s, s.apply(c, "move", actor, pit)
}
func (s Session) Expire(now time.Time, c Command) (Session, error) {
	if replay, e := s.replay(c, "expire", "", -1); replay || e != nil {
		return s, e
	}
	if s.status != StatusActive || now.Before(s.deadline) {
		return Session{}, ErrTransition
	}
	s.status = StatusExpired
	return s, s.apply(c, "expire", "", -1)
}
func (s Session) Project(now time.Time) Projection {
	return Projection{ID: s.id, RoomRef: s.roomRef, Players: append([]string(nil), s.players...), Turn: s.turn, Houses: s.board.Houses(), Captured: s.board.Captured(), Status: s.status, Winner: s.board.Winner(), Revision: s.revision, MoveDeadline: s.deadline, ServerTime: now.UTC()}
}
func (s *Session) apply(c Command, a, actor string, pit int) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != s.revision {
		return ErrStale
	}
	f := fingerprint(s.id, c, a, actor, pit)
	s.revision++
	s.events = append(s.events, Event{s.revision, c.ID, a, actor, pit, c.At.UTC()})
	s.commands = append(s.commands, Applied{c.ID, f, s.revision})
	return nil
}
func (s Session) replay(c Command, a, actor string, pit int) (bool, error) {
	f := fingerprint(s.id, c, a, actor, pit)
	for _, x := range s.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func validPlayers(p []string) bool {
	return len(p) == 2 && p[0] != p[1] && key.MatchString(p[0]) && key.MatchString(p[1])
}
func fingerprint(id string, c Command, a, actor string, pit int) string {
	x := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + actor + "\x00" + strconv.Itoa(pit) + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(x[:])
}
func (s Session) ID() string                { return s.id }
func (s Session) RoomRef() string           { return s.roomRef }
func (s Session) Players() []string         { return append([]string(nil), s.players...) }
func (s Session) Board() oware.Board        { return s.board }
func (s Session) Turn() oware.Player        { return s.turn }
func (s Session) MoveWindow() time.Duration { return s.moveWindow }
func (s Session) Deadline() time.Time       { return s.deadline }
func (s Session) Status() Status            { return s.status }
func (s Session) Revision() uint64          { return s.revision }
func (s Session) Events() []Event           { return append([]Event(nil), s.events...) }
func (s Session) Commands() []Applied       { return append([]Applied(nil), s.commands...) }
