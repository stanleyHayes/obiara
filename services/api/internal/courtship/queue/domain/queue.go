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
)
var opaqueKey = regexp.MustCompile(`^[a-f0-9]{64}$`)
var opaqueID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type State struct {
	RoomKey            string
	Sequence, Revision uint64
}

func Open(roomKey string) (State, error) {
	if !opaqueKey.MatchString(roomKey) {
		return State{}, ErrInvalid
	}
	return State{RoomKey: roomKey}, nil
}
func Rehydrate(roomKey string, sequence, revision uint64) (State, error) {
	if !opaqueKey.MatchString(roomKey) || sequence != revision {
		return State{}, ErrInvalid
	}
	return State{roomKey, sequence, revision}, nil
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
	next := s
	next.Sequence++
	next.Revision++
	event := Event{s.RoomKey, next.Sequence, c.ID, c.DeviceKey, c.ActorKey, c.PayloadKey, c.Fingerprint, c.At.UTC()}
	return next, event, nil
}
