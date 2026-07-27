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

const (
	ResponseWindow = 48 * time.Hour
	ArchiveAfter   = 30 * 24 * time.Hour
)

type Status string
type Action string

const (
	StatusActive   Status = "active"
	StatusResting  Status = "resting"
	StatusArchived Status = "archived"
	ActionStarted  Action = "started"
	ActionRested   Action = "rested"
	ActionRelight  Action = "relight"
	ActionArchived Action = "archived"
)

var (
	ErrInvalid         = errors.New("invalid courtship pace")
	ErrDenied          = errors.New("courtship pace unavailable")
	ErrStaleRevision   = errors.New("stale courtship pace revision")
	ErrCommandMismatch = errors.New("pace command replay mismatch")
)

var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)

type Command struct {
	ID               string
	ActorKey         string
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
type Pace struct {
	id                     string
	members, relightGrants []string
	status                 Status
	responseAt, archiveAt  time.Time
	revision               uint64
	events                 []Event
	commands               []Applied
}
type State struct {
	ID                     string
	Members, RelightGrants []string
	Status                 Status
	ResponseAt, ArchiveAt  time.Time
	Revision               uint64
	Events                 []Event
	Commands               []Applied
}

func New(id string, members []string, commandID, actorKey string, at time.Time) (Pace, error) {
	normalized := append([]string(nil), members...)
	slices.Sort(normalized)
	normalized = slices.Compact(normalized)
	if !idPattern.MatchString(id) || len(normalized) != 2 || !keyPattern.MatchString(normalized[0]) ||
		!keyPattern.MatchString(normalized[1]) || !slices.Contains(normalized, actorKey) ||
		!idPattern.MatchString(commandID) || at.IsZero() {
		return Pace{}, ErrInvalid
	}
	pace := Pace{id: id, members: normalized, status: StatusActive, responseAt: at.UTC().Add(ResponseWindow)}
	return pace.record(ActionStarted, Command{ID: commandID, ActorKey: actorKey, At: at})
}

func Rehydrate(state State) (Pace, error) {
	members := append([]string(nil), state.Members...)
	slices.Sort(members)
	members = slices.Compact(members)
	grants := append([]string(nil), state.RelightGrants...)
	slices.Sort(grants)
	grants = slices.Compact(grants)
	pace := Pace{
		id: state.ID, members: members, relightGrants: grants, status: state.Status,
		responseAt: state.ResponseAt.UTC(), archiveAt: state.ArchiveAt.UTC(), revision: state.Revision,
		events: append([]Event(nil), state.Events...), commands: append([]Applied(nil), state.Commands...),
	}
	if !idPattern.MatchString(pace.id) || len(members) != 2 || !keyPattern.MatchString(members[0]) ||
		!keyPattern.MatchString(members[1]) || len(grants) > 2 ||
		(len(grants) > 0 && (!slices.Contains(members, grants[0]) || (len(grants) == 2 && !slices.Contains(members, grants[1])))) ||
		len(pace.events) != int(pace.revision) || len(pace.commands) != int(pace.revision) ||
		pace.status != StatusActive && pace.status != StatusResting && pace.status != StatusArchived {
		return Pace{}, ErrInvalid
	}
	if pace.status == StatusActive && (pace.responseAt.IsZero() || !pace.archiveAt.IsZero() || len(grants) != 0) {
		return Pace{}, ErrInvalid
	}
	if pace.status == StatusResting && (pace.archiveAt.IsZero() || !pace.responseAt.IsZero()) {
		return Pace{}, ErrInvalid
	}
	if pace.status == StatusArchived && (!pace.responseAt.IsZero() || !pace.archiveAt.IsZero() || len(grants) != 0) {
		return Pace{}, ErrInvalid
	}
	return pace, nil
}

// Advance performs only server-time transitions. It never exposes message-read state.
func (pace Pace) Advance(command Command) (Pace, error) {
	command.ActorKey = ""
	if replayed, err := pace.replay(command); replayed || err != nil {
		return pace, err
	}
	action := ActionRested
	switch {
	case pace.status == StatusActive && !command.At.Before(pace.responseAt):
		pace.status = StatusResting
		pace.archiveAt = pace.responseAt.Add(ArchiveAfter)
		pace.responseAt = time.Time{}
	case pace.status == StatusResting && !command.At.Before(pace.archiveAt):
		action = ActionArchived
		pace.status = StatusArchived
		pace.archiveAt = time.Time{}
		pace.relightGrants = nil
	default:
		return Pace{}, ErrDenied
	}
	return pace.record(action, command)
}

func (pace Pace) Relight(command Command) (Pace, error) {
	if replayed, err := pace.replay(command); replayed || err != nil {
		return pace, err
	}
	if pace.status != StatusResting || !slices.Contains(pace.members, command.ActorKey) ||
		command.At.IsZero() || !command.At.Before(pace.archiveAt) {
		return Pace{}, ErrDenied
	}
	if slices.Contains(pace.relightGrants, command.ActorKey) {
		return pace.record(ActionRelight, command)
	}
	pace.relightGrants = append(append([]string(nil), pace.relightGrants...), command.ActorKey)
	slices.Sort(pace.relightGrants)
	if len(pace.relightGrants) == 2 {
		pace.status = StatusActive
		pace.relightGrants = nil
		pace.responseAt = command.At.UTC().Add(ResponseWindow)
		pace.archiveAt = time.Time{}
	}
	return pace.record(ActionRelight, command)
}

func (pace Pace) replay(command Command) (bool, error) {
	for _, seen := range pace.commands {
		if seen.ID != command.ID {
			continue
		}
		for _, event := range pace.events {
			if event.CommandID == command.ID {
				if seen.Fingerprint != fingerprint(pace.id, event.Action, command) {
					return true, ErrCommandMismatch
				}
				return true, nil
			}
		}
		return true, ErrInvalid
	}
	return false, nil
}

func (pace Pace) record(action Action, command Command) (Pace, error) {
	if !idPattern.MatchString(command.ID) || command.At.IsZero() {
		return Pace{}, ErrDenied
	}
	fp := fingerprint(pace.id, action, command)
	for _, seen := range pace.commands {
		if seen.ID == command.ID {
			if seen.Fingerprint != fp {
				return Pace{}, ErrCommandMismatch
			}
			return pace, nil
		}
	}
	if command.ExpectedRevision != pace.revision {
		return Pace{}, ErrStaleRevision
	}
	pace.revision++
	pace.events = append(append([]Event(nil), pace.events...), Event{pace.revision, command.ID, command.ActorKey, action, command.At.UTC()})
	pace.commands = append(append([]Applied(nil), pace.commands...), Applied{command.ID, fp, pace.revision})
	return pace, nil
}

func fingerprint(id string, action Action, command Command) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + command.ID + "\x00" +
		command.ActorKey + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10) + "\x00" +
		command.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(sum[:])
}

func (pace Pace) ID() string              { return pace.id }
func (pace Pace) Members() []string       { return append([]string(nil), pace.members...) }
func (pace Pace) RelightGrants() []string { return append([]string(nil), pace.relightGrants...) }
func (pace Pace) Status() Status          { return pace.status }
func (pace Pace) ResponseAt() time.Time   { return pace.responseAt }
func (pace Pace) ArchiveAt() time.Time    { return pace.archiveAt }
func (pace Pace) Revision() uint64        { return pace.revision }
func (pace Pace) Events() []Event         { return append([]Event(nil), pace.events...) }
func (pace Pace) Commands() []Applied     { return append([]Applied(nil), pace.commands...) }
