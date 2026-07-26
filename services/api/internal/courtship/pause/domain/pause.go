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
type Action string

const (
	StatusOpen   Status = "open"
	StatusPaused Status = "paused"
	ActionPause  Action = "pause"
	ActionAck    Action = "acknowledge"
	ActionResume Action = "resume"
)

var (
	ErrInvalid         = errors.New("invalid pause stone")
	ErrDenied          = errors.New("pause stone unavailable")
	ErrSuspended       = errors.New("room sending suspended")
	ErrStaleRevision   = errors.New("stale pause stone revision")
	ErrCommandMismatch = errors.New("pause command replay mismatch")
)

var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)

type Command struct {
	ID, ActorKey     string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence            uint64
	CommandID, ActorKey string
	Action              Action
	At                  time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Stone struct {
	id           string
	members      []string
	status       Status
	pausedBy     string
	acknowledged []string
	revision     uint64
	events       []Event
	commands     []Applied
}
type State struct {
	ID           string
	Members      []string
	Status       Status
	PausedBy     string
	Acknowledged []string
	Revision     uint64
	Events       []Event
	Commands     []Applied
}

func New(id string, members []string) (Stone, error) {
	normalized := append([]string(nil), members...)
	slices.Sort(normalized)
	normalized = slices.Compact(normalized)
	if !idPattern.MatchString(id) || len(normalized) != 2 || !keyPattern.MatchString(normalized[0]) || !keyPattern.MatchString(normalized[1]) {
		return Stone{}, ErrInvalid
	}
	return Stone{id: id, members: normalized, status: StatusOpen}, nil
}
func Rehydrate(state State) (Stone, error) {
	stone, err := New(state.ID, state.Members)
	if err != nil {
		return Stone{}, err
	}
	stone.status = state.Status
	stone.pausedBy = state.PausedBy
	stone.acknowledged = append([]string(nil), state.Acknowledged...)
	slices.Sort(stone.acknowledged)
	stone.revision = state.Revision
	stone.events = append([]Event(nil), state.Events...)
	stone.commands = append([]Applied(nil), state.Commands...)
	if (stone.status != StatusOpen && stone.status != StatusPaused) || len(stone.events) != int(stone.revision) || len(stone.commands) != int(stone.revision) {
		return Stone{}, ErrInvalid
	}
	return stone, nil
}
func (stone Stone) Pause(command Command) (Stone, error) { return stone.apply(ActionPause, command) }
func (stone Stone) Acknowledge(command Command) (Stone, error) {
	return stone.apply(ActionAck, command)
}
func (stone Stone) Resume(command Command) (Stone, error) { return stone.apply(ActionResume, command) }
func (stone Stone) apply(action Action, command Command) (Stone, error) {
	if !stone.member(command.ActorKey) || !idPattern.MatchString(command.ID) || command.At.IsZero() {
		return Stone{}, ErrDenied
	}
	fp := fingerprint(stone.id, action, command)
	for _, seen := range stone.commands {
		if seen.ID == command.ID {
			if seen.Fingerprint != fp {
				return Stone{}, ErrCommandMismatch
			}
			return stone, nil
		}
	}
	if command.ExpectedRevision != stone.revision {
		return Stone{}, ErrStaleRevision
	}
	switch action {
	case ActionPause:
		if stone.status != StatusOpen {
			return Stone{}, ErrDenied
		}
		stone.status = StatusPaused
		stone.pausedBy = command.ActorKey
		stone.acknowledged = []string{command.ActorKey}
	case ActionAck:
		if stone.status != StatusPaused {
			return Stone{}, ErrDenied
		}
		if !slices.Contains(stone.acknowledged, command.ActorKey) {
			stone.acknowledged = append(stone.acknowledged, command.ActorKey)
			slices.Sort(stone.acknowledged)
		}
	case ActionResume:
		if stone.status != StatusPaused || len(stone.acknowledged) != 2 {
			return Stone{}, ErrDenied
		}
		stone.status = StatusOpen
		stone.pausedBy = ""
		stone.acknowledged = nil
	default:
		return Stone{}, ErrInvalid
	}
	stone.revision++
	stone.events = append(stone.events, Event{stone.revision, command.ID, command.ActorKey, action, command.At.UTC()})
	stone.commands = append(stone.commands, Applied{command.ID, fp, stone.revision})
	return stone, nil
}
func (stone Stone) CanSend(actor string) error {
	if !stone.member(actor) {
		return ErrDenied
	}
	if stone.status != StatusOpen {
		return ErrSuspended
	}
	return nil
}
func (stone Stone) member(key string) bool { return slices.Contains(stone.members, key) }
func fingerprint(id string, action Action, c Command) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(sum[:])
}
func (stone Stone) ID() string             { return stone.id }
func (stone Stone) Members() []string      { return append([]string(nil), stone.members...) }
func (stone Stone) Status() Status         { return stone.status }
func (stone Stone) PausedBy() string       { return stone.pausedBy }
func (stone Stone) Acknowledged() []string { return append([]string(nil), stone.acknowledged...) }
func (stone Stone) Revision() uint64       { return stone.revision }
func (stone Stone) Events() []Event        { return append([]Event(nil), stone.events...) }
func (stone Stone) Commands() []Applied    { return append([]Applied(nil), stone.commands...) }
