package domain

import (
	"math/rand"
	"testing"
	"time"
)

var playbackNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestRecordMergesAndClamps(t *testing.T) {
	playback, err := NewPlayback("m-1", "asset-1", 60)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range [][2]float64{{0, 10}, {5, 15}, {20, 25}, {55, 90}} {
		if err := playback.Record(r[0], r[1]); err != nil {
			t.Fatal(err)
		}
	}
	playback.Commit(playbackNow)

	// {0,10}∪{5,15}={0,15}, {20,25}, {55,90}→{55,60}: total 15+5+5=25.
	if total := playback.TotalSeconds(); total != 25 {
		t.Fatalf("total = %v, want 25", total)
	}
	if len(playback.Intervals()) != 3 {
		t.Fatalf("intervals = %#v", playback.Intervals())
	}
	if !playback.Eligible() {
		t.Fatal("25s >= 20s must be eligible")
	}
}

func TestRecordValidation(t *testing.T) {
	playback, _ := NewPlayback("m-1", "asset-1", 60)
	for _, r := range [][2]float64{{-1, 5}, {10, 10}, {12, 8}} {
		if err := playback.Record(r[0], r[1]); err != ErrInvalidRange {
			t.Fatalf("range %v = %v, want invalid", r, err)
		}
	}
	if err := playback.Record(70, 80); err != ErrInvalidRange {
		t.Fatalf("fully beyond duration = %v, want invalid", err)
	}
	if _, err := NewPlayback("", "asset-1", 60); err != ErrListenerRequired {
		t.Fatalf("missing listener = %v", err)
	}
	if _, err := NewPlayback("m-1", "asset-1", 0); err != ErrInvalidDuration {
		t.Fatalf("zero duration = %v", err)
	}
}

func TestEligibilityThreshold(t *testing.T) {
	playback, _ := NewPlayback("m-1", "asset-1", 120)
	if playback.Eligible() {
		t.Fatal("no playback must not be eligible")
	}
	_ = playback.Record(0, RequiredSeconds-0.5)
	if playback.Eligible() {
		t.Fatal("just below threshold must not be eligible")
	}
	_ = playback.Record(RequiredSeconds-0.5, RequiredSeconds)
	if !playback.Eligible() {
		t.Fatal("exactly 20s must be eligible")
	}
}

// TestMergeIsOrderIndependentAndIdempotent is the FR-202 property: unique
// cumulative playback never depends on heartbeat order or duplication
// (TP-E06-S03-01 in the traceability matrix).
func TestMergeIsOrderIndependentAndIdempotent(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 200; trial++ {
		count := 1 + rng.Intn(12)
		var ranges [][2]float64
		for i := 0; i < count; i++ {
			start := float64(rng.Intn(100))
			end := start + 1 + float64(rng.Intn(20))
			ranges = append(ranges, [2]float64{start, end})
		}

		record := func(seq [][2]float64) Playback {
			playback, err := NewPlayback("m-1", "asset-1", 120)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range seq {
				if err := playback.Record(r[0], r[1]); err != nil {
					t.Fatal(err)
				}
			}
			return playback
		}

		forward := record(ranges)
		reversed := make([][2]float64, len(ranges))
		for i, r := range ranges {
			reversed[len(ranges)-1-i] = r
		}
		backward := record(reversed)
		if forward.TotalSeconds() != backward.TotalSeconds() {
			t.Fatalf("order changed total: %v vs %v (ranges %v)", forward.TotalSeconds(), backward.TotalSeconds(), ranges)
		}

		// Replaying the whole batch changes nothing.
		replayed := forward
		for _, r := range ranges {
			if err := replayed.Record(r[0], r[1]); err != nil {
				t.Fatal(err)
			}
		}
		if replayed.TotalSeconds() != forward.TotalSeconds() {
			t.Fatalf("replay changed total: %v vs %v", replayed.TotalSeconds(), forward.TotalSeconds())
		}
	}
}
