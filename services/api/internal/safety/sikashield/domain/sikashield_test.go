package domain

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

var testTime = time.Date(2026, 7, 26, 17, 0, 0, 0, time.UTC)

func reviewedPattern(t *testing.T, source Source) Pattern {
	t.Helper()
	p, err := NewPattern("coercive.payment", 3, source, "pattern:reviewed:3", "reviewer:key", testTime.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEvaluateRoutesOnlyCertainReviewedMatchesToHuman(t *testing.T) {
	p := reviewedPattern(t, SourceText)
	base := Signal{PatternKey: p.Key, PatternVersion: p.Version, Source: p.Source, EvidenceRef: "evidence:aggregate", Confidence: MinPrecision}
	d, err := Evaluate([]Pattern{p}, base, testTime)
	if err != nil || d.Outcome != OutcomeHumanReview {
		t.Fatalf("decision=%+v err=%v", d, err)
	}
	for _, mutate := range []func(*Signal){
		func(s *Signal) { s.ModelError = true },
		func(s *Signal) { s.Uncertain = true },
		func(s *Signal) { s.Confidence = MinPrecision - 0.001 },
		func(s *Signal) { s.PatternVersion++ },
	} {
		s := base
		mutate(&s)
		d, err = Evaluate([]Pattern{p}, s, testTime)
		if err != nil || d.Outcome != OutcomeNoAction {
			t.Fatalf("fail closed decision=%+v err=%v", d, err)
		}
	}
}

func TestMetricThresholdProperty(t *testing.T) {
	r := rand.New(rand.NewSource(812))
	for i := 0; i < 1000; i++ {
		evidence := MinEvidence + r.Intn(10000)
		positive := MinPositive + r.Intn(1000)
		falsePositive := r.Intn(positive + 1)
		truePositive := positive - falsePositive
		reviews := r.Intn(evidence + 1)
		m, err := NewMetrics(evidence, positive, truePositive, falsePositive, reviews)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		want := m.Precision >= MinPrecision && m.ReviewRate <= MaxReviewRate
		if m.Pass() != want {
			t.Fatalf("case %d pass=%v want=%v metrics=%+v", i, m.Pass(), want, m)
		}
	}
}

func FuzzEvaluateNeverRoutesInvalidConfidence(f *testing.F) {
	f.Add(0.99)
	f.Add(math.NaN())
	f.Add(math.Inf(1))
	f.Add(-1.0)
	p, _ := NewPattern("coercive.payment", 1, SourceVoiceMetadata, "pattern:1", "reviewer:key", testTime)
	f.Fuzz(func(t *testing.T, confidence float64) {
		d, err := Evaluate([]Pattern{p}, Signal{PatternKey: p.Key, PatternVersion: 1, Source: p.Source, EvidenceRef: "evidence:metadata", Confidence: confidence}, testTime)
		valid := !math.IsNaN(confidence) && !math.IsInf(confidence, 0) && confidence >= 0 && confidence <= 1
		if !valid && err == nil {
			t.Fatalf("invalid confidence accepted: %v decision=%+v", confidence, d)
		}
		if err == nil && d.Outcome == OutcomeHumanReview && confidence < MinPrecision {
			t.Fatalf("low confidence routed: %v", confidence)
		}
	})
}
