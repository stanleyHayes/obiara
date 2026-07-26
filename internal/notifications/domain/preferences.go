// Package domain models notification preferences and the delivery
// decision rules (E13-S01): member-controlled mutes, quiet hours in the
// member's own time zone, and the six-per-day cap. Safety and care
// notifications break through every limit — member physical safety
// outranks every other concern (Doc 09 §1). All rules are enforced
// server-side (E13 exit).
package domain

import (
	"errors"
	"strings"
	"time"
)

// DailyCap is the maximum non-safety notifications per member per
// member-local day (E13-S01).
const DailyCap = 6

// Category is a notification class.
type Category string

const (
	CategoryRitual Category = "ritual" // dawn, Monday, fire, Sunday rituals
	CategoryPods   Category = "pods"
	CategoryRooms  Category = "rooms"
	CategorySafety Category = "safety" // never capped, muted or quieted
)

var (
	ErrMemberIDRequired    = errors.New("member id is required")
	ErrInvalidTimezone     = errors.New("timezone must be a valid IANA name")
	ErrSafetyCannotBeMuted = errors.New("safety notifications cannot be muted")
)

// Decision is the delivery ruling with a machine-readable reason.
type Decision struct {
	Allowed bool
	Reason  string
}

// Preferences is a member's notification configuration.
type Preferences struct {
	memberID   string
	muted      map[Category]bool
	quietStart int // minutes from midnight, member-local
	quietEnd   int
	timezone   string
	version    int64
	updatedAt  time.Time
}

// NewPreferences builds preferences with defaults: nothing muted, quiet
// hours 21:00-07:00, Africa/Accra.
func NewPreferences(memberID string, now time.Time) (Preferences, error) {
	if strings.TrimSpace(memberID) == "" {
		return Preferences{}, ErrMemberIDRequired
	}
	return Preferences{
		memberID:   memberID,
		muted:      map[Category]bool{},
		quietStart: 21 * 60,
		quietEnd:   7 * 60,
		timezone:   "Africa/Accra",
		version:    1,
		updatedAt:  now.UTC(),
	}, nil
}

// ReconstitutePreferences rebuilds stored preferences without checks.
func ReconstitutePreferences(memberID string, muted map[Category]bool, quietStart, quietEnd int, timezone string, version int64, updatedAt time.Time) Preferences {
	return Preferences{
		memberID:   memberID,
		muted:      muted,
		quietStart: quietStart,
		quietEnd:   quietEnd,
		timezone:   timezone,
		version:    version,
		updatedAt:  updatedAt,
	}
}

// Configure replaces the configuration after validation.
func (preferences *Preferences) Configure(muted map[Category]bool, quietStart, quietEnd int, timezone string, now time.Time) error {
	if muted[CategorySafety] {
		return ErrSafetyCannotBeMuted
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return ErrInvalidTimezone
	}
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 24*60 {
			return 24 * 60
		}
		return v
	}
	preferences.muted = muted
	preferences.quietStart = clamp(quietStart)
	preferences.quietEnd = clamp(quietEnd)
	preferences.timezone = timezone
	preferences.version++
	preferences.updatedAt = now.UTC()
	return nil
}

// Allows rules on a category at a moment, before the daily cap is applied.
// Quiet hours are evaluated in the member's time zone (agent_plan.md §7.4:
// server time with explicit IANA zones).
func (preferences Preferences) Allows(category Category, now time.Time) Decision {
	if category == CategorySafety {
		return Decision{Allowed: true, Reason: "safety_override"}
	}
	if preferences.muted[category] {
		return Decision{Allowed: false, Reason: "muted"}
	}
	location, err := time.LoadLocation(preferences.timezone)
	if err != nil {
		location = time.UTC
	}
	local := now.UTC().In(location)
	minutes := local.Hour()*60 + local.Minute()
	if inWindow(minutes, preferences.quietStart, preferences.quietEnd) {
		return Decision{Allowed: false, Reason: "quiet_hours"}
	}
	return Decision{Allowed: true, Reason: "allowed"}
}

// LocalDate returns the member-local date key for daily-cap accounting.
func (preferences Preferences) LocalDate(now time.Time) string {
	location, err := time.LoadLocation(preferences.timezone)
	if err != nil {
		location = time.UTC
	}
	return now.UTC().In(location).Format("2006-01-02")
}

// inWindow reports whether minutes-from-midnight falls inside the window,
// which may cross midnight (e.g. 21:00-07:00).
func inWindow(minutes, start, end int) bool {
	if start == end {
		return false
	}
	if start < end {
		return minutes >= start && minutes < end
	}
	return minutes >= start || minutes < end
}

func (preferences Preferences) MemberID() string         { return preferences.memberID }
func (preferences Preferences) Muted() map[Category]bool { return preferences.muted }
func (preferences Preferences) QuietStart() int          { return preferences.quietStart }
func (preferences Preferences) QuietEnd() int            { return preferences.quietEnd }
func (preferences Preferences) Timezone() string         { return preferences.timezone }
func (preferences Preferences) Version() int64           { return preferences.version }
func (preferences Preferences) UpdatedAt() time.Time     { return preferences.updatedAt }
