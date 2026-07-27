package domain

import (
	"math/rand"
	"slices"
	"testing"
	"time"
)

const (
	keyA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	keyB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	keyC = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	keyD = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
)

func definition(t *testing.T, gap uint16) Definition {
	t.Helper()
	d, err := NewDefinition(DefinitionSpec{ID: "fairness.v1", ReviewID: "mpanyimfo.review", ReviewerKey: keyA, Version: 1, MaxParityGapPermille: gap, ReviewedAt: time.Unix(1, 0)})
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func completeSnapshot() Snapshot {
	return Snapshot{ID: keyA, QuarterKey: keyB, SourceWatermark: keyC, Version: 1, WindowStartedAt: time.Unix(1, 0), WindowEndedAt: time.Unix(2, 0),
		Cohorts:                []CohortAggregate{{CohortKey: keyA, Eligible: 100, Exposed: 50}, {CohortKey: keyB, Eligible: 100, Exposed: 55}},
		PreviousRegretEligible: 1000, PreviousRegretReports: 20, CurrentRegretEligible: 1000, CurrentRegretReports: 10,
		ColorismAuditComplete: true, CompleteMetrics: []Metric{MetricExposureParity, MetricColorismDrift, MetricRegretTrend, MetricTierASafety}}
}
func result(report Report, metric Metric) Result {
	for _, r := range report.Results {
		if r.Metric == metric {
			return r
		}
	}
	return Result{}
}

func TestQuarterlyAggregatePassUsesExplicitDenominatorsAndThreshold(t *testing.T) {
	report, err := Evaluate(keyD, definition(t, 50), completeSnapshot(), time.Unix(3, 0))
	if err != nil || report.Outcome != OutcomePass || report.ObservedParityGapPermille != 50 || report.MaxParityGapPermille != 50 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	regret := result(report, MetricRegretTrend)
	if regret.Numerator != 10 || regret.Denominator != 1000 || regret.ObservedPermille != 10 || regret.ComparisonPermille != 20 {
		t.Fatalf("regret=%+v", regret)
	}
	if _, err = RehydrateReport(report); err != nil {
		t.Fatal(err)
	}
}
func TestEveryFailureClassIsExplicit(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Snapshot)
		metric Metric
	}{
		{"parity", func(s *Snapshot) { s.Cohorts[1].Exposed = 56 }, MetricExposureParity},
		{"colorism", func(s *Snapshot) { s.ColorismDriftDetected = true }, MetricColorismDrift},
		{"regret", func(s *Snapshot) { s.CurrentRegretReports = 20 }, MetricRegretTrend},
		{"tier-a", func(s *Snapshot) { s.UnresolvedTierA = 1 }, MetricTierASafety},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := completeSnapshot()
			tc.mutate(&s)
			report, err := Evaluate(keyD, definition(t, 50), s, time.Unix(3, 0))
			if err != nil || report.Outcome != OutcomeFail || result(report, tc.metric).State != StateFail {
				t.Fatalf("report=%+v err=%v", report, err)
			}
		})
	}
}
func TestIncompleteEvidenceNeverPasses(t *testing.T) {
	s := completeSnapshot()
	s.CompleteMetrics = []Metric{MetricExposureParity, MetricRegretTrend, MetricTierASafety}
	report, err := Evaluate(keyD, definition(t, 50), s, time.Unix(3, 0))
	if err != nil || report.Outcome != OutcomeIncomplete || result(report, MetricColorismDrift).State != StateIncomplete {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}
func TestSmallCohortsRejected(t *testing.T) {
	s := completeSnapshot()
	s.Cohorts[0].Eligible = 49
	s.Cohorts[0].Exposed = 20
	if _, err := Evaluate(keyD, definition(t, 50), s, time.Unix(3, 0)); err == nil {
		t.Fatal("expected privacy threshold rejection")
	}
}
func TestEvaluationIsOrderIndependentProperty(t *testing.T) {
	base := completeSnapshot()
	expected, _ := Evaluate(keyD, definition(t, 50), base, time.Unix(3, 0))
	for seed := int64(0); seed < 1000; seed++ {
		s := base
		s.Cohorts = append([]CohortAggregate(nil), base.Cohorts...)
		s.CompleteMetrics = append([]Metric(nil), base.CompleteMetrics...)
		r := rand.New(rand.NewSource(seed))
		r.Shuffle(len(s.Cohorts), func(i, j int) { s.Cohorts[i], s.Cohorts[j] = s.Cohorts[j], s.Cohorts[i] })
		r.Shuffle(len(s.CompleteMetrics), func(i, j int) {
			s.CompleteMetrics[i], s.CompleteMetrics[j] = s.CompleteMetrics[j], s.CompleteMetrics[i]
		})
		got, err := Evaluate(keyD, definition(t, 50), s, time.Unix(3, 0))
		if err != nil || got.Fingerprint != expected.Fingerprint || !slices.Equal(got.Results, expected.Results) {
			t.Fatalf("seed=%d err=%v", seed, err)
		}
	}
}
func FuzzCohortBounds(f *testing.F) {
	f.Add(uint64(50), uint64(25))
	f.Add(uint64(49), uint64(1))
	f.Add(uint64(100), uint64(101))
	f.Fuzz(func(t *testing.T, eligible, exposed uint64) {
		s := completeSnapshot()
		s.Cohorts[0].Eligible = eligible
		s.Cohorts[0].Exposed = exposed
		_, err := Evaluate(keyD, definition(t, 50), s, time.Unix(3, 0))
		valid := eligible >= MinCohortSize && eligible <= MaxAggregateCount && exposed <= eligible
		if (err == nil) != valid {
			t.Fatalf("eligible=%d exposed=%d err=%v", eligible, exposed, err)
		}
	})
}
