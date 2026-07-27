// Package domain evaluates privacy-thresholded quarterly aggregates. It has no
// source-event, member, matching, enforcement, or rollout concepts.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	MaxAggregateCount uint64 = 100_000_000
	MinCohortSize     uint64 = 50
	MaxCohorts               = 64
)

var (
	ErrInvalidDefinition = errors.New("invalid fairness definition")
	ErrInvalidSnapshot   = errors.New("invalid fairness aggregate snapshot")
	opaque               = regexp.MustCompile(`^[a-f0-9]{64}$`)
	token                = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
)

type Metric string

const (
	MetricExposureParity Metric = "exposure_parity"
	MetricColorismDrift  Metric = "colorism_drift"
	MetricRegretTrend    Metric = "regret_trend"
	MetricTierASafety    Metric = "tier_a_safety"
)

type DefinitionSpec struct {
	ID, ReviewID, ReviewerKey string
	Version                   uint64
	MaxParityGapPermille      uint16
	ReviewedAt                time.Time
}
type Definition struct{ spec DefinitionSpec }

func NewDefinition(spec DefinitionSpec) (Definition, error) {
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !token.MatchString(spec.ID) || !token.MatchString(spec.ReviewID) || !opaque.MatchString(spec.ReviewerKey) ||
		spec.Version == 0 || spec.MaxParityGapPermille > 1000 || spec.ReviewedAt.IsZero() {
		return Definition{}, ErrInvalidDefinition
	}
	return Definition{spec}, nil
}
func (d Definition) Spec() DefinitionSpec { return d.spec }

type CohortAggregate struct {
	CohortKey string `bson:"cohortKey"`
	Eligible  uint64 `bson:"eligible"`
	Exposed   uint64 `bson:"exposed"`
}

type Snapshot struct {
	ID, QuarterKey, SourceWatermark               string
	Version                                       uint64
	WindowStartedAt, WindowEndedAt                time.Time
	Cohorts                                       []CohortAggregate
	PreviousRegretEligible, PreviousRegretReports uint64
	CurrentRegretEligible, CurrentRegretReports   uint64
	UnresolvedTierA                               uint64
	ColorismAuditComplete, ColorismDriftDetected  bool
	CompleteMetrics                               []Metric
}

type MetricState string

const (
	StatePass       MetricState = "pass"
	StateFail       MetricState = "fail"
	StateIncomplete MetricState = "incomplete"
)

type Outcome string

const (
	OutcomePass       Outcome = "pass"
	OutcomeFail       Outcome = "fail"
	OutcomeIncomplete Outcome = "incomplete"
)

type CohortResult struct {
	CohortKey        string `bson:"cohortKey"`
	Exposed          uint64 `bson:"exposed"`
	Eligible         uint64 `bson:"eligible"`
	ExposurePermille uint16 `bson:"exposurePermille"`
}
type Result struct {
	Metric             Metric      `bson:"metric"`
	State              MetricState `bson:"state"`
	Numerator          uint64      `bson:"numerator"`
	Denominator        uint64      `bson:"denominator"`
	ObservedPermille   uint16      `bson:"observedPermille"`
	ComparisonPermille uint16      `bson:"comparisonPermille"`
}
type Report struct {
	ID                        string         `bson:"_id"`
	QuarterKey                string         `bson:"quarterKey"`
	SnapshotID                string         `bson:"snapshotId"`
	SourceWatermark           string         `bson:"sourceWatermark"`
	DefinitionID              string         `bson:"definitionId"`
	Fingerprint               string         `bson:"fingerprint"`
	SnapshotVersion           uint64         `bson:"snapshotVersion"`
	DefinitionVersion         uint64         `bson:"definitionVersion"`
	MaxParityGapPermille      uint16         `bson:"maxParityGapPermille"`
	ObservedParityGapPermille uint16         `bson:"observedParityGapPermille"`
	Cohorts                   []CohortResult `bson:"cohorts"`
	Results                   []Result       `bson:"results"`
	Outcome                   Outcome        `bson:"outcome"`
	EvaluatedAt               time.Time      `bson:"evaluatedAt"`
}

