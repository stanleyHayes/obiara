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
	ThemeOnePromptRef     = "theme-one:origin-story"
	ThemeOnePromptVersion = uint64(1)
)

type Action string

const (
	ActionOpened    Action = "opened"
	ActionSubmitted Action = "submitted"
)

var (
	ErrInvalid          = errors.New("invalid guided theme")
	ErrNotMember        = errors.New("guided theme member denied")
	ErrAlreadySubmitted = errors.New("guided theme response already submitted")
	ErrStaleRevision    = errors.New("stale guided theme revision")
	ErrCommandMismatch  = errors.New("guided theme command replay mismatch")
	keyPattern          = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	reasonPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Command struct {
	ID               string
	ActorKey         string
	ReasonCode       string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence   uint64
	CommandID  string
	ActorKey   string
	Action     Action
	ContentKey string
	ReasonCode string
	At         time.Time
}
type AppliedCommand struct {
	ID          string
	Fingerprint string
	Revision    uint64
}
type RevealedSubmission struct {
	MemberKey  string
	ContentKey string
}
type Projection struct {
	PromptRef      string
	PromptVersion  uint64
	SubmittedCount uint8
	Revealed       bool
	Submissions    []RevealedSubmission
}
type Theme struct {
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

func Open(id string, members []string, command Command) (Theme, error) {
	pair, ok := normalizePair(members)
	if !idPattern.MatchString(id) || !ok || command.ExpectedRevision != 0 ||
		!slices.Contains(pair, command.ActorKey) {
		return Theme{}, ErrInvalid
	}
	theme := Theme{id: id, members: pair}
	return theme.append(ActionOpened, "", command)
}

func (theme Theme) Submit(contentKey string, command Command) (Theme, error) {
	if replayed, err := theme.replay(ActionSubmitted, contentKey, command); replayed || err != nil {
		return theme, err
	}
	if !slices.Contains(theme.members, command.ActorKey) {
		return Theme{}, ErrNotMember
	}
	if !keyPattern.MatchString(contentKey) {
		return Theme{}, ErrInvalid
	}
	for _, event := range theme.events {
		if event.Action == ActionSubmitted && event.ActorKey == command.ActorKey {
			return Theme{}, ErrAlreadySubmitted
		}
	}
	return theme.append(ActionSubmitted, contentKey, command)
}

func (theme Theme) append(action Action, contentKey string, command Command) (Theme, error) {
	if !validCommand(command) || command.ExpectedRevision != uint64(len(theme.events)) {
		if validCommand(command) {
			return Theme{}, ErrStaleRevision
		}
		return Theme{}, ErrInvalid
	}
	if len(theme.events) > 0 && command.At.Before(theme.events[len(theme.events)-1].At) {
		return Theme{}, ErrInvalid
	}
	revision := uint64(len(theme.events) + 1)
	event := Event{
		Sequence: revision, CommandID: command.ID, ActorKey: command.ActorKey,
		Action: action, ContentKey: contentKey, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	}
	theme.events = append(append([]Event(nil), theme.events...), event)
	theme.commands = append(append([]AppliedCommand(nil), theme.commands...), AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(theme.id, action, contentKey, command), Revision: revision,
	})
	projection, err := Project(theme.members, theme.events)
	if err != nil {
		return Theme{}, err
	}
	theme.projection = projection
	return theme, nil
}

func Project(members []string, events []Event) (Projection, error) {
	pair, ok := normalizePair(members)
	if !ok || len(events) == 0 || events[0].Action != ActionOpened {
		return Projection{}, ErrInvalid
	}
	responses := map[string]string{}
	for index, event := range events {
		if event.Sequence != uint64(index+1) || !slices.Contains(pair, event.ActorKey) {
			return Projection{}, ErrInvalid
		}
		switch event.Action {
		case ActionOpened:
			if index != 0 || event.ContentKey != "" {
				return Projection{}, ErrInvalid
			}
		case ActionSubmitted:
			if !keyPattern.MatchString(event.ContentKey) || responses[event.ActorKey] != "" {
				return Projection{}, ErrInvalid
			}
			responses[event.ActorKey] = event.ContentKey
		default:
			return Projection{}, ErrInvalid
		}
	}
	projection := Projection{
		PromptRef: ThemeOnePromptRef, PromptVersion: ThemeOnePromptVersion,
		SubmittedCount: uint8(len(responses)),
	}
	if len(responses) == 2 {
		projection.Revealed = true
		for _, member := range pair {
			projection.Submissions = append(projection.Submissions, RevealedSubmission{
				MemberKey: member, ContentKey: responses[member],
			})
		}
	}
	return projection, nil
}

func Rehydrate(state State) (Theme, error) {
	members, ok := normalizePair(state.Members)
	if !idPattern.MatchString(state.ID) || !ok || len(state.Events) == 0 ||
		len(state.Events) != len(state.Commands) {
		return Theme{}, ErrInvalid
	}
	seen := map[string]struct{}{}
	for index, event := range state.Events {
		applied := state.Commands[index]
		command := Command{
			ID: event.CommandID, ActorKey: event.ActorKey, ReasonCode: event.ReasonCode,
			ExpectedRevision: uint64(index), At: event.At,
		}
		if event.Sequence != uint64(index+1) || applied.ID != event.CommandID ||
			applied.Revision != event.Sequence || !validCommand(command) ||
			applied.Fingerprint != fingerprint(state.ID, event.Action, event.ContentKey, command) {
			return Theme{}, ErrInvalid
		}
		if _, duplicate := seen[applied.ID]; duplicate {
			return Theme{}, ErrInvalid
		}
		seen[applied.ID] = struct{}{}
	}
	projection, err := Project(members, state.Events)
	if err != nil {
		return Theme{}, err
	}
	return Theme{
		id: state.ID, members: members, events: append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...), projection: projection,
	}, nil
}

func (theme Theme) replay(action Action, contentKey string, command Command) (bool, error) {
	expected := fingerprint(theme.id, action, contentKey, command)
	for _, applied := range theme.commands {
		if applied.ID != command.ID {
			continue
		}
		if applied.Fingerprint != expected {
			return false, ErrCommandMismatch
		}
		return true, nil
	}
	return false, nil
}
func normalizePair(members []string) ([]string, bool) {
	if len(members) != 2 || !keyPattern.MatchString(members[0]) ||
		!keyPattern.MatchString(members[1]) || members[0] == members[1] {
		return nil, false
	}
	pair := append([]string(nil), members...)
	slices.Sort(pair)
	return pair, true
}
func validCommand(command Command) bool {
	return idPattern.MatchString(command.ID) && keyPattern.MatchString(command.ActorKey) &&
		reasonPattern.MatchString(command.ReasonCode) && !command.At.IsZero()
}
func fingerprint(id string, action Action, contentKey string, command Command) string {
	value := id + "\x00" + string(action) + "\x00" + contentKey + "\x00" + command.ID +
		"\x00" + command.ActorKey + "\x00" + command.ReasonCode + "\x00" +
		strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (theme Theme) ID() string        { return theme.id }
func (theme Theme) Members() []string { return append([]string(nil), theme.members...) }
func (theme Theme) Events() []Event   { return append([]Event(nil), theme.events...) }
func (theme Theme) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), theme.commands...)
}
func (theme Theme) Revision() uint64 { return uint64(len(theme.events)) }
func (theme Theme) Projection() Projection {
	projection := theme.projection
	projection.Submissions = append([]RevealedSubmission(nil), projection.Submissions...)
	return projection
}
func (theme Theme) HasCommand(id string) bool {
	for _, command := range theme.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}
