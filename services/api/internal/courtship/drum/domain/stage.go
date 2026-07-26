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

type Medium string

const (
	MediumVoice Medium = "voice"
	MediumText  Medium = "text"
)

var (
	ErrInvalid         = errors.New("invalid courtship drum stage")
	ErrNotTurn         = errors.New("courtship drum turn denied")
	ErrVoiceRequired   = errors.New("voice is required to open the drum stage")
	ErrStaleRevision   = errors.New("stale courtship drum revision")
	ErrCommandMismatch = errors.New("courtship drum command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	reasonPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Command struct {
	ID               string
	ActorKey         string
	ReasonCode       string
	ExpectedRevision uint64
	At               time.Time
}
type Beat struct {
	Sequence   uint64
	CommandID  string
	ActorKey   string
	Medium     Medium
	ContentKey string
	ReasonCode string
	At         time.Time
}
type AppliedCommand struct {
	ID          string
	Fingerprint string
	Revision    uint64
}
type Stage struct {
	id       string
	members  []string
	beats    []Beat
	commands []AppliedCommand
}
type State struct {
	ID       string
	Members  []string
	Beats    []Beat
	Commands []AppliedCommand
}

func Open(id string, members []string, voiceKey string, command Command) (Stage, error) {
	pair, ok := normalizePair(members)
	if !idPattern.MatchString(id) || !ok || command.ExpectedRevision != 0 ||
		!slices.Contains(pair, command.ActorKey) || !keyPattern.MatchString(voiceKey) {
		return Stage{}, ErrVoiceRequired
	}
	stage := Stage{id: id, members: pair}
	return stage.append(MediumVoice, voiceKey, command)
}

func (stage Stage) Add(medium Medium, contentKey string, command Command) (Stage, error) {
	if replayed, err := stage.replay(medium, contentKey, command); replayed || err != nil {
		return stage, err
	}
	if !keyPattern.MatchString(contentKey) || medium != MediumVoice && medium != MediumText {
		return Stage{}, ErrInvalid
	}
	if command.ActorKey != stage.NextActorKey() {
		return Stage{}, ErrNotTurn
	}
	return stage.append(medium, contentKey, command)
}

func (stage Stage) append(medium Medium, contentKey string, command Command) (Stage, error) {
	if !validCommand(command) {
		return Stage{}, ErrInvalid
	}
	if command.ExpectedRevision != uint64(len(stage.beats)) {
		return Stage{}, ErrStaleRevision
	}
	if len(stage.beats) > 0 && command.At.Before(stage.beats[len(stage.beats)-1].At) {
		return Stage{}, ErrInvalid
	}
	revision := uint64(len(stage.beats) + 1)
	beat := Beat{
		Sequence: revision, CommandID: command.ID, ActorKey: command.ActorKey,
		Medium: medium, ContentKey: contentKey, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	}
	stage.beats = append(append([]Beat(nil), stage.beats...), beat)
	stage.commands = append(append([]AppliedCommand(nil), stage.commands...), AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(stage.id, medium, contentKey, command), Revision: revision,
	})
	return stage, nil
}

func Rehydrate(state State) (Stage, error) {
	members, ok := normalizePair(state.Members)
	if !idPattern.MatchString(state.ID) || !ok || len(state.Beats) == 0 || len(state.Beats) != len(state.Commands) {
		return Stage{}, ErrInvalid
	}
	stage := Stage{id: state.ID, members: members}
	seen := map[string]struct{}{}
	for index, beat := range state.Beats {
		applied := state.Commands[index]
		command := Command{
			ID: beat.CommandID, ActorKey: beat.ActorKey, ReasonCode: beat.ReasonCode,
			ExpectedRevision: uint64(index), At: beat.At,
		}
		if beat.Sequence != uint64(index+1) || applied.ID != beat.CommandID ||
			applied.Revision != beat.Sequence || !validCommand(command) ||
			!slices.Contains(members, beat.ActorKey) || !keyPattern.MatchString(beat.ContentKey) ||
			beat.Medium != MediumVoice && beat.Medium != MediumText {
			return Stage{}, ErrInvalid
		}
		if index == 0 && beat.Medium != MediumVoice {
			return Stage{}, ErrVoiceRequired
		}
		if index > 0 && beat.ActorKey != stage.NextActorKey() {
			return Stage{}, ErrNotTurn
		}
		if _, duplicate := seen[applied.ID]; duplicate ||
			applied.Fingerprint != fingerprint(state.ID, beat.Medium, beat.ContentKey, command) {
			return Stage{}, ErrInvalid
		}
		seen[applied.ID] = struct{}{}
		stage.beats = append(stage.beats, beat)
		stage.commands = append(stage.commands, applied)
	}
	return stage, nil
}

func (stage Stage) replay(medium Medium, contentKey string, command Command) (bool, error) {
	expected := fingerprint(stage.id, medium, contentKey, command)
	for _, applied := range stage.commands {
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

func (stage Stage) NextActorKey() string {
	if len(stage.beats) == 0 {
		return ""
	}
	last := stage.beats[len(stage.beats)-1].ActorKey
	if stage.members[0] == last {
		return stage.members[1]
	}
	return stage.members[0]
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
func fingerprint(id string, medium Medium, contentKey string, command Command) string {
	value := id + "\x00" + string(medium) + "\x00" + contentKey + "\x00" +
		command.ID + "\x00" + command.ActorKey + "\x00" + command.ReasonCode + "\x00" +
		strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (stage Stage) ID() string        { return stage.id }
func (stage Stage) Members() []string { return append([]string(nil), stage.members...) }
func (stage Stage) Beats() []Beat     { return append([]Beat(nil), stage.beats...) }
func (stage Stage) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), stage.commands...)
}
func (stage Stage) Revision() uint64 { return uint64(len(stage.beats)) }
func (stage Stage) HasCommand(id string) bool {
	for _, command := range stage.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}
