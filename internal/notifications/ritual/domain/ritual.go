// Package domain defines the notification ritual calendar (E13-S02): dawn
// summary every morning, Monday intention, Sunday reflection, and the fire
// herald before a fire starts. Rituals are evaluated in the member's own
// time zone and dispatched through preferences (E13-S01) exactly once per
// member per day.
package domain

import (
	"time"
)

// Kind is a ritual class.
type Kind string

const (
	KindDawn       Kind = "dawn"
	KindMonday     Kind = "monday_intention"
	KindSunday     Kind = "sunday_reflection"
	KindFireHerald Kind = "fire_herald"
)

// heraldLeadTime is how long before a fire starts the herald dispatches
// (Doc 06 S-61: 19:15 herald for a 20:00 fire).
const heraldLeadTime = 45 * time.Minute

// ClockTime is a member-local time of day.
type ClockTime struct {
	Hour   int
	Minute int
}

var schedule = map[Kind]struct {
	weekday time.Weekday // -1 means every day
	at      ClockTime
}{
	KindDawn:   {weekday: -1, at: ClockTime{Hour: 6, Minute: 0}},
	KindMonday: {weekday: time.Monday, at: ClockTime{Hour: 8, Minute: 0}},
	KindSunday: {weekday: time.Sunday, at: ClockTime{Hour: 18, Minute: 0}},
}

// CalendarKinds lists the rituals dispatched on a clock schedule.
func CalendarKinds() []Kind {
	return []Kind{KindDawn, KindMonday, KindSunday}
}

// DueAt returns the member-local dispatch time for a ritual on a date.
// ok is false for kinds that do not fall on that date.
func DueAt(kind Kind, date time.Time, location *time.Location) (time.Time, bool) {
	entry, ok := schedule[kind]
	if !ok {
		return time.Time{}, false
	}
	local := date.In(location)
	if entry.weekday >= 0 && local.Weekday() != entry.weekday {
		return time.Time{}, false
	}
	return time.Date(local.Year(), local.Month(), local.Day(), entry.at.Hour, entry.at.Minute, 0, 0, location), true
}

// HeraldAt returns the dispatch time for a fire's herald.
func HeraldAt(fireStartsAt time.Time) time.Time {
	return fireStartsAt.Add(-heraldLeadTime)
}

// DedupKey identifies one ritual delivery for one member.
func DedupKey(memberID string, kind Kind, discriminator string) string {
	return memberID + "|" + string(kind) + "|" + discriminator
}
