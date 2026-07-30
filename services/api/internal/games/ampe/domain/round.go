// Package domain implements the pure server-authoritative E10-S07 Ampe
// realtime round. It models explicit manual choices only: no camera, image,
// pose, body, matching, rating, trust, or public-score data enters the model.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const MaxCommands = 32

type Choice string

const (
	ChoiceTogether Choice = "together"
	ChoiceApart    Choice = "apart"
)

type Action string

const (
	ActionReady      Action = "ready"
	ActionLock       Action = "lock"
	ActionDisconnect Action = "disconnect"
	ActionReconnect  Action = "reconnect"
)

var (
	ErrInvalidRound      = errors.New("invalid private Ampe round")
	ErrInvalidCommand    = errors.New("invalid private Ampe command")
	ErrInvalidTransition = errors.New("private Ampe transition unavailable")
	ErrStaleSequence     = errors.New("stale private Ampe sequence")
	ErrCommandMismatch   = errors.New("private Ampe command replay mismatch")
	ErrRoundComplete     = errors.New("private Ampe round is complete")
	ErrTranscript        = errors.New("private Ampe transcript mismatch")
)

var (
	opaqueKeyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
)

type Spec struct {
	ID         string
	RoomKey    string
	PlayerKeys [2]string
}

type Command struct {
	ID               string
	ActorKey         string
	Action           Action
	Choice           Choice
	ExpectedSequence uint64
}

type Event struct {
	Sequence  uint64
	CommandID string
	ActorKey  string
	Action    Action
}

type PlayerView struct {
	Key       string
	Ready     bool
	Connected bool
	Locked    bool
}

type Reveal struct {
	Sequence uint64
	Choices  [2]Choice
}

// View is safe to return to one of the two players. OpponentChoice does not
// exist: before reveal the only choice field is the viewer's own.
type View struct {
	ID, RoomKey string
	Sequence    uint64
	Players     [2]PlayerView
	Paused      bool
	OwnChoice   *Choice
	Reveal      *Reveal
}

type Transcript struct {
	Commands    []Command
	FinalDigest string
}

type playerState struct {
	key       string
	ready     bool
	connected bool
	locked    bool
	choice    Choice
}

type Round struct {
	spec         Spec
	players      [2]playerState
	sequence     uint64
	paused       bool
	reveal       *Reveal
	events       []Event
	stateDigests []string
	commands     []Command
	fingerprints map[string]string
}

func Open(spec Spec) (Round, error) {
	if !idPattern.MatchString(spec.ID) ||
		!opaqueKeyPattern.MatchString(spec.RoomKey) ||
		!opaqueKeyPattern.MatchString(spec.PlayerKeys[0]) ||
		!opaqueKeyPattern.MatchString(spec.PlayerKeys[1]) ||
		spec.PlayerKeys[0] == spec.PlayerKeys[1] {
		return Round{}, ErrInvalidRound
	}
	return Round{
		spec: spec,
		players: [2]playerState{
			{key: spec.PlayerKeys[0], connected: true},
			{key: spec.PlayerKeys[1], connected: true},
		},
		fingerprints: make(map[string]string),
	}, nil
}

// Apply validates and atomically applies one server command. The second lock
// produces the reveal in that same sequence; there is no observable state in
// which both are locked but only one choice is visible.
func (round Round) Apply(command Command) (Round, bool, error) {
	fingerprint := commandFingerprint(command)
	if prior, exists := round.fingerprints[command.ID]; exists {
		if prior != fingerprint {
			return round, false, ErrCommandMismatch
		}
		return round, true, nil
	}
	if round.reveal != nil {
		return round, false, ErrRoundComplete
	}
	if len(round.commands) >= MaxCommands {
		return round, false, ErrInvalidTransition
	}
	if !validCommand(command) {
		return round, false, ErrInvalidCommand
	}
	if command.ExpectedSequence != round.sequence {
		return round, false, ErrStaleSequence
	}
	index := round.playerIndex(command.ActorKey)
	if index < 0 {
		return round, false, ErrInvalidCommand
	}

	next := round.clone()
	player := &next.players[index]
	switch command.Action {
	case ActionReady:
		if !player.connected || player.ready || command.Choice != "" {
			return round, false, ErrInvalidTransition
		}
		player.ready = true
	case ActionLock:
		if next.paused || !player.connected || !next.players[0].ready || !next.players[1].ready ||
			player.locked || !validChoice(command.Choice) {
			return round, false, ErrInvalidTransition
		}
		player.locked = true
		player.choice = command.Choice
	case ActionDisconnect:
		if !player.connected || command.Choice != "" {
			return round, false, ErrInvalidTransition
		}
		player.connected = false
		next.paused = true
	case ActionReconnect:
		if player.connected || command.Choice != "" {
			return round, false, ErrInvalidTransition
		}
		player.connected = true
		next.paused = !next.players[0].connected || !next.players[1].connected
	default:
		return round, false, ErrInvalidCommand
	}

	next.sequence++
	if next.players[0].locked && next.players[1].locked {
		next.reveal = &Reveal{
			Sequence: next.sequence,
			Choices:  [2]Choice{next.players[0].choice, next.players[1].choice},
		}
	}
	event := Event{
		Sequence:  next.sequence,
		CommandID: command.ID,
		ActorKey:  command.ActorKey,
		Action:    command.Action,
	}
	stateDigest := next.stateDigest(fingerprint)
	next.events = append(next.events, event)
	next.stateDigests = append(next.stateDigests, stateDigest)
	next.commands = append(next.commands, command)
	next.fingerprints[command.ID] = fingerprint
	return next, false, nil
}

