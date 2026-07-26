package domain

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"
)

const (
	MinCohort    = 200
	MinEvidence  = 1000
	MinPrecision = 0.98
	MaxNodes     = 10_000
	MaxEdges     = 50_000
)

var ErrInvalid = errors.New("invalid anomaly aggregate")
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var privacyKey = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Shape string

const (
	ShapeSyndicate     Shape = "syndicate"
	ShapeVouchRing     Shape = "vouch_ring"
	ShapeDeviceAnomaly Shape = "device_anomaly"
)

type Rule struct {
	Key           string
	Version       uint64
	Shape         Shape
	MinNodes      int
	MinEdges      int
	MinDensity    float64
	ReviewedByKey string
	ReviewedAt    time.Time
}

func NewRule(key string, version uint64, shape Shape, minNodes, minEdges int, minDensity float64, reviewer string, reviewedAt time.Time) (Rule, error) {
	r := Rule{Key: strings.TrimSpace(key), Version: version, Shape: shape, MinNodes: minNodes, MinEdges: minEdges, MinDensity: minDensity, ReviewedByKey: reviewer, ReviewedAt: reviewedAt.UTC()}
	if !validRule(r, reviewedAt) {
		return Rule{}, ErrInvalid
	}
	return r, nil
}
func validRule(r Rule, at time.Time) bool {
	return token.MatchString(r.Key) && r.Version > 0 && validShape(r.Shape) && r.MinNodes >= 2 && r.MinNodes <= MaxNodes && r.MinEdges >= 1 && r.MinEdges <= MaxEdges && finite(r.MinDensity) && r.MinDensity >= 0 && r.MinDensity <= 1 && privacyKey.MatchString(r.ReviewedByKey) && !r.ReviewedAt.IsZero() && !r.ReviewedAt.After(at.UTC())
}
func validShape(s Shape) bool {
	return s == ShapeSyndicate || s == ShapeVouchRing || s == ShapeDeviceAnomaly
}

type Aggregate struct {
	EvidenceRef string
	RuleKey     string
	RuleVersion uint64
	Shape       Shape
	Cohort      int
	Evidence    int
	Nodes       int
	Edges       int
	Density     float64
	Precision   float64
	Uncertain   bool
	ModelError  bool
}

func NewAggregate(evidenceRef, ruleKey string, ruleVersion uint64, shape Shape, cohort, evidence, nodes, edges int, precision float64, uncertain, modelError bool) (Aggregate, error) {
	a := Aggregate{EvidenceRef: evidenceRef, RuleKey: ruleKey, RuleVersion: ruleVersion, Shape: shape, Cohort: cohort, Evidence: evidence, Nodes: nodes, Edges: edges, Precision: precision, Uncertain: uncertain, ModelError: modelError}
	if nodes > 1 {
		a.Density = float64(edges) / (float64(nodes) * float64(nodes-1))
	}
	if !validAggregate(a) {
		return Aggregate{}, ErrInvalid
	}
	return a, nil
}
func validAggregate(a Aggregate) bool {
	maxDirected := a.Nodes * (a.Nodes - 1)
	return privacyKey.MatchString(a.EvidenceRef) && token.MatchString(a.RuleKey) && a.RuleVersion > 0 && validShape(a.Shape) && a.Cohort >= 0 && a.Evidence >= 0 && a.Nodes >= 2 && a.Nodes <= MaxNodes && a.Edges >= 0 && a.Edges <= MaxEdges && a.Edges <= maxDirected && finite(a.Density) && a.Density >= 0 && a.Density <= 1 && finite(a.Precision) && a.Precision >= 0 && a.Precision <= 1
}

type Outcome string

const (
	OutcomeNoAction    Outcome = "no_action"
	OutcomeHumanReview Outcome = "human_review"
)

type Decision struct {
	Outcome     Outcome
	RuleKey     string
	RuleVersion uint64
	Shape       Shape
	EvidenceRef string
	EvaluatedAt time.Time
}

func Evaluate(rule Rule, a Aggregate, at time.Time) (Decision, error) {
	if at.IsZero() || !validRule(rule, at) || !validAggregate(a) {
		return Decision{}, ErrInvalid
	}
	d := Decision{Outcome: OutcomeNoAction, RuleKey: a.RuleKey, RuleVersion: a.RuleVersion, Shape: a.Shape, EvidenceRef: a.EvidenceRef, EvaluatedAt: at.UTC()}
	if rule.Key == a.RuleKey && rule.Version == a.RuleVersion && rule.Shape == a.Shape && !a.Uncertain && !a.ModelError && a.Cohort >= MinCohort && a.Evidence >= MinEvidence && a.Precision >= MinPrecision && a.Nodes >= rule.MinNodes && a.Edges >= rule.MinEdges && a.Density >= rule.MinDensity {
		d.Outcome = OutcomeHumanReview
	}
	return d, nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
