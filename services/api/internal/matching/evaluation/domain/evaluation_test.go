package domain

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)

func cmd(id string, revision uint64, at time.Time) Command {
	return Command{ID: id, ExpectedRevision: revision, At: at}
}
func goodMetrics() Metrics {
	return Metrics{Cohort: 500, Evidence: 4000, Quality: .79, ErrorRate: .18, MaxDisparity: .04, Slices: []SliceMetric{{PolicyKey: "approved.region", Cohort: 250, Quality: .78, ErrorRate: .19}, {PolicyKey: "approved.access", Cohort: 250, Quality: .80, ErrorRate: .17}}}
}
func readyEvaluation(t *testing.T) Evaluation {
	t.Helper()
	e, err := Create("evaluation:1", "candidate.compatibility", 1, cmd("create", 0, now))
	if err != nil {
		t.Fatal(err)
	}
	e, err = e.Record(Snapshot{ID: "snapshot:1", Version: 2, ConsentVersion: 7, EvaluatedAt: now}, goodMetrics(), cmd("evaluate", 1, now))
	if err != nil {
		t.Fatal(err)
	}
	e, err = e.AttachCard(ModelCard{Version: 1, Purpose: "matching.compatibility", EvaluationRef: "report:1", LimitationsRef: "limits:1", Owner: "matching.team"}, cmd("card", 2, now))
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func TestFailsClosedUntilEveryGateAndFreshHumanApproval(t *testing.T) {
	e, err := Create("evaluation:1", "candidate.compatibility", 1, cmd("create", 0, now))
	if err != nil || e.Ready(now) {
		t.Fatal("draft ready")
	}
	e = readyEvaluation(t)
	if e.Ready(now) {
		t.Fatal("unapproved ready")
	}
	e, err = e.Approve("reviewer:key", now.Add(24*time.Hour), cmd("approve", 3, now))
	if err != nil || !e.Ready(now) {
		t.Fatalf("approved=%v err=%v", e.Ready(now), err)
	}
	if e.Ready(now.Add(24 * time.Hour)) {
		t.Fatal("expired approval ready")
	}
}
func TestWeakAndSmallCohortsAlwaysFailProperty(t *testing.T) {
	r := rand.New(rand.NewSource(77))
	base := readyEvaluation(t)
	_ = base
	for i := 0; i < 1000; i++ {
		m := goodMetrics()
		switch r.Intn(5) {
		case 0:
			m.Cohort = r.Intn(MinCohort)
		case 1:
			m.Evidence = r.Intn(MinEvidence)
		case 2:
			m.Quality = r.Float64() * MinQuality
		case 3:
			m.MaxDisparity = MaxDisparity + r.Float64()
		case 4:
			m.Slices[0].Cohort = r.Intn(MinSliceCohort)
		}
		e, _ := Create(fmt.Sprintf("evaluation:%d", i+10), "candidate.compatibility", 1, cmd(fmt.Sprintf("create:%d", i), 0, now))
		if _, err := e.Record(Snapshot{ID: "snapshot:x", Version: 1, ConsentVersion: 1, EvaluatedAt: now}, m, cmd(fmt.Sprintf("eval:%d", i), 1, now)); err == nil {
			t.Fatalf("weak case %d accepted: %+v", i, m)
		}
	}
}
func FuzzMetricsNeverAcceptNonFiniteOrOutOfBounds(f *testing.F) {
	f.Add(0.7, 0.2, 0.05)
	f.Add(-1.0, 2.0, 4.0)
	f.Fuzz(func(t *testing.T, quality, errorRate, disparity float64) {
		m := goodMetrics()
		m.Quality, m.ErrorRate, m.MaxDisparity = quality, errorRate, disparity
		valid := validMetrics(m)
		expected := quality >= MinQuality && quality <= MaxQuality && errorRate >= 0 && errorRate <= MaxErrorRate && disparity >= 0 && disparity <= MaxDisparity
		if valid != expected {
			t.Fatalf("valid=%v expected=%v", valid, expected)
		}
	})
}