func Evaluate(id string, definition Definition, s Snapshot, at time.Time) (Report, error) {
	if !opaque.MatchString(id) || !validSnapshot(s) || at.IsZero() {
		return Report{}, ErrInvalidSnapshot
	}
	complete := make(map[Metric]bool, len(s.CompleteMetrics))
	for _, metric := range s.CompleteMetrics {
		complete[metric] = true
	}
	cohorts := make([]CohortResult, 0, len(s.Cohorts))
	var minRate, maxRate uint16
	for i, c := range s.Cohorts {
		rate := uint16(c.Exposed * 1000 / c.Eligible)
		if i == 0 || rate < minRate {
			minRate = rate
		}
		if rate > maxRate {
			maxRate = rate
		}
		cohorts = append(cohorts, CohortResult{c.CohortKey, c.Exposed, c.Eligible, rate})
	}
	gap := maxRate - minRate
	spec := definition.Spec()
	results := []Result{
		stateResult(MetricExposureParity, uint64(gap), 1000, complete[MetricExposureParity], gap <= spec.MaxParityGapPermille, gap, spec.MaxParityGapPermille),
		stateResult(MetricColorismDrift, boolCount(s.ColorismDriftDetected), 1, complete[MetricColorismDrift] && s.ColorismAuditComplete, !s.ColorismDriftDetected, uint16(boolCount(s.ColorismDriftDetected)*1000), 0),
		trendResult(s, complete[MetricRegretTrend]),
		stateResult(MetricTierASafety, s.UnresolvedTierA, 0, complete[MetricTierASafety], s.UnresolvedTierA == 0, 0, 0),
	}
	outcome := OutcomePass
	for _, result := range results {
		if result.State == StateIncomplete {
			outcome = OutcomeIncomplete
			break
		}
		if result.State == StateFail {
			outcome = OutcomeFail
		}
	}
	r := Report{id, s.QuarterKey, s.ID, s.SourceWatermark, spec.ID, "", s.Version, spec.Version, spec.MaxParityGapPermille, gap, cohorts, results, outcome, at.UTC()}
	r = r.Canonical()
	r.Fingerprint = fingerprint(r)
	return r, nil
}

func RehydrateReport(r Report) (Report, error) {
	r = r.Canonical()
	if !opaque.MatchString(r.ID) || !opaque.MatchString(r.QuarterKey) || !opaque.MatchString(r.SnapshotID) || !opaque.MatchString(r.SourceWatermark) ||
		!token.MatchString(r.DefinitionID) || !opaque.MatchString(r.Fingerprint) || r.SnapshotVersion == 0 || r.DefinitionVersion == 0 ||
		r.MaxParityGapPermille > 1000 || r.ObservedParityGapPermille > 1000 || len(r.Cohorts) < 2 || len(r.Cohorts) > MaxCohorts ||
		len(r.Results) != 4 || (r.Outcome != OutcomePass && r.Outcome != OutcomeFail && r.Outcome != OutcomeIncomplete) || r.EvaluatedAt.IsZero() {
		return Report{}, ErrInvalidSnapshot
	}
	cohorts := map[string]bool{}
	for _, c := range r.Cohorts {
		if !opaque.MatchString(c.CohortKey) || cohorts[c.CohortKey] || c.Eligible < MinCohortSize || c.Eligible > MaxAggregateCount || c.Exposed > c.Eligible || c.ExposurePermille != uint16(c.Exposed*1000/c.Eligible) {
			return Report{}, ErrInvalidSnapshot
		}
		cohorts[c.CohortKey] = true
	}
	metrics := map[Metric]bool{}
	for _, x := range r.Results {
		if !validMetric(x.Metric) || metrics[x.Metric] || (x.State != StatePass && x.State != StateFail && x.State != StateIncomplete) || x.ObservedPermille > 1000 || x.ComparisonPermille > 1000 {
			return Report{}, ErrInvalidSnapshot
		}
		metrics[x.Metric] = true
	}
	if r.Fingerprint != fingerprint(r) {
		return Report{}, ErrInvalidSnapshot
	}
	return r, nil
}

