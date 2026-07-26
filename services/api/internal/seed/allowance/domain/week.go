package domain

import (
	"fmt"
	"time"
)

// WeekPolicy defines a week as Monday 00:00 in a named IANA timezone.
type WeekPolicy struct{ location *time.Location }

func NewWeekPolicy(timezone string) (WeekPolicy, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return WeekPolicy{}, fmt.Errorf("load allowance timezone: %w", err)
	}
	return WeekPolicy{location: location}, nil
}

func (p WeekPolicy) Start(at time.Time) time.Time {
	local := at.In(p.location)
	offset := (int(local.Weekday()) + 6) % 7
	date := local.AddDate(0, 0, -offset)
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, p.location).UTC()
}
