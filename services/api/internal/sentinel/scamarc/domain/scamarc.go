// Package domain models scam-arc sequence signals and the action ladder
// (E11-S11; Doc 08 §7): the documented West-African playbooks — accelerated
// affection cadence, emergency-narrative onset, off-platform pull, ask
// patterns — scored from metadata and screened signals, driving the ladder
// silent watch → education → friction → T&S case.
package domain

import (
	"time"
)

// SignalKind is a scam-arc indicator class.
type SignalKind string

const (
	SignalAffectionCadence   SignalKind = "affection_cadence"
	SignalEmergencyNarrative SignalKind = "emergency_narrative"
	SignalOffPlatformPull    SignalKind = "off_platform_pull"
	SignalAskPattern         SignalKind = "ask_pattern"
)

// LadderState is the intervention rung (Doc 08 §7 action ladder).
type LadderState string

const (
	LadderNone      LadderState = "none"
	LadderWatch     LadderState = "watch"
	LadderEducation LadderState = "education"
	LadderFriction  LadderState = "friction"
	LadderCase      LadderState = "case"
)

// signalWeights: each indicator contributes to the room's risk score.
// Off-platform pull and ask patterns are the strongest predictors.
var signalWeights = map[SignalKind]float64{
	SignalAffectionCadence:   1.0,
	SignalEmergencyNarrative: 2.0,
	SignalOffPlatformPull:    2.5,
	SignalAskPattern:         3.0,
}

// Ladder thresholds: score = Σ kind weights × distinct-kind diversity
// bonus (1.0 for one kind, 1.25 for two, 1.5 for three or more).
const (
	watchThreshold     = 2.0
	educationThreshold = 4.0
	frictionThreshold  = 6.0
	caseThreshold      = 8.0
)

// Score sums weights over distinct signal kinds with a diversity bonus.
func Score(kinds []SignalKind) float64 {
	seen := map[SignalKind]bool{}
	total := 0.0
	for _, kind := range kinds {
		if !seen[kind] {
			total += signalWeights[kind]
			seen[kind] = true
		}
	}
	switch len(seen) {
	case 0:
		return 0
	case 1:
		return total
	case 2:
		return total * 1.25
	default:
		return total * 1.5
	}
}

// LadderFor maps a score to its rung.
func LadderFor(score float64) LadderState {
	switch {
	case score >= caseThreshold:
		return LadderCase
	case score >= frictionThreshold:
		return LadderFriction
	case score >= educationThreshold:
		return LadderEducation
	case score >= watchThreshold:
		return LadderWatch
	default:
		return LadderNone
	}
}

// Signal is one observed indicator on a room.
type Signal struct {
	ID         string
	RoomID     string
	ActorID    string
	Kind       SignalKind
	ObservedAt time.Time
}

// Valid reports whether a signal is well-formed.
func (signal Signal) Valid() bool {
	switch signal.Kind {
	case SignalAffectionCadence, SignalEmergencyNarrative, SignalOffPlatformPull, SignalAskPattern:
		return signal.RoomID != "" && signal.ActorID != ""
	}
	return false
}

// RoomState is the persisted risk position of a room.
type RoomState struct {
	RoomID    string
	Score     float64
	Ladder    LadderState
	UpdatedAt time.Time
}

// Recompute re-scores the room from its full signal set (deterministic,
// never cached-elsewhere).
func Recompute(roomID string, kinds []SignalKind, now time.Time) RoomState {
	score := Score(kinds)
	return RoomState{RoomID: roomID, Score: score, Ladder: LadderFor(score), UpdatedAt: now.UTC()}
}
