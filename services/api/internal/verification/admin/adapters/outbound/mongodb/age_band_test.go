package mongodb

import (
	"testing"
	"time"
)

func TestAgeBandDoesNotInventAnAgeItNoLongerHolds(t *testing.T) {
	// Retention strips dateOfBirth from decided cases, so the reviewer desk
	// will meet absent birth dates as a matter of course. Computing from the
	// zero time yields an age of about two thousand and reports "50_plus" —
	// a confident, wrong fact on the screen of someone making a decision
	// about a real person.
	if band := ageBand(time.Time{}, time.Date(2026, time.September, 5, 0, 0, 0, 0, time.UTC)); band != "unknown" {
		t.Fatalf("a stripped date of birth reported %q", band)
	}
}

func TestAgeBandStillBandsTheDatesItDoesHold(t *testing.T) {
	at := time.Date(2026, time.September, 5, 0, 0, 0, 0, time.UTC)
	for band, dateOfBirth := range map[string]time.Time{
		"under_18": time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC),
		"18_24":    time.Date(2004, time.January, 1, 0, 0, 0, 0, time.UTC),
		"25_34":    time.Date(1996, time.January, 1, 0, 0, 0, 0, time.UTC),
		"35_49":    time.Date(1986, time.January, 1, 0, 0, 0, 0, time.UTC),
		"50_plus":  time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC),
	} {
		if got := ageBand(dateOfBirth, at); got != band {
			t.Fatalf("%s dated %s, want %s", got, dateOfBirth.Format(time.DateOnly), band)
		}
	}
}
