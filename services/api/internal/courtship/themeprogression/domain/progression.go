package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type ThemeNumber uint8
type Action string

const (
	ThemeTwo   ThemeNumber = 2
	ThemeThree ThemeNumber = 3
	ThemeFour  ThemeNumber = 4

	ActionOpened    Action = "opened"
	ActionSubmitted Action = "submitted"
)

type CatalogEntry struct {
	Number        ThemeNumber
	PromptRef     string
	PromptVersion uint64
}

var fixedCatalog = [...]CatalogEntry{
	{Number: ThemeTwo, PromptRef: "theme-two:daily-rhythms", PromptVersion: 1},
	{Number: ThemeThree, PromptRef: "theme-three:care-and-conflict", PromptVersion: 1},
	{Number: ThemeFour, PromptRef: "theme-four:shared-horizon", PromptVersion: 1},
}

// Catalog returns a defensive copy of the fixed, versioned progression catalog.
func Catalog() []CatalogEntry {
	return append([]CatalogEntry(nil), fixedCatalog[:]...)
}

var (
	ErrInvalid          = errors.New("invalid theme progression")
	ErrNotMember        = errors.New("theme progression member denied")
	ErrLocked           = errors.New("guided theme is locked")
	ErrAlreadySubmitted = errors.New("guided theme response already submitted")
	ErrStaleRevision    = errors.New("stale theme progression revision")
	ErrCommandMismatch  = errors.New("theme progression command replay mismatch")
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
	Sequence    uint64
	CommandID   string
	ActorKey    string
	Action      Action
	Theme       ThemeNumber
	ContentKey  string
	EvidenceKey string
	ReasonCode  string
	At          time.Time
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
type ThemeState struct {
	Number         ThemeNumber
	PromptRef      string
	PromptVersion  uint64
	Unlocked       bool
	SubmittedCount uint8
	Revealed       bool
	Submissions    []RevealedSubmission
}
type Projection struct {
	Themes []ThemeState
}
type Progression struct {
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

func Open(id string, members []string, themeOneEvidenceKey string, command Command) (Progression, error) {
	pair, ok := normalizePair(members)
	if !idPattern.MatchString(id) || !ok || !keyPattern.MatchString(themeOneEvidenceKey) ||
		command.ExpectedRevision != 0 || !slices.Contains(pair, command.ActorKey) {
		return Progression{}, ErrInvalid
	}
	progression := Progression{id: id, members: pair}
	return progression.append(ActionOpened, ThemeTwo, "", themeOneEvidenceKey, command)
}

func (progression Progression) Submit(theme ThemeNumber, contentKey string, command Command) (Progression, error) {
	if replayed, err := progression.replay(ActionSubmitted, theme, contentKey, "", command); replayed || err != nil {
		return progression, err
	}
	if !slices.Contains(progression.members, command.ActorKey) {
		return Progression{}, ErrNotMember
	}
	state, ok := progression.themeState(theme)
	if !ok || !state.Unlocked {
		return Progression{}, ErrLocked
	}
	if state.Revealed || !keyPattern.MatchString(contentKey) {
		return Progression{}, ErrInvalid
	}
	for _, event := range progression.events {
		if event.Action == ActionSubmitted && event.Theme == theme && event.ActorKey == command.ActorKey {
			return Progression{}, ErrAlreadySubmitted
		}
	}
	return progression.append(ActionSubmitted, theme, contentKey, "", command)
}

func (progression Progression) append(action Action, theme ThemeNumber, contentKey, evidenceKey string, command Command) (Progression, error) {
	if !validCommand(command) {
		return Progression{}, ErrInvalid
	}
	if command.ExpectedRevision != uint64(len(progression.events)) {
		return Progression{}, ErrStaleRevision
	}
	if len(progression.events) > 0 && command.At.Before(progression.events[len(progression.events)-1].At) {
		return Progression{}, ErrInvalid
	}
	revision := uint64(len(progression.events) + 1)
	event := Event{
		Sequence: revision, CommandID: command.ID, ActorKey: command.ActorKey,
		Action: action, Theme: theme, ContentKey: contentKey, EvidenceKey: evidenceKey,
		ReasonCode: command.ReasonCode, At: command.At.UTC(),
	}
	progression.events = append(append([]Event(nil), progression.events...), event)
	progression.commands = append(append([]AppliedCommand(nil), progression.commands...), AppliedCommand{
		ID:          command.ID,
		Fingerprint: fingerprint(progression.id, action, theme, contentKey, evidenceKey, command),
		Revision:    revision,
	})
	projection, err := Project(progression.members, progression.events)
	if err != nil {
		return Progression{}, err
	}
	progression.projection = projection
	return progression, nil
}

func Project(members []string, events []Event) (Projection, error) {
	pair, ok := normalizePair(members)
	if !ok || len(events) == 0 || events[0].Action != ActionOpened ||
		events[0].Theme != ThemeTwo || !keyPattern.MatchString(events[0].EvidenceKey) {
		return Projection{}, ErrInvalid
	}
	responses := map[ThemeNumber]map[string]string{
		ThemeTwo: {}, ThemeThree: {}, ThemeFour: {},
	}
	for index, event := range events {
		if event.Sequence != uint64(index+1) || !slices.Contains(pair, event.ActorKey) {
			return Projection{}, ErrInvalid
		}
		switch event.Action {
		case ActionOpened:
			if index != 0 || event.Theme != ThemeTwo || event.ContentKey != "" {
				return Projection{}, ErrInvalid
			}
		case ActionSubmitted:
			if !validTheme(event.Theme) || !keyPattern.MatchString(event.ContentKey) ||
				responses[event.Theme][event.ActorKey] != "" || !unlocked(responses, event.Theme) {
				return Projection{}, ErrInvalid
			}
			responses[event.Theme][event.ActorKey] = event.ContentKey
		default:
			return Projection{}, ErrInvalid
		}
	}
	projection := Projection{Themes: make([]ThemeState, 0, len(fixedCatalog))}
	for _, entry := range fixedCatalog {
		state := ThemeState{
			Number: entry.Number, PromptRef: entry.PromptRef, PromptVersion: entry.PromptVersion,
			Unlocked: unlocked(responses, entry.Number), SubmittedCount: uint8(len(responses[entry.Number])),
		}
		if len(responses[entry.Number]) == 2 {
			state.Revealed = true
			for _, member := range pair {
				state.Submissions = append(state.Submissions, RevealedSubmission{
					MemberKey: member, ContentKey: responses[entry.Number][member],
				})
			}
		}
		projection.Themes = append(projection.Themes, state)
	}
	return projection, nil
}

func unlocked(responses map[ThemeNumber]map[string]string, theme ThemeNumber) bool {
	switch theme {
	case ThemeTwo:
		return true
	case ThemeThree:
		return len(responses[ThemeTwo]) == 2
	case ThemeFour:
		return len(responses[ThemeThree]) == 2
	default:
		return false
	}
}

func Rehydrate(state State) (Progression, error) {
	members, ok := normalizePair(state.Members)
	if !idPattern.MatchString(state.ID) || !ok || len(state.Events) == 0 ||
		len(state.Events) != len(state.Commands) {
		return Progression{}, ErrInvalid
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
			applied.Fingerprint != fingerprint(state.ID, event.Action, event.Theme, event.ContentKey, event.EvidenceKey, command) {
			return Progression{}, ErrInvalid
		}
		if _, duplicate := seen[applied.ID]; duplicate {
			return Progression{}, ErrInvalid
		}
		seen[applied.ID] = struct{}{}
	}
	projection, err := Project(members, state.Events)
	if err != nil {
		return Progression{}, err
	}
	return Progression{
		id: state.ID, members: members, events: append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...), projection: projection,
	}, nil
}

func (progression Progression) replay(action Action, theme ThemeNumber, contentKey, evidenceKey string, command Command) (bool, error) {
	expected := fingerprint(progression.id, action, theme, contentKey, evidenceKey, command)
	for _, applied := range progression.commands {
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
func (progression Progression) themeState(number ThemeNumber) (ThemeState, bool) {
	for _, state := range progression.projection.Themes {
		if state.Number == number {
			return state, true
		}
	}
	return ThemeState{}, false
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
func validTheme(theme ThemeNumber) bool { return theme >= ThemeTwo && theme <= ThemeFour }
func validCommand(command Command) bool {
	return idPattern.MatchString(command.ID) && keyPattern.MatchString(command.ActorKey) &&
		reasonPattern.MatchString(command.ReasonCode) && !command.At.IsZero()
}
func fingerprint(id string, action Action, theme ThemeNumber, contentKey, evidenceKey string, command Command) string {
	value := id + "\x00" + string(action) + "\x00" + strconv.Itoa(int(theme)) + "\x00" +
		contentKey + "\x00" + evidenceKey + "\x00" + command.ID + "\x00" + command.ActorKey +
		"\x00" + command.ReasonCode + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (progression Progression) ID() string { return progression.id }
func (progression Progression) Members() []string {
	return append([]string(nil), progression.members...)
}
func (progression Progression) Events() []Event { return append([]Event(nil), progression.events...) }
func (progression Progression) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), progression.commands...)
}
func (progression Progression) Revision() uint64 { return uint64(len(progression.events)) }
func (progression Progression) Projection() Projection {
	projection := Projection{Themes: make([]ThemeState, 0, len(progression.projection.Themes))}
	for _, state := range progression.projection.Themes {
		state.Submissions = append([]RevealedSubmission(nil), state.Submissions...)
		projection.Themes = append(projection.Themes, state)
	}
	return projection
}
func (progression Progression) HasCommand(id string) bool {
	for _, command := range progression.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}

func (number ThemeNumber) String() string { return fmt.Sprintf("%d", number) }
