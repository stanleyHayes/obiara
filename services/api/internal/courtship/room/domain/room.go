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

type Kind string
type Status string

const (
	KindOpened   Kind   = "room_opened"
	KindMessage  Kind   = "message_added"
	KindClosed   Kind   = "room_closed"
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

var (
	ErrInvalidRoom     = errors.New("invalid courtship room")
	ErrNotMember       = errors.New("courtship room member denied")
	ErrClosed          = errors.New("courtship room closed")
	ErrStaleRevision   = errors.New("stale courtship room revision")
	ErrCommandMismatch = errors.New("courtship room command replay mismatch")
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
	Kind                            Kind
	ContentKey                      string
	At                              time.Time
}
type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}
type Projection struct {
	Status         Status
	MessageCount   uint64
	LastActivityAt time.Time
}
type Room struct {
	id         string
	members    []string
	events     []Event
	commands   []AppliedCommand
	projection Projection
}
type State struct {
	ID       string
	Members  []string
	Events   []Event
	Commands []AppliedCommand
}

func Open(id string, members []string, c Command) (Room, error) {
	ms, ok := pair(members)
	if !opaque.MatchString(id) || !ok || c.ExpectedRevision != 0 || !slices.Contains(ms, c.ActorKey) {
		return Room{}, ErrInvalidRoom
	}
	r := Room{id: id, members: ms}
	return r.append(KindOpened, "", c)
}
func (r Room) Message(content string, c Command) (Room, error) {
	if !key.MatchString(content) {
		return Room{}, ErrInvalidRoom
	}
	return r.change(KindMessage, content, c)
}
func (r Room) Close(c Command) (Room, error) { return r.change(KindClosed, "", c) }
func (r Room) change(kind Kind, content string, c Command) (Room, error) {
	if replay, e := r.replay(kind, content, c); replay || e != nil {
		return r, e
	}
	if !r.IsMember(c.ActorKey) {
		return Room{}, ErrNotMember
	}
	if r.projection.Status != StatusOpen {
		return Room{}, ErrClosed
	}
	return r.append(kind, content, c)
}
func (r Room) append(kind Kind, content string, c Command) (Room, error) {
	if !valid(c) {
		return Room{}, ErrInvalidRoom
	}
	if c.ExpectedRevision != uint64(len(r.events)) {
		return Room{}, ErrStaleRevision
	}
	seq := uint64(len(r.events) + 1)
	e := Event{Sequence: seq, CommandID: c.ID, ActorKey: c.ActorKey, ReasonCode: c.ReasonCode, Kind: kind, ContentKey: content, At: c.At.UTC()}
	r.events = append(r.events, e)
	r.commands = append(r.commands, AppliedCommand{ID: c.ID, Fingerprint: fingerprint(r.id, kind, content, c), Revision: seq})
	p, x := Project(r.events)
	if x != nil {
		return Room{}, x
	}
	r.projection = p
	return r, nil
}
func Rehydrate(s State) (Room, error) {
	ms, ok := pair(s.Members)
	if !ok || !opaque.MatchString(s.ID) || len(s.Events) == 0 || len(s.Events) != len(s.Commands) {
		return Room{}, ErrInvalidRoom
	}
	seen := map[string]bool{}
	for i, e := range s.Events {
		a := s.Commands[i]
		c := Command{ID: e.CommandID, ActorKey: e.ActorKey, ReasonCode: e.ReasonCode, ExpectedRevision: uint64(i), At: e.At}
		if e.Sequence != uint64(i+1) || a.ID != e.CommandID || a.Revision != e.Sequence || seen[a.ID] || !valid(c) || !slices.Contains(ms, e.ActorKey) || a.Fingerprint != fingerprint(s.ID, e.Kind, e.ContentKey, c) {
			return Room{}, ErrInvalidRoom
		}
		seen[a.ID] = true
	}
	p, e := Project(s.Events)
	if e != nil {
		return Room{}, e
	}
	return Room{id: s.ID, members: ms, events: append([]Event(nil), s.Events...), commands: append([]AppliedCommand(nil), s.Commands...), projection: p}, nil
}
func Project(events []Event) (Projection, error) {
	p := Projection{}
	for i, e := range events {
		if e.Sequence != uint64(i+1) || i == 0 && e.Kind != KindOpened || i > 0 && p.Status != StatusOpen {
			return Projection{}, ErrInvalidRoom
		}
		switch e.Kind {
		case KindOpened:
			p.Status = StatusOpen
		case KindMessage:
			if !key.MatchString(e.ContentKey) {
				return Projection{}, ErrInvalidRoom
			}
			p.MessageCount++
		case KindClosed:
			p.Status = StatusClosed
		default:
			return Projection{}, ErrInvalidRoom
		}
		p.LastActivityAt = e.At.UTC()
	}
	return p, nil
}
func (r Room) replay(k Kind, content string, c Command) (bool, error) {
	f := fingerprint(r.id, k, content, c)
	for _, x := range r.commands {
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
	y := append([]string(nil), x...)
	slices.Sort(y)
	return y, true
}
func valid(c Command) bool {
	return opaque.MatchString(c.ID) && key.MatchString(c.ActorKey) && reason.MatchString(c.ReasonCode) && !c.At.IsZero()
}
func fingerprint(id string, k Kind, content string, c Command) string {
	s := sha256.Sum256([]byte(id + "\x00" + string(k) + "\x00" + content + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.ReasonCode + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(s[:])
}
func (r Room) ID() string                 { return r.id }
func (r Room) Members() []string          { return append([]string(nil), r.members...) }
func (r Room) Events() []Event            { return append([]Event(nil), r.events...) }
func (r Room) Commands() []AppliedCommand { return append([]AppliedCommand(nil), r.commands...) }
func (r Room) Projection() Projection     { return r.projection }
func (r Room) Revision() uint64           { return uint64(len(r.events)) }
func (r Room) IsMember(k string) bool     { return slices.Contains(r.members, k) }
func (r Room) HasCommand(id string) bool {
	for _, c := range r.commands {
		if c.ID == id {
			return true
		}
	}
	return false
}
