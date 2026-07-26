package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type Status string

const (
	StatusAwaiting    Status = "awaiting_mutual"
	StatusRoomCreated Status = "private_room_created"
)

var (
	ErrInvalidWater    = errors.New("invalid mutual water")
	ErrNotMember       = errors.New("water actor is not a member")
	ErrAlreadyWatered  = errors.New("member already watered")
	ErrStaleRevision   = errors.New("stale mutual water revision")
	ErrCommandMismatch = errors.New("mutual water command replay mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)
var reason = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)

type Command struct {
	ID, ActorKey, ReasonCode string
	ExpectedRevision         uint64
	At                       time.Time
}
type Event struct {
	Sequence                        uint64
	CommandID, ActorKey, ReasonCode string
	Mutual                          bool
	RoomKey                         string
	At                              time.Time
}
type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}
type Water struct {
	id               string
	members, watered []string
	roomKey          string
	status           Status
	revision         uint64
	events           []Event
	commands         []AppliedCommand
}
type State struct {
	ID               string
	Members, Watered []string
	RoomKey          string
	Status           Status
	Revision         uint64
	Events           []Event
	Commands         []AppliedCommand
}

func Start(id string, members []string, c Command) (Water, error) {
	ms, ok := pair(members)
	if !opaque.MatchString(id) || !ok || c.ExpectedRevision != 0 || c.ActorKey != ms[0] && c.ActorKey != ms[1] {
		return Water{}, ErrInvalidWater
	}
	w := Water{id: id, members: ms, status: StatusAwaiting}
	return w.apply(c, "")
}
func (w Water) Water(c Command, roomKey string) (Water, error) {
	if replay, e := w.replay(c, roomKey); replay || e != nil {
		return w, e
	}
	if !w.IsMember(c.ActorKey) {
		return Water{}, ErrNotMember
	}
	if slices.Contains(w.watered, c.ActorKey) {
		return Water{}, ErrAlreadyWatered
	}
	if len(w.watered) == 1 && !key.MatchString(roomKey) {
		return Water{}, ErrInvalidWater
	}
	if len(w.watered) == 0 && roomKey != "" {
		return Water{}, ErrInvalidWater
	}
	return w.apply(c, roomKey)
}
func (w Water) apply(c Command, room string) (Water, error) {
	if !valid(c) {
		return Water{}, ErrInvalidWater
	}
	if c.ExpectedRevision != w.revision {
		return Water{}, ErrStaleRevision
	}
	w.revision++
	w.watered = append(w.watered, c.ActorKey)
	slices.Sort(w.watered)
	mutual := len(w.watered) == 2
	if mutual {
		w.roomKey = room
		w.status = StatusRoomCreated
	}
	e := Event{Sequence: w.revision, CommandID: c.ID, ActorKey: c.ActorKey, ReasonCode: c.ReasonCode, Mutual: mutual, RoomKey: room, At: c.At.UTC()}
	w.events = append(w.events, e)
	w.commands = append(w.commands, AppliedCommand{ID: c.ID, Fingerprint: fingerprint(w.id, c, room), Revision: w.revision})
	return w, nil
}
func Rehydrate(s State) (Water, error) {
	ms, ok := pair(s.Members)
	ws := append([]string(nil), s.Watered...)
	slices.Sort(ws)
	w := Water{id: s.ID, members: ms, watered: ws, roomKey: s.RoomKey, status: s.Status, revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]AppliedCommand(nil), s.Commands...)}
	if !ok || !opaque.MatchString(w.id) || len(ws) < 1 || len(ws) > 2 || len(w.events) != int(w.revision) || len(w.commands) != int(w.revision) {
		return Water{}, ErrInvalidWater
	}
	if len(ws) == 2 && (w.status != StatusRoomCreated || !key.MatchString(w.roomKey)) || len(ws) == 1 && (w.status != StatusAwaiting || w.roomKey != "") {
		return Water{}, ErrInvalidWater
	}
	seen := map[string]bool{}
	for i, e := range w.events {
		a := w.commands[i]
		if e.Sequence != uint64(i+1) || a.ID != e.CommandID || a.Revision != e.Sequence || seen[a.ID] || !valid(Command{ID: e.CommandID, ActorKey: e.ActorKey, ReasonCode: e.ReasonCode, ExpectedRevision: uint64(i), At: e.At}) || a.Fingerprint != fingerprint(w.id, Command{ID: e.CommandID, ActorKey: e.ActorKey, ReasonCode: e.ReasonCode, ExpectedRevision: uint64(i), At: e.At}, e.RoomKey) {
			return Water{}, ErrInvalidWater
		}
		seen[a.ID] = true
	}
	return w, nil
}
func (w Water) replay(c Command, room string) (bool, error) {
	f := fingerprint(w.id, c, room)
	for _, x := range w.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrCommandMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func pair(x []string) ([]string, bool) {
	if len(x) != 2 || !key.MatchString(x[0]) || !key.MatchString(x[1]) || x[0] == x[1] {
		return nil, false
	}
	out := append([]string(nil), x...)
	slices.Sort(out)
	return out, true
}
func valid(c Command) bool {
	return opaque.MatchString(c.ID) && key.MatchString(c.ActorKey) && reason.MatchString(c.ReasonCode) && !c.At.IsZero()
}
func fingerprint(id string, c Command, room string) string {
	s := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.ReasonCode + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + room))
	return hex.EncodeToString(s[:])
}
func (w Water) ID() string                 { return w.id }
func (w Water) Members() []string          { return append([]string(nil), w.members...) }
func (w Water) Watered() []string          { return append([]string(nil), w.watered...) }
func (w Water) RoomKey() string            { return w.roomKey }
func (w Water) Status() Status             { return w.status }
func (w Water) Revision() uint64           { return w.revision }
func (w Water) Events() []Event            { return append([]Event(nil), w.events...) }
func (w Water) Commands() []AppliedCommand { return append([]AppliedCommand(nil), w.commands...) }
func (w Water) IsMember(k string) bool     { return slices.Contains(w.members, k) }
func (w Water) HasCommand(id string) bool {
	for _, c := range w.commands {
		if c.ID == id {
			return true
		}
	}
	return false
}