func trendResult(s Snapshot, complete bool) Result {
	result := Result{Metric: MetricRegretTrend, Numerator: s.CurrentRegretReports, Denominator: s.CurrentRegretEligible, State: StateIncomplete}
	if !complete || s.PreviousRegretEligible == 0 || s.CurrentRegretEligible == 0 {
		return result
	}
	current := uint16(s.CurrentRegretReports * 1000 / s.CurrentRegretEligible)
	previous := uint16(s.PreviousRegretReports * 1000 / s.PreviousRegretEligible)
	result.ObservedPermille, result.ComparisonPermille = current, previous
	result.State = StateFail
	if current < previous {
		result.State = StatePass
	}
	return result
}
func stateResult(metric Metric, numerator, denominator uint64, complete, pass bool, observed, comparison uint16) Result {
	state := StateIncomplete
	if complete {
		state = StateFail
		if pass {
			state = StatePass
		}
	}
	return Result{metric, state, numerator, denominator, observed, comparison}
}
func validSnapshot(s Snapshot) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.QuarterKey) || !opaque.MatchString(s.SourceWatermark) || s.Version == 0 ||
		s.WindowStartedAt.IsZero() || !s.WindowEndedAt.After(s.WindowStartedAt) || len(s.Cohorts) < 2 || len(s.Cohorts) > MaxCohorts ||
		s.PreviousRegretEligible > MaxAggregateCount || s.CurrentRegretEligible > MaxAggregateCount ||
		s.PreviousRegretReports > s.PreviousRegretEligible || s.CurrentRegretReports > s.CurrentRegretEligible || s.UnresolvedTierA > MaxAggregateCount {
		return false
	}
	seenCohorts := map[string]bool{}
	for _, c := range s.Cohorts {
		if !opaque.MatchString(c.CohortKey) || seenCohorts[c.CohortKey] || c.Eligible < MinCohortSize || c.Eligible > MaxAggregateCount || c.Exposed > c.Eligible {
			return false
		}
		seenCohorts[c.CohortKey] = true
	}
	seenMetrics := map[Metric]bool{}
	for _, m := range s.CompleteMetrics {
		if !validMetric(m) || seenMetrics[m] {
			return false
		}
		seenMetrics[m] = true
	}
	return true
}
func validMetric(m Metric) bool {
	return slices.Contains([]Metric{MetricExposureParity, MetricColorismDrift, MetricRegretTrend, MetricTierASafety}, m)
}
func boolCount(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}
func (r Report) Canonical() Report {
	r.Cohorts = append([]CohortResult(nil), r.Cohorts...)
	r.Results = append([]Result(nil), r.Results...)
	slices.SortFunc(r.Cohorts, func(a, b CohortResult) int { return strings.Compare(a.CohortKey, b.CohortKey) })
	slices.SortFunc(r.Results, func(a, b Result) int { return strings.Compare(string(a.Metric), string(b.Metric)) })
	return r
}
func fingerprint(r Report) string {
	var b strings.Builder
	b.WriteString(strings.Join([]string{r.QuarterKey, r.SnapshotID, r.SourceWatermark, r.DefinitionID, strconv.FormatUint(r.SnapshotVersion, 10), strconv.FormatUint(r.DefinitionVersion, 10), strconv.Itoa(int(r.MaxParityGapPermille)), string(r.Outcome)}, "\x00"))
	for _, c := range r.Cohorts {
		b.WriteString("\x00" + c.CohortKey + "\x00" + strconv.FormatUint(c.Exposed, 10) + "\x00" + strconv.FormatUint(c.Eligible, 10))
	}
	for _, x := range r.Results {
		b.WriteString("\x00" + string(x.Metric) + "\x00" + string(x.State) + "\x00" + strconv.FormatUint(x.Numerator, 10) + "\x00" + strconv.FormatUint(x.Denominator, 10) + "\x00" + strconv.Itoa(int(x.ObservedPermille)) + "\x00" + strconv.Itoa(int(x.ComparisonPermille)))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
