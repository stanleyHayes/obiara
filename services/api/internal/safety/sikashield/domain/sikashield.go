package domain

import (
	"errors"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	MinEvidence   = 1000
	MinPositive   = 100
	MinPrecision  = 0.97
	MaxReviewRate = 0.20
	MaxPatterns   = 64
)

var ErrInvalid = errors.New("invalid sika shield evaluation")
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var ref = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Source string

const (
	SourceText          Source = "text"
	SourceVoiceMetadata Source = "voice_metadata"
)

type Pattern struct {
	Key           string
	Version       uint64
	Source        Source
	ExpressionRef string
	ReviewedByKey string
	ReviewedAt    time.Time
}

func NewPattern(key string, version uint64, source Source, expressionRef, reviewer string, at time.Time) (Pattern, error) {
	p := Pattern{strings.TrimSpace(key), version, source, strings.TrimSpace(expressionRef), reviewer, at.UTC()}
	if !token.MatchString(p.Key) || version == 0 || (source != SourceText && source != SourceVoiceMetadata) || !ref.MatchString(p.ExpressionRef) || !ref.MatchString(reviewer) || at.IsZero() {
		return Pattern{}, ErrInvalid
	}
	return p, nil
}

type Metrics struct {
	Evidence, Positive, TruePositive, FalsePositive, ReviewCount int
	Precision, ReviewRate                                        float64
}

func NewMetrics(evidence, positive, truePositive, falsePositive, reviewCount int) (Metrics, error) {
	m := Metrics{Evidence: evidence, Positive: positive, TruePositive: truePositive, FalsePositive: falsePositive, ReviewCount: reviewCount}
	if evidence <= 0 {
		return Metrics{}, ErrInvalid
	}
	m.Precision = float64(truePositive) / float64(truePositive+falsePositive)
	m.ReviewRate = float64(reviewCount) / float64(evidence)
	if !m.finite() || positive != truePositive+falsePositive || truePositive < 0 || falsePositive < 0 || reviewCount < 0 || reviewCount > evidence {
		return Metrics{}, ErrInvalid
	}
	return m, nil
}
func (m Metrics) finite() bool {
	return !math.IsNaN(m.Precision) && !math.IsInf(m.Precision, 0) && !math.IsNaN(m.ReviewRate) && !math.IsInf(m.ReviewRate, 0)
}
func (m Metrics) Pass() bool {
	return m.finite() && m.Positive == m.TruePositive+m.FalsePositive && m.TruePositive >= 0 && m.FalsePositive >= 0 &&
		m.ReviewCount >= 0 && m.ReviewCount <= m.Evidence && m.Evidence >= MinEvidence && m.Positive >= MinPositive &&
		m.Precision >= MinPrecision && m.Precision <= 1 && m.ReviewRate >= 0 && m.ReviewRate <= MaxReviewRate
}

type Outcome string

const (
	OutcomeNoAction    Outcome = "no_action"
	OutcomeHumanReview Outcome = "human_review"
)

type Signal struct {
	PatternKey     string
	PatternVersion uint64
	Source         Source
	EvidenceRef    string
	Confidence     float64
	Uncertain      bool
	ModelError     bool
}
type Decision struct {
	Outcome        Outcome
	PatternKey     string
	PatternVersion uint64
	EvidenceRef    string
	EvaluatedAt    time.Time
}

func Evaluate(patterns []Pattern, signal Signal, at time.Time) (Decision, error) {
	if at.IsZero() || len(patterns) == 0 || len(patterns) > MaxPatterns || !ref.MatchString(signal.EvidenceRef) || math.IsNaN(signal.Confidence) || math.IsInf(signal.Confidence, 0) || signal.Confidence < 0 || signal.Confidence > 1 {
		return Decision{}, ErrInvalid
	}
	keys := make([]string, 0, len(patterns))
	var found bool
	for _, p := range patterns {
		if !token.MatchString(p.Key) || p.Version == 0 || (p.Source != SourceText && p.Source != SourceVoiceMetadata) ||
			!ref.MatchString(p.ExpressionRef) || !ref.MatchString(p.ReviewedByKey) || p.ReviewedAt.IsZero() || p.ReviewedAt.After(at.UTC()) {
			return Decision{}, ErrInvalid
		}
		keys = append(keys, p.Key)
		if p.Key == signal.PatternKey && p.Version == signal.PatternVersion && p.Source == signal.Source {
			found = true
		}
	}
	slices.Sort(keys)
	for i := 1; i < len(keys); i++ {
		if keys[i] == keys[i-1] {
			return Decision{}, ErrInvalid
		}
	}
	d := Decision{Outcome: OutcomeNoAction, PatternKey: signal.PatternKey, PatternVersion: signal.PatternVersion, EvidenceRef: signal.EvidenceRef, EvaluatedAt: at.UTC()}
	if found && !signal.ModelError && !signal.Uncertain && signal.Confidence >= MinPrecision {
		d.Outcome = OutcomeHumanReview
	}
	return d, nil
}
