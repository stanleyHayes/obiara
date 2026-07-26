package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAgeAtBirthdayBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		birth time.Time
		now   time.Time
		want  int
	}{
		{"day before eighteenth", time.Date(2008, 7, 27, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 26, 23, 59, 59, 0, time.UTC), 17},
		{"on eighteenth", time.Date(2008, 7, 26, 20, 0, 0, 0, time.UTC), time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC), 18},
		{"leap birthday before march", time.Date(2008, 2, 29, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC), 17},
		{"leap birthday after march", time.Date(2008, 2, 29, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), 18},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AgeAt(test.birth, test.now)
			if err != nil || got != test.want {
				t.Fatalf("AgeAt() = %d, %v; want %d", got, err, test.want)
			}
		})
	}
}

func TestRestrictionIsAlwaysBlockedAndProofIsImmutable(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	restriction, err := NewRestriction("restriction:1", "command:1", digest, strings.Repeat("b", 64), now)
	if err != nil {
		t.Fatal(err)
	}
	if !restriction.Blocked() || restriction.PurgeDueAt() != now.Add(24*time.Hour) {
		t.Fatal("underage restriction must hard block with an exact 24-hour deadline")
	}
	purged, err := restriction.MarkPurged(now.Add(time.Hour), 1)
	if err != nil {
		t.Fatal(err)
	}
	if restriction.PurgeStatus() != PurgePending || !purged.PurgedWithinSLA() || !purged.Blocked() {
		t.Fatal("purge must preserve the original hard block and immutable value")
	}
	if _, err := restriction.MarkPurged(now, 2); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("expected stale version, got %v", err)
	}
}

func FuzzAgeGateNeverAllowsUnder18(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(365 * 17))
	f.Fuzz(func(t *testing.T, days int64) {
		if days < 0 || days > 365*30 {
			t.Skip()
		}
		now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
		birth := now.Add(-time.Duration(days) * 24 * time.Hour)
		age, err := AgeAt(birth, now)
		eligible, eligibilityErr := Eligible(birth, now)
		if err == nil && eligibilityErr == nil && age < MinimumAge && eligible {
			t.Fatal("under-18 age was allowed")
		}
	})
}
