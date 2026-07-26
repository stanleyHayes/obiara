package domain

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC)

func rule(t *testing.T) Rule {
	t.Helper()
	r, e := NewRule("reviewed.vouch", 2, ShapeVouchRing, 4, 4, .20, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", now.Add(-time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestGraphAggregateIsDeterministicAcrossEdgeOrder(t *testing.T) {
	edges := [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}
	build := func(v [][2]int) Aggregate {
		a, e := NewAggregate("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "reviewed.vouch", 2, ShapeVouchRing, 300, 2000, 4, len(v), .99, false, false)
		if e != nil {
			t.Fatal(e)
		}
		return a
	}
	a := build(edges)
	for i, j := 0, len(edges)-1; i < j; i, j = i+1, j-1 {
		edges[i], edges[j] = edges[j], edges[i]
	}
	b := build(edges)
	if a.Density != b.Density {
		t.Fatalf("density changed: %v %v", a.Density, b.Density)
	}
	d, e := Evaluate(rule(t), a, now)
	if e != nil || d.Outcome != OutcomeHumanReview {
		t.Fatalf("decision=%+v err=%v", d, e)
	}
}
func TestReviewedShapeAllowlist(t *testing.T) {
	for _, shape := range []Shape{ShapeSyndicate, ShapeVouchRing, ShapeDeviceAnomaly} {
		if _, err := NewRule("reviewed.shape", 1, shape, 3, 3, .2, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", now); err != nil {
			t.Fatalf("%s: %v", shape, err)
		}
	}
	if _, err := NewRule("reviewed.shape", 1, Shape("unknown"), 3, 3, .2, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", now); err == nil {
		t.Fatal("unknown shape accepted")
	}
}
func TestWeakEvidenceNeverRoutesProperty(t *testing.T) {
	r := rand.New(rand.NewSource(817))
	definition := rule(t)
	for i := 0; i < 1000; i++ {
		cohort, evidence, precision := MinCohort+r.Intn(500), MinEvidence+r.Intn(5000), MinPrecision+r.Float64()*(1-MinPrecision)
		switch r.Intn(3) {
		case 0:
			cohort = r.Intn(MinCohort)
		case 1:
			evidence = r.Intn(MinEvidence)
		case 2:
			precision = r.Float64() * MinPrecision
		}
		a, err := NewAggregate("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "reviewed.vouch", 2, ShapeVouchRing, cohort, evidence, 4, 4, precision, false, false)
		if err != nil {
			t.Fatal(err)
		}
		d, err := Evaluate(definition, a, now)
		if err != nil || d.Outcome != OutcomeNoAction {
			t.Fatalf("case %d routed %+v err=%v", i, d, err)
		}
	}
}
func TestUncertaintyAndModelErrorsFailClosed(t *testing.T) {
	r := rule(t)
	for _, flags := range [][2]bool{{true, false}, {false, true}, {true, true}} {
		a, _ := NewAggregate("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", r.Key, r.Version, r.Shape, 300, 2000, 4, 4, .99, flags[0], flags[1])
		d, e := Evaluate(r, a, now)
		if e != nil || d.Outcome != OutcomeNoAction {
			t.Fatalf("decision=%+v err=%v", d, e)
		}
	}
}
func FuzzAggregateRejectsInvalidPrecision(f *testing.F) {
	f.Add(.99)
	f.Add(math.NaN())
	f.Add(math.Inf(1))
	f.Add(-1.0)
	f.Fuzz(func(t *testing.T, p float64) {
		a, err := NewAggregate("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "reviewed.vouch", 2, ShapeVouchRing, 300, 2000, 4, 4, p, false, false)
		valid := !math.IsNaN(p) && !math.IsInf(p, 0) && p >= 0 && p <= 1
		if valid != (err == nil) {
			t.Fatalf("p=%v valid=%v err=%v aggregate=%+v", p, valid, err, a)
		}
	})
}
