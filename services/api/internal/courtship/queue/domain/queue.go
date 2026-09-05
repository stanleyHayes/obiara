package domain

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalid         = errors.New("invalid courtship queue command")
	ErrStaleDevice     = errors.New("device queue cursor is stale")
	ErrCommandMismatch = errors.New("command id reused with different input")
	// ErrNotYourTurn refuses a second consecutive turn (FR-301, IM-029). It
	// is the rule that makes this a courtship rather than an inbox: you may
	// not send again until the other member has answered.
	ErrNotYourTurn = errors.New("the other member has not answered yet")
)
var opaqueKey = regexp.MustCompile(`^[a-f0-9]{64}$`)
var opaqueID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type State struct {
	RoomKey            string
	Sequence, Revision uint64
	// LastActorKey is whoever took the previous turn, empty in a room where
	// nobody has spoken yet. Alternation lives on the aggregate that owns the
	// log rather than in a handler above it: FR-301 asks for consecutive
	// sends to be impossible, and a check in one route is only a rule for as
	// long as that route is the only way in.
	LastActorKey string
}

func Open(roomKey string) (State, error) {
	if !opaqueKey.MatchString(roomKey) {
		return State{}, ErrInvalid
	}
	// No last actor: the first turn in a room belongs to whoever takes it.
	return State{RoomKey: roomKey}, nil
}
func Rehydrate(roomKey string, sequence, revision uint64, lastActorKey string) (State, error) {
	if !opaqueKey.MatchString(roomKey) || sequence != revision {
		return State{}, ErrInvalid
	}
	// Empty is legitimate — an unopened room, or one written before turns
	// were recorded — but anything present must be a real key, or a
	// malformed value would silently never match and disable the rule.
	if lastActorKey != "" && !opaqueKey.MatchString(lastActorKey) {
		return State{}, ErrInvalid
	}
	return State{roomKey, sequence, revision, lastActorKey}, nil
}

type Command struct {
	ID, DeviceKey, ActorKey, PayloadKey, Fingerprint string
	BaseSequence                                     uint64
	At                                               time.Time
}
type Event struct {
	RoomKey                                                 string
	Sequence                                                uint64
	CommandID, DeviceKey, ActorKey, PayloadKey, Fingerprint string
	AcceptedAt                                              time.Time
}

func (s State) Accept(c Command) (State, Event, error) {
	if !opaqueID.MatchString(c.ID) || !opaqueKey.MatchString(c.DeviceKey) || !opaqueKey.MatchString(c.ActorKey) || !opaqueKey.MatchString(c.PayloadKey) || !opaqueKey.MatchString(c.Fingerprint) || c.At.IsZero() {
		return s, Event{}, ErrInvalid
	}
	if c.BaseSequence != s.Sequence {
		return s, Event{}, ErrStaleDevice
	}
	// Checked after the cursor on purpose: a member whose partner has just
	// replied is stale rather than out of turn, and telling them to catch up
	// is both true and the thing that lets them send.
	if c.ActorKey == s.LastActorKey {
		return s, Event{}, ErrNotYourTurn
	}
	next := s
	next.Sequence++
	next.Revision++
	next.LastActorKey = c.ActorKey
	event := Event{s.RoomKey, next.Sequence, c.ID, c.DeviceKey, c.ActorKey, c.PayloadKey, c.Fingerprint, c.At.UTC()}
	return next, event, nil
}
