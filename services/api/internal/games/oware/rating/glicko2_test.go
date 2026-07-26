package rating

import (
	"math"
	"math/rand"
	"slices"
	"sync"
	"testing"
)

func TestGlicko2ReferenceVector(t *testing.T) {
	player := Rating{Value: 1500, Deviation: 200, Volatility: 0.06}
	results := []Result{
		{Opponent: Rating{Value: 1400, Deviation: 30, Volatility: 0.06}, Score: 1},
		{Opponent: Rating{Value: 1550, Deviation: 100, Volatility: 0.06}, Score: 0},
		{Opponent: Rating{Value: 1700, Deviation: 300, Volatility: 0.06}, Score: 0},
	}
	got, err := UpdatePeriod(player, results, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	// Mark Glickman's Glicko-2 worked example, rounded as published.
	assertNear(t, "rating", got.Value, 1464.06, 0.01)
	assertNear(t, "deviation", got.Deviation, 151.52, 0.01)
	assertNear(t, "volatility", got.Volatility, 0.059996, 0.000001)
}

func TestInactivePeriodOnlyIncreasesDeviation(t *testing.T) {
	player := Rating{Value: 1500, Deviation: 200, Volatility: 0.06}
	got, err := UpdatePeriod(player, nil, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != player.Value || got.Volatility != player.Volatility || got.Deviation <= player.Deviation {
		t.Fatalf("inactive update = %+v", got)
	}
}

func TestRatingPeriodIsIndependentOfResultLoadOrder(t *testing.T) {
	player := Rating{Value: 1720, Deviation: 82, Volatility: 0.055}
	results := []Result{
		{Opponent: Rating{Value: 1400, Deviation: 80, Volatility: 0.06}, Score: 1},
		{Opponent: Rating{Value: 1900, Deviation: 120, Volatility: 0.05}, Score: 0.5},
		{Opponent: Rating{Value: 1650, Deviation: 60, Volatility: 0.04}, Score: 0},
	}
	forward, err := UpdatePeriod(player, results, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(results)
	reversed, err := UpdatePeriod(player, results, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if forward != reversed {
		t.Fatalf("period order changed result: forward=%+v reversed=%+v", forward, reversed)
	}
}

func TestRatingUpdatesRemainFiniteAcrossDeterministicPeriods(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	player := NewPlayer()
	for period := 0; period < 1000; period++ {
		count := rng.Intn(12)
		results := make([]Result, count)
		for index := range results {
			results[index] = Result{
				Opponent: Rating{
					Value:      800 + rng.Float64()*1400,
					Deviation:  20 + rng.Float64()*330,
					Volatility: 0.02 + rng.Float64()*0.12,
				},
				Score: []float64{0, 0.5, 1}[rng.Intn(3)],
			}
		}
		var err error
		player, err = UpdatePeriod(player, results, 0.5)
		if err != nil {
			t.Fatalf("period %d: %v", period, err)
		}
		if !validRating(player) {
			t.Fatalf("period %d produced invalid rating %+v", period, player)
		}
	}
}

func TestPureRatingKernelIsRaceSafe(t *testing.T) {
	player := Rating{Value: 1500, Deviation: 200, Volatility: 0.06}
	results := []Result{{Opponent: NewPlayer(), Score: 1}}
	want, err := UpdatePeriod(player, results, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				got, updateErr := UpdatePeriod(player, results, 0.5)
				if updateErr != nil || got != want {
					t.Errorf("concurrent update = %+v, %v", got, updateErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzRatingPeriodsStayFinite(f *testing.F) {
	f.Add(uint64(1), uint8(12))
	f.Add(uint64(20260726), uint8(50))
	f.Fuzz(func(t *testing.T, seed uint64, count uint8) {
		rng := rand.New(rand.NewSource(int64(seed)))
		player := NewPlayer()
		for range int(count)%100 + 1 {
			result := Result{
				Opponent: Rating{
					Value:      -1000 + rng.Float64()*5000,
					Deviation:  1 + rng.Float64()*999,
					Volatility: 0.001 + rng.Float64(),
				},
				Score: []float64{0, 0.5, 1}[rng.Intn(3)],
			}
			var err error
			player, err = UpdatePeriod(player, []Result{result}, 0.5)
			if err != nil {
				t.Fatal(err)
			}
			if math.IsNaN(player.Value) || math.IsInf(player.Value, 0) || !validRating(player) {
				t.Fatalf("non-finite rating %+v", player)
			}
		}
	})
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %.9f, want %.9f ± %.9f", name, got, want, tolerance)
	}
}
