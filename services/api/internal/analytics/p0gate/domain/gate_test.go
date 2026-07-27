package domain

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)

func definition(t *testing.T) Definition {
	t.Helper()
	d, e := NewDefinition(DefinitionSpec{ID: "p0.gates", Version: 1, ReviewID: "founder.gates", ReviewerKey: key(1), ReviewedAt: now, PodsHeardPermille: 650, SeedToSproutPermille: 250, SproutToRoomPermille: 350, WeeklyFirePermille: 400, Day30RetentionPermille: 450})
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func snapshot() Snapshot {
	return Snapshot{ID: key(2), WindowKey: key(3), SourceWatermark: key(4), Version: 1, WindowStartedAt: now.Add(-30 * 24 * time.Hour), WindowEndedAt: now, CohortSize: 300, PodEligible: 100, PodsHeard: 65, SeedsSown: 100, SproutsOpened: 25, SproutEligible: 100, RoomsOpened: 35, WeeklyFireAttendees: 120, Day30Eligible: 100, Day30Retained: 45, PreviousRegretReports: 10, CurrentRegretReports: 9, UnresolvedTierA: 0, CompleteMetrics: []Metric{MetricPodsHeard, MetricSeedToSprout, MetricSproutToRoom, MetricWeeklyFire, MetricDay30Retention, MetricRegretTrend, MetricTierAResolved}}
}
func TestExactFounderGatesPass(t *testing.T) {
	r, e := Evaluate(key(5), definition(t), snapshot(), now)
	if e != nil || r.Outcome != OutcomePass {
		t.Fatal(r, e)
	}
	if len(r.Results) != 7 {
		t.Fatal("missing gate")
	}
}
func TestAnyMissingEvidencePreventsPass(t *testing.T) {
	for _, metric := range snapshot().CompleteMetrics {
		s := snapshot()
		for i, x := range s.CompleteMetrics {
			if x == metric {
				s.CompleteMetrics = append(s.CompleteMetrics[:i], s.CompleteMetrics[i+1:]...)
				break
			}
		}
		r, e := Evaluate(key(5), definition(t), s, now)
		if e != nil || r.Outcome != OutcomeIncomplete {
			t.Fatalf("%s %+v %v", metric, r, e)
		}
	}
}
func TestEveryFailedGatePreventsPass(t *testing.T) {
	tests := []func(*Snapshot){func(s *Snapshot) { s.PodsHeard-- }, func(s *Snapshot) { s.SproutsOpened-- }, func(s *Snapshot) { s.RoomsOpened-- }, func(s *Snapshot) { s.WeeklyFireAttendees-- }, func(s *Snapshot) { s.Day30Retained-- }, func(s *Snapshot) { s.CurrentRegretReports = s.PreviousRegretReports }, func(s *Snapshot) { s.UnresolvedTierA = 1 }}
	for i, fail := range tests {
		s := snapshot()
		fail(&s)
		r, e := Evaluate(key(5), definition(t), s, now)
		if e != nil || r.Outcome != OutcomeFail {
			t.Fatalf("%d %+v %v", i, r, e)
		}
	}
}
func TestCompleteMetricOrderingDoesNotChangeProjection(t *testing.T) {
	s := snapshot()
	want, _ := Evaluate(key(5), definition(t), s, now)
	random := rand.New(rand.NewSource(42))
	for range 1000 {
		random.Shuffle(len(s.CompleteMetrics), func(i, j int) {
			s.CompleteMetrics[i], s.CompleteMetrics[j] = s.CompleteMetrics[j], s.CompleteMetrics[i]
		})
		got, e := Evaluate(key(5), definition(t), s, now)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("order changed projection")
		}
	}
}
func FuzzRatesNeverFalsePass(f *testing.F) {
	f.Add(uint64(64), uint64(100))
	f.Fuzz(func(t *testing.T, heard, eligible uint64) {
		eligible %= 1_000_000
		if eligible == 0 {
			eligible = 1
		}
		heard %= eligible + 1
		s := snapshot()
		s.PodEligible = eligible
		s.PodsHeard = heard
		r, e := Evaluate(key(5), definition(t), s, now)
		if e != nil {
			t.Fatal(e)
		}
		var state MetricState
		for _, x := range r.Results {
			if x.Metric == MetricPodsHeard {
				state = x.State
			}
		}
		expected := heard*1000/eligible >= 650
		if expected != (state == StatePass) {
			t.Fatalf("%d/%d %s", heard, eligible, state)
		}
	})
}
