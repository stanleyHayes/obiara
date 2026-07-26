package domain

import (
	"testing"
	"time"
)

var prefsNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestDefaultsAllowEverythingExceptQuietHours(t *testing.T) {
	preferences, err := NewPreferences("m-1", prefsNow)
	if err != nil {
		t.Fatal(err)
	}
	// 12:00 UTC == 12:00 Africa/Accra (UTC+0).
	if decision := preferences.Allows(CategoryRitual, prefsNow); !decision.Allowed {
		t.Fatalf("midday ritual = %v, want allowed", decision)
	}
	// 22:30 local is inside 21:00-07:00.
	late := prefsNow.Add(10*time.Hour + 30*time.Minute)
	if decision := preferences.Allows(CategoryRitual, late); decision.Allowed || decision.Reason != "quiet_hours" {
		t.Fatalf("quiet hours = %v, want suppressed quiet_hours", decision)
	}
	// 08:00 local is outside.
	morning := prefsNow.Add(20 * time.Hour)
	if decision := preferences.Allows(CategoryPods, morning); !decision.Allowed {
		t.Fatalf("morning = %v, want allowed", decision)
	}
}

func TestSafetyBreaksThrough(t *testing.T) {
	preferences, _ := NewPreferences("m-1", prefsNow)
	if err := preferences.Configure(map[Category]bool{CategoryRitual: true}, 0, 24*60, "Africa/Accra", prefsNow); err != nil {
		t.Fatal(err)
	}
	late := prefsNow.Add(10 * time.Hour)
	if decision := preferences.Allows(CategorySafety, late); !decision.Allowed || decision.Reason != "safety_override" {
		t.Fatalf("safety = %v, want always allowed", decision)
	}
}

func TestMutedCategorySuppressed(t *testing.T) {
	preferences, _ := NewPreferences("m-1", prefsNow)
	if err := preferences.Configure(map[Category]bool{CategoryPods: true}, 21*60, 7*60, "Africa/Accra", prefsNow); err != nil {
		t.Fatal(err)
	}
	if decision := preferences.Allows(CategoryPods, prefsNow); decision.Allowed || decision.Reason != "muted" {
		t.Fatalf("muted = %v, want suppressed muted", decision)
	}
	if decision := preferences.Allows(CategoryRooms, prefsNow); !decision.Allowed {
		t.Fatalf("unmuted category = %v, want allowed", decision)
	}
}

func TestConfigureValidation(t *testing.T) {
	preferences, _ := NewPreferences("m-1", prefsNow)
	if err := preferences.Configure(map[Category]bool{CategorySafety: true}, 0, 0, "Africa/Accra", prefsNow); err != ErrSafetyCannotBeMuted {
		t.Fatalf("muting safety = %v, want rejected", err)
	}
	if err := preferences.Configure(nil, 0, 0, "Not/AZone", prefsNow); err != ErrInvalidTimezone {
		t.Fatalf("bad timezone = %v, want rejected", err)
	}
	if _, err := NewPreferences(" ", prefsNow); err != ErrMemberIDRequired {
		t.Fatalf("blank member = %v", err)
	}
}

func TestQuietWindowMidnightCrossing(t *testing.T) {
	for name, tc := range map[string]struct {
		minutes, start, end int
		want                bool
	}{
		"inside crossing window":  {23 * 60, 21 * 60, 7 * 60, true},
		"early morning inside":    {5 * 60, 21 * 60, 7 * 60, true},
		"midday outside":          {12 * 60, 21 * 60, 7 * 60, false},
		"start boundary included": {21 * 60, 21 * 60, 7 * 60, true},
		"end boundary excluded":   {7 * 60, 21 * 60, 7 * 60, false},
		"zero window never quiet": {3 * 60, 0, 0, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := inWindow(tc.minutes, tc.start, tc.end); got != tc.want {
				t.Fatalf("inWindow(%d, %d, %d) = %v, want %v", tc.minutes, tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestLocalDateKey(t *testing.T) {
	preferences, _ := NewPreferences("m-1", prefsNow)
	// 23:30 UTC is 23:30 in Accra (UTC+0) but the next day in Apia (UTC+13).
	if date := preferences.LocalDate(prefsNow.Add(11*time.Hour + 30*time.Minute)); date != "2026-07-26" {
		t.Fatalf("accra date = %q", date)
	}
	if err := preferences.Configure(nil, 21*60, 7*60, "Pacific/Apia", prefsNow); err != nil {
		t.Fatal(err)
	}
	if date := preferences.LocalDate(prefsNow.Add(11*time.Hour + 30*time.Minute)); date != "2026-07-27" {
		t.Fatalf("apia date = %q, want next day", date)
	}
}
