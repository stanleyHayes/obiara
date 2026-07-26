// Package domain implements the pure E11-S11 scam-arc sequence policy.
// Inputs are opaque categorized events only. No message, voice, payment,
// member identity, free text, score, accusation, or enforcement action is
// representable.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	MaxEvents = 64
	MaxSteps  = 4
)

type Category string

const (
	CategoryChannelShift    Category = "channel_shift"
	CategoryUrgencyPressure Category = "urgency_pressure"
	CategorySecrecyRequest  Category = "secrecy_request"
	CategoryIsolationPrompt Category = "isolation_prompt"
	CategoryResourceRequest Category = "resource_request"
)

type Recommendation string

const (
	RecommendObserve   Recommendation = "observe_only"
	RecommendEducation Recommendation = "education"
	RecommendFriction  Recommendation = "friction"
	RecommendHumanCase Recommendation = "human_case"
)

type ReasonCode string

const (
	ReasonRepeatedPattern ReasonCode = "repeated_pattern"
	ReasonDiversePattern  ReasonCode = "diverse_pattern"
)

var (
	ErrInvalidSequence = errors.New("invalid scam-arc event sequence")
	ErrInvalidRules    = errors.New("invalid reviewed scam-arc rules")
	ErrNoPattern       = errors.New("no multi-event scam-arc pattern")
)

var opaquePattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var tokenPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

type Event struct {
	ID            string
	PairKey       string
	Category      Category
	ObservedAt    time.Time
	SourceVersion uint64
}

type Sequence struct {
	key     string
	pairKey string
	events  []Event
}

func NewSequence(key, pairKey string, events []Event) (Sequence, error) {
	if !opaquePattern.MatchString(key) || !opaquePattern.MatchString(pairKey) ||
		len(events) == 0 || len(events) > MaxEvents {
		return Sequence{}, ErrInvalidSequence
	}
	canonical := append([]Event(nil), events...)
	seen := make(map[string]struct{}, len(canonical))
	for _, event := range canonical {
		if !opaquePattern.MatchString(event.ID) || event.PairKey != pairKey ||
			!validCategory(event.Category) || event.ObservedAt.IsZero() ||
			event.SourceVersion == 0 {
			return Sequence{}, ErrInvalidSequence
		}
		if _, duplicate := seen[event.ID]; duplicate {
			return Sequence{}, ErrInvalidSequence
		}
		seen[event.ID] = struct{}{}
	}
	slices.SortFunc(canonical, func(a, b Event) int {
		if comparison := a.ObservedAt.Compare(b.ObservedAt); comparison != 0 {
			return comparison
		}
		return strings.Compare(a.ID, b.ID)
	})
	return Sequence{key: key, pairKey: pairKey, events: canonical}, nil
}

func (sequence Sequence) Events() []Event {
	return append([]Event(nil), sequence.events...)
}

func (sequence Sequence) Key() string     { return sequence.key }
func (sequence Sequence) PairKey() string { return sequence.pairKey }

type Review struct {
	ID          string
	ReviewerKey string
	ReviewedAt  time.Time
}

type Step struct {
	MinimumEvents     int
	MinimumCategories int
	Recommendation    Recommendation
}

type RuleSpec struct {
	ID      string
	Version uint64
	Window  time.Duration
	Review  Review
	Steps   []Step
}

type RuleSet struct {
	spec RuleSpec
}

func NewRuleSet(spec RuleSpec) (RuleSet, error) {
	spec.Steps = append([]Step(nil), spec.Steps...)
	spec.Review.ReviewedAt = spec.Review.ReviewedAt.UTC()
	if !tokenPattern.MatchString(spec.ID) || spec.Version == 0 ||
		spec.Window < time.Hour || spec.Window > 30*24*time.Hour ||
		!tokenPattern.MatchString(spec.Review.ID) ||
		!opaquePattern.MatchString(spec.Review.ReviewerKey) ||
		spec.Review.ReviewedAt.IsZero() ||
		len(spec.Steps) == 0 || len(spec.Steps) > MaxSteps {
		return RuleSet{}, ErrInvalidRules
	}
	previousEvents, previousCategories, previousRecommendation := 1, 0, -1
	for _, step := range spec.Steps {
		order := recommendationOrder(step.Recommendation)
		if step.MinimumEvents < 2 || step.MinimumEvents > MaxEvents ||
			step.MinimumCategories < 1 || step.MinimumCategories > 5 ||
			step.MinimumEvents <= previousEvents ||
			step.MinimumCategories < previousCategories ||
			order <= previousRecommendation {
			return RuleSet{}, ErrInvalidRules
		}
		previousEvents = step.MinimumEvents
		previousCategories = step.MinimumCategories
		previousRecommendation = order
	}
	return RuleSet{spec: spec}, nil
}

func (rules RuleSet) Spec() RuleSpec {
	spec := rules.spec
	spec.Steps = append([]Step(nil), rules.spec.Steps...)
	return spec
}

type Signal struct {
	ID                    string
	SequenceKey           string
	PairKey               string
	RuleID                string
	RuleVersion           uint64
	Reasons               []ReasonCode
	Recommendation        Recommendation
	EventCount            int
	DistinctCategoryCount int
	WindowStartedAt       time.Time
	EvaluatedAt           time.Time
}

// Evaluate requires at least two events inside the reviewed time window and
// chooses the highest satisfied least-harm step. A single event can never
// create a signal.
func Evaluate(signalID string, sequence Sequence, rules RuleSet, at time.Time) (Signal, error) {
	if !opaquePattern.MatchString(signalID) || at.IsZero() {
		return Signal{}, ErrInvalidSequence
	}
	spec := rules.Spec()
	windowStart := at.UTC().Add(-spec.Window)
	var inWindow []Event
	categories := make(map[Category]struct{})
	for _, event := range sequence.events {
		if event.ObservedAt.After(at.UTC()) || event.ObservedAt.Before(windowStart) {
			continue
		}
		inWindow = append(inWindow, event)
		categories[event.Category] = struct{}{}
	}
	if len(inWindow) < 2 {
		return Signal{}, ErrNoPattern
	}
	selected := Step{}
	found := false
	for _, step := range spec.Steps {
		if len(inWindow) >= step.MinimumEvents && len(categories) >= step.MinimumCategories {
			selected, found = step, true
		}
	}
	if !found {
		return Signal{}, ErrNoPattern
	}
	reasons := []ReasonCode{ReasonRepeatedPattern}
	if len(categories) > 1 {
		reasons = append(reasons, ReasonDiversePattern)
	}
	return Signal{
		ID:                    signalID,
		SequenceKey:           sequence.key,
		PairKey:               sequence.pairKey,
		RuleID:                spec.ID,
		RuleVersion:           spec.Version,
		Reasons:               reasons,
		Recommendation:        selected.Recommendation,
		EventCount:            len(inWindow),
		DistinctCategoryCount: len(categories),
		WindowStartedAt:       windowStart,
		EvaluatedAt:           at.UTC(),
	}, nil
}

func validCategory(category Category) bool {
	return slices.Contains([]Category{
		CategoryChannelShift,
		CategoryUrgencyPressure,
		CategorySecrecyRequest,
		CategoryIsolationPrompt,
		CategoryResourceRequest,
	}, category)
}

func recommendationOrder(recommendation Recommendation) int {
	switch recommendation {
	case RecommendObserve:
		return 0
	case RecommendEducation:
		return 1
	case RecommendFriction:
		return 2
	case RecommendHumanCase:
		return 3
	default:
		return -1
	}
}
