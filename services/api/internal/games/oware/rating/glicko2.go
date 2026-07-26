package rating

import (
	"errors"
	"math"
	"slices"
)

const (
	glickoScale      = 173.7178
	defaultRating    = 1500.0
	maxPeriodResults = 500
	convergence      = 0.000001
)

var (
	ErrInvalidRating = errors.New("invalid Glicko-2 rating")
	ErrInvalidResult = errors.New("invalid Glicko-2 result")
	ErrInvalidPeriod = errors.New("invalid Glicko-2 rating period")
)

// Rating is the public Glicko-2 scale: rating, rating deviation and
// volatility. A new unrated player conventionally starts at 1500/350/0.06.
type Rating struct {
	Value      float64
	Deviation  float64
	Volatility float64
}

func NewPlayer() Rating {
	return Rating{Value: defaultRating, Deviation: 350, Volatility: 0.06}
}

// Result is one game in a rating period. Score is 0, 0.5, or 1 from the
// rated player's perspective.
type Result struct {
	Opponent Rating
	Score    float64
}

// UpdatePeriod applies one standards-correct Glicko-2 rating period.
// tau is the system volatility constraint; 0.5 is the reference-paper value.
func UpdatePeriod(player Rating, results []Result, tau float64) (Rating, error) {
	if !validRating(player) || !finite(tau) || tau <= 0 || tau > 1.2 || len(results) > maxPeriodResults {
		return Rating{}, ErrInvalidPeriod
	}
	ordered := slices.Clone(results)
	for _, result := range ordered {
		if !validRating(result.Opponent) || result.Score != 0 && result.Score != 0.5 && result.Score != 1 {
			return Rating{}, ErrInvalidResult
		}
	}
	// Stable ordering makes a period deterministic even when callers load
	// games from unordered storage.
	slices.SortFunc(ordered, func(a, b Result) int {
		if comparison := compareFloat(a.Opponent.Value, b.Opponent.Value); comparison != 0 {
			return comparison
		}
		if comparison := compareFloat(a.Opponent.Deviation, b.Opponent.Deviation); comparison != 0 {
			return comparison
		}
		if comparison := compareFloat(a.Opponent.Volatility, b.Opponent.Volatility); comparison != 0 {
			return comparison
		}
		return compareFloat(a.Score, b.Score)
	})

	mu := (player.Value - defaultRating) / glickoScale
	phi := player.Deviation / glickoScale
	if len(ordered) == 0 {
		deviation := glickoScale * math.Sqrt(phi*phi+player.Volatility*player.Volatility)
		return Rating{Value: player.Value, Deviation: deviation, Volatility: player.Volatility}, nil
	}

	information, improvement := 0.0, 0.0
	for _, result := range ordered {
		opponentMu := (result.Opponent.Value - defaultRating) / glickoScale
		opponentPhi := result.Opponent.Deviation / glickoScale
		weight := g(opponentPhi)
		expected := expectation(mu, opponentMu, opponentPhi)
		information += weight * weight * expected * (1 - expected)
		improvement += weight * (result.Score - expected)
	}
	variance := 1 / information
	delta := variance * improvement

	volatility, err := nextVolatility(phi, player.Volatility, variance, delta, tau)
	if err != nil {
		return Rating{}, err
	}
	preRatingPhi := math.Sqrt(phi*phi + volatility*volatility)
	nextPhi := 1 / math.Sqrt(1/(preRatingPhi*preRatingPhi)+1/variance)
	nextMu := mu + nextPhi*nextPhi*improvement
	next := Rating{
		Value:      defaultRating + glickoScale*nextMu,
		Deviation:  glickoScale * nextPhi,
		Volatility: volatility,
	}
	if !validRating(next) {
		return Rating{}, ErrInvalidRating
	}
	return next, nil
}

func nextVolatility(phi, sigma, variance, delta, tau float64) (float64, error) {
	a := math.Log(sigma * sigma)
	objective := func(x float64) float64 {
		exponential := math.Exp(x)
		numerator := exponential * (delta*delta - phi*phi - variance - exponential)
		denominator := 2 * math.Pow(phi*phi+variance+exponential, 2)
		return numerator/denominator - (x-a)/(tau*tau)
	}

	lower := a
	var upper float64
	if delta*delta > phi*phi+variance {
		upper = math.Log(delta*delta - phi*phi - variance)
	} else {
		for step := 1; ; step++ {
			upper = a - float64(step)*tau
			if objective(upper) >= 0 {
				break
			}
			if step > 1000 {
				return 0, ErrInvalidPeriod
			}
		}
	}

	fLower, fUpper := objective(lower), objective(upper)
	for math.Abs(upper-lower) > convergence {
		candidate := lower + (lower-upper)*fLower/(fUpper-fLower)
		fCandidate := objective(candidate)
		if fCandidate*fUpper <= 0 {
			lower, fLower = upper, fUpper
		} else {
			fLower /= 2
		}
		upper, fUpper = candidate, fCandidate
	}
	next := math.Exp(lower / 2)
	if !finite(next) || next <= 0 {
		return 0, ErrInvalidPeriod
	}
	return next, nil
}

func g(phi float64) float64 {
	return 1 / math.Sqrt(1+3*phi*phi/(math.Pi*math.Pi))
}

func expectation(mu, opponentMu, opponentPhi float64) float64 {
	return 1 / (1 + math.Exp(-g(opponentPhi)*(mu-opponentMu)))
}

func validRating(value Rating) bool {
	return finite(value.Value) && finite(value.Deviation) && finite(value.Volatility) &&
		value.Deviation > 0 && value.Volatility > 0
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func compareFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