func (round Round) View(viewerKey string) (View, error) {
	index := round.playerIndex(viewerKey)
	if index < 0 {
		return View{}, ErrInvalidCommand
	}
	view := View{
		ID:       round.spec.ID,
		RoomKey:  round.spec.RoomKey,
		Sequence: round.sequence,
		Paused:   round.paused,
	}
	for playerIndex, player := range round.players {
		view.Players[playerIndex] = PlayerView{
			Key:       player.key,
			Ready:     player.ready,
			Connected: player.connected,
			Locked:    player.locked,
		}
	}
	if round.players[index].choice != "" {
		choice := round.players[index].choice
		view.OwnChoice = &choice
	}
	if round.reveal != nil {
		reveal := *round.reveal
		view.Reveal = &reveal
	}
	return view, nil
}

func (round Round) Sequence() uint64 { return round.sequence }

// Specification returns server-private replay material. It must never be
// projected to clients because it contains privacy-keyed room and player refs.
func (round Round) Specification() Spec { return round.spec }

func (round Round) Events() []Event {
	return append([]Event(nil), round.events...)
}

// PrivateTranscript is server-private replay material. Public-safe events
// never contain raw choices; those belong only in this bounded transcript.
func (round Round) PrivateTranscript() Transcript {
	return Transcript{
		Commands:    append([]Command(nil), round.commands...),
		FinalDigest: round.finalDigest(),
	}
}

func Replay(spec Spec, transcript Transcript) (Round, error) {
	if len(transcript.Commands) > MaxCommands || !opaqueKeyPattern.MatchString(transcript.FinalDigest) {
		return Round{}, ErrTranscript
	}
	round, err := Open(spec)
	if err != nil {
		return Round{}, err
	}
	for _, command := range transcript.Commands {
		next, replayed, applyErr := round.Apply(command)
		if applyErr != nil || replayed {
			return Round{}, errors.Join(ErrTranscript, applyErr)
		}
		round = next
	}
	if round.finalDigest() != transcript.FinalDigest {
		return Round{}, ErrTranscript
	}
	return round, nil
}

func (round Round) clone() Round {
	next := round
	next.events = append([]Event(nil), round.events...)
	next.stateDigests = append([]string(nil), round.stateDigests...)
	next.commands = append([]Command(nil), round.commands...)
	next.fingerprints = make(map[string]string, len(round.fingerprints)+1)
	for id, fingerprint := range round.fingerprints {
		next.fingerprints[id] = fingerprint
	}
	if round.reveal != nil {
		reveal := *round.reveal
		next.reveal = &reveal
	}
	return next
}

func (round Round) playerIndex(key string) int {
	for index, player := range round.players {
		if player.key == key {
			return index
		}
	}
	return -1
}

func validCommand(command Command) bool {
	return idPattern.MatchString(command.ID) && opaqueKeyPattern.MatchString(command.ActorKey)
}

func validChoice(choice Choice) bool {
	return choice == ChoiceTogether || choice == ChoiceApart
}

func commandFingerprint(command Command) string {
	payload := strings.Join([]string{
		command.ID,
		command.ActorKey,
		string(command.Action),
		string(command.Choice),
		strconv.FormatUint(command.ExpectedSequence, 10),
	}, "\x1f")
	return digest(payload)
}

func (round Round) stateDigest(commandFingerprint string) string {
	parts := []string{
		round.spec.ID,
		round.spec.RoomKey,
		strconv.FormatUint(round.sequence, 10),
		strconv.FormatBool(round.paused),
		commandFingerprint,
	}
	for _, player := range round.players {
		parts = append(parts,
			player.key,
			strconv.FormatBool(player.ready),
			strconv.FormatBool(player.connected),
			strconv.FormatBool(player.locked),
			string(player.choice),
		)
	}
	if round.reveal != nil {
		parts = append(parts, string(round.reveal.Choices[0]), string(round.reveal.Choices[1]))
	}
	return digest(strings.Join(parts, "\x1e"))
}

func (round Round) finalDigest() string {
	parts := []string{round.spec.ID, round.spec.RoomKey, strconv.FormatUint(round.sequence, 10)}
	parts = append(parts, round.stateDigests...)
	return digest(strings.Join(parts, "\x1f"))
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
