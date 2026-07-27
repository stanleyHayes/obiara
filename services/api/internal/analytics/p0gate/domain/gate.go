// Package domain computes a privacy-minimal P0 phase-exit projection from
// bounded cohort aggregates. It never receives source events or member IDs.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	ErrInvalidDefinition = errors.New("invalid P0 gate definition")
	ErrInvalidSnapshot   = errors.New("invalid P0 aggregate snapshot")
)

var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

const MaxAggregateCount uint64 = 100_000_000

type Metric string

const (
	MetricPodsHeard      Metric = "pods_heard"
	MetricSeedToSprout   Metric = "seed_to_sprout"
	MetricSproutToRoom   Metric = "sprout_to_room"
	MetricWeeklyFire     Metric = "weekly_fire_attendance"
	MetricDay30Retention Metric = "day_30_retention"
	MetricRegretTrend    Metric = "regret_trend"
	MetricTierAResolved  Metric = "tier_a_resolved"
)

type DefinitionSpec struct {
	ID, ReviewID, ReviewerKey                                     string
	Version                                                       uint64
	ReviewedAt                                                    time.Time
	PodsHeardPermille, SeedToSproutPermille, SproutToRoomPermille uint16
	WeeklyFirePermille, Day30RetentionPermille                    uint16
}
type Definition struct{ spec DefinitionSpec }

func NewDefinition(spec DefinitionSpec) (Definition, error) {
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !token.MatchString(spec.ID) || !token.MatchString(spec.ReviewID) || !opaque.MatchString(spec.ReviewerKey) ||
		spec.Version == 0 || spec.ReviewedAt.IsZero() ||
		spec.PodsHeardPermille != 650 || spec.SeedToSproutPermille != 250 ||
		spec.SproutToRoomPermille != 350 || spec.WeeklyFirePermille != 400 ||
		spec.Day30RetentionPermille != 450 {
		return Definition{}, ErrInvalidDefinition
	}
	return Definition{spec}, nil
}
func (d Definition) Spec() DefinitionSpec { return d.spec }

type Snapshot struct {
	ID, WindowKey, SourceWatermark              string
	Version                                     uint64
	WindowStartedAt, WindowEndedAt              time.Time
	CohortSize                                  uint64
	PodEligible, PodsHeard                      uint64
	SeedsSown, SproutsOpened                    uint64
	SproutEligible, RoomsOpened                 uint64
	WeeklyFireAttendees                         uint64
	Day30Eligible, Day30Retained                uint64
	PreviousRegretReports, CurrentRegretReports uint64
	UnresolvedTierA                             uint64
	CompleteMetrics                             []Metric
}

type MetricState string

const (
	StatePass       MetricState = "pass"
	StateFail       MetricState = "fail"
	StateIncomplete MetricState = "incomplete"
)

type Result struct {
	Metric                 Metric
	Numerator, Denominator uint64
	ObservedPermille       uint16
	State                  MetricState
}
type Outcome string

const (
	OutcomePass       Outcome = "pass"
	OutcomeFail       Outcome = "fail"
	OutcomeIncomplete Outcome = "incomplete"
)

type Report struct {
	ID, WindowKey, SnapshotID, SourceWatermark, DefinitionID string
	SnapshotVersion, DefinitionVersion                       uint64
	Results                                                  []Result
	Outcome                                                  Outcome
	EvaluatedAt                                              time.Time
}

func Evaluate(reportID string, definition Definition, s Snapshot, at time.Time) (Report, error) {
	if !opaque.MatchString(reportID) || !validSnapshot(s) || at.IsZero() {
		return Report{}, ErrInvalidSnapshot
	}
	d := definition.Spec()
	complete := map[Metric]bool{}
	for _, m := range s.CompleteMetrics {
		complete[m] = true
	}
	results := []Result{
		ratio(MetricPodsHeard, s.PodsHeard, s.PodEligible, d.PodsHeardPermille, complete),
		ratio(MetricSeedToSprout, s.SproutsOpened, s.SeedsSown, d.SeedToSproutPermille, complete),
		ratio(MetricSproutToRoom, s.RoomsOpened, s.SproutEligible, d.SproutToRoomPermille, complete),
		ratio(MetricWeeklyFire, s.WeeklyFireAttendees, s.CohortSize, d.WeeklyFirePermille, complete),
		ratio(MetricDay30Retention, s.Day30Retained, s.Day30Eligible, d.Day30RetentionPermille, complete),
		binary(MetricRegretTrend, s.CurrentRegretReports, s.PreviousRegretReports, complete[MetricRegretTrend], s.PreviousRegretReports > 0 && s.CurrentRegretReports < s.PreviousRegretReports),
		binary(MetricTierAResolved, s.UnresolvedTierA, 0, complete[MetricTierAResolved], s.UnresolvedTierA == 0),
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
	return Report{reportID, s.WindowKey, s.ID, s.SourceWatermark, d.ID, s.Version, d.Version, results, outcome, at.UTC()}, nil
}

func ratio(metric Metric, numerator, denominator uint64, threshold uint16, complete map[Metric]bool) Result {
	result := Result{Metric: metric, Numerator: numerator, Denominator: denominator, State: StateIncomplete}
	if !complete[metric] || denominator == 0 {
		return result
	}
	value := numerator * 1000 / denominator
	if value > 1000 {
		value = 1000
	}
	result.ObservedPermille = uint16(value)
	result.State = StateFail
	if result.ObservedPermille >= threshold {
		result.State = StatePass
	}
	return result
}
func binary(metric Metric, numerator, denominator uint64, complete, passed bool) Result {
	state := StateIncomplete
	if complete {
		state = StateFail
		if passed {
			state = StatePass
		}
	}
	return Result{Metric: metric, Numerator: numerator, Denominator: denominator, State: state}
}
func validSnapshot(s Snapshot) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.WindowKey) || !opaque.MatchString(s.SourceWatermark) || s.Version == 0 ||
		s.WindowStartedAt.IsZero() || !s.WindowEndedAt.After(s.WindowStartedAt) || s.CohortSize == 0 || s.CohortSize > MaxAggregateCount ||
		s.PodsHeard > s.PodEligible || s.SproutsOpened > s.SeedsSown || s.RoomsOpened > s.SproutEligible ||
		s.WeeklyFireAttendees > s.CohortSize || s.Day30Retained > s.Day30Eligible ||
		s.PodEligible > MaxAggregateCount || s.SeedsSown > MaxAggregateCount || s.SproutEligible > MaxAggregateCount ||
		s.Day30Eligible > MaxAggregateCount || s.PreviousRegretReports > MaxAggregateCount ||
		s.CurrentRegretReports > MaxAggregateCount || s.UnresolvedTierA > MaxAggregateCount {
		return false
	}
	seen := map[Metric]bool{}
	for _, m := range s.CompleteMetrics {
		if !validMetric(m) || seen[m] {
			return false
		}
		seen[m] = true
	}
	return true
}
func validMetric(m Metric) bool {
	return slices.Contains([]Metric{MetricPodsHeard, MetricSeedToSprout, MetricSproutToRoom, MetricWeeklyFire, MetricDay30Retention, MetricRegretTrend, MetricTierAResolved}, m)
}
func (r Report) Canonical() Report {
	r.Results = append([]Result(nil), r.Results...)
	slices.SortFunc(r.Results, func(a, b Result) int { return strings.Compare(string(a.Metric), string(b.Metric)) })
	return r
}
