package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strconv"
	"strings"
)

const MaxTurns = 12

var (
	ErrDuelInvalid      = errors.New("invalid private Ebe duel")
	ErrDuelComplete     = errors.New("private Ebe duel is complete")
	ErrUnexpectedPlayer = errors.New("not this player's turn")
	ErrStaleRevision    = errors.New("stale private Ebe duel revision")
	ErrReplayMismatch   = errors.New("private Ebe replay mismatch")
)

type DuelSpec struct {
	ID         string
	PlayerKeys [2]string
	Prompts    []Prompt
}

type Turn struct {
	Number        uint64
	PlayerKey     string
	PromptID      string
	PromptVersion uint64
	PromptDigest  string
	Answer        string
	Correct       bool
	StateDigest   string
}

type Duel struct {
	spec     DuelSpec
	turns    []Turn
	scores   [2]uint8
	complete bool
	revision uint64
}

func NewDuel(spec DuelSpec) (Duel, error) {
	spec.Prompts = append([]Prompt(nil), spec.Prompts...)
	if !idPattern.MatchString(spec.ID) ||
		!opaqueKeyPattern.MatchString(spec.PlayerKeys[0]) ||
		!opaqueKeyPattern.MatchString(spec.PlayerKeys[1]) ||
		spec.PlayerKeys[0] == spec.PlayerKeys[1] ||
		len(spec.Prompts) == 0 || len(spec.Prompts) > MaxTurns {
		return Duel{}, ErrDuelInvalid
	}
	seen := make(map[string]struct{}, len(spec.Prompts))
	for _, prompt := range spec.Prompts {
		promptSpec := prompt.Spec()
		key := promptSpec.ID + ":" + strconv.FormatUint(promptSpec.Version, 10)
		if !opaqueKeyPattern.MatchString(prompt.Digest()) || promptSpec.Review.Decision != DecisionApproved {
			return Duel{}, ErrPromptUnapproved
		}
		if _, duplicate := seen[key]; duplicate {
			return Duel{}, ErrDuelInvalid
		}
		seen[key] = struct{}{}
	}
	return Duel{spec: spec}, nil
}

func (duel Duel) Revision() uint64 { return duel.revision }
func (duel Duel) Complete() bool   { return duel.complete }
func (duel Duel) Scores() [2]uint8 { return duel.scores }

// Specification is server-private replay material. Player keys and accepted
// prompt forms must never be projected to member clients.
func (duel Duel) Specification() DuelSpec {
	return DuelSpec{
		ID: duel.spec.ID, PlayerKeys: duel.spec.PlayerKeys,
		Prompts: append([]Prompt(nil), duel.spec.Prompts...),
	}
}

func (duel Duel) Turns() []Turn {
	return append([]Turn(nil), duel.turns...)
}

func (duel Duel) Prompts() []Prompt {
	return append([]Prompt(nil), duel.spec.Prompts...)
}

// Answer records a private bounded answer. The answer remains inside the
// private aggregate and is not a publishable content object.
func (duel Duel) Answer(playerKey, answer string, expectedRevision uint64) (Duel, error) {
	if expectedRevision != duel.revision {
		return duel, ErrStaleRevision
	}
	if duel.complete || len(duel.turns) >= len(duel.spec.Prompts) || len(duel.turns) >= MaxTurns {
		return duel, ErrDuelComplete
	}
	playerIndex := len(duel.turns) % 2
	if playerKey != duel.spec.PlayerKeys[playerIndex] {
		return duel, ErrUnexpectedPlayer
	}
	prompt := duel.spec.Prompts[len(duel.turns)]
	correct, err := prompt.Accepts(answer)
	if err != nil {
		return duel, err
	}
	promptSpec := prompt.Spec()
	trimmedAnswer := strings.TrimSpace(answer)
	next := duel
	next.turns = append([]Turn(nil), duel.turns...)
	if correct {
		next.scores[playerIndex]++
	}
	next.revision++
	next.complete = len(next.turns)+1 == len(next.spec.Prompts)
	turn := Turn{
		Number:        next.revision,
		PlayerKey:     playerKey,
		PromptID:      promptSpec.ID,
		PromptVersion: promptSpec.Version,
		PromptDigest:  prompt.Digest(),
		Answer:        trimmedAnswer,
		Correct:       correct,
	}
	turn.StateDigest = turnDigest(turn, next.scores, next.complete)
	next.turns = append(next.turns, turn)
	return next, nil
}

// Replay reconstructs a duel from its reviewed prompt snapshot and private
// turns. All derived fields must match; a caller cannot rewrite correctness,
// provenance, order, or state.
func Replay(spec DuelSpec, transcript []Turn) (Duel, error) {
	if len(transcript) > MaxTurns {
		return Duel{}, ErrDuelInvalid
	}
	duel, err := NewDuel(spec)
	if err != nil {
		return Duel{}, err
	}
	for _, recorded := range transcript {
		next, answerErr := duel.Answer(recorded.PlayerKey, recorded.Answer, duel.revision)
		if answerErr != nil {
			return Duel{}, errors.Join(ErrReplayMismatch, answerErr)
		}
		actual := next.turns[len(next.turns)-1]
		if actual != recorded {
			return Duel{}, ErrReplayMismatch
		}
		duel = next
	}
	return duel, nil
}

func turnDigest(turn Turn, scores [2]uint8, complete bool) string {
	payload := strings.Join([]string{
		strconv.FormatUint(turn.Number, 10),
		turn.PlayerKey,
		turn.PromptID,
		strconv.FormatUint(turn.PromptVersion, 10),
		turn.PromptDigest,
		turn.Answer,
		strconv.FormatBool(turn.Correct),
		strconv.Itoa(int(scores[0])),
		strconv.Itoa(int(scores[1])),
		strconv.FormatBool(complete),
	}, "\x1f")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// PromptVersions is the exact private content snapshot used for the duel.
// It is useful for replay manifests without exposing accepted answers.
func (duel Duel) PromptVersions() []string {
	versions := make([]string, 0, len(duel.spec.Prompts))
	for _, prompt := range duel.spec.Prompts {
		spec := prompt.Spec()
		versions = append(versions, spec.ID+":"+strconv.FormatUint(spec.Version, 10)+":"+prompt.Digest())
	}
	return slices.Clone(versions)
}
