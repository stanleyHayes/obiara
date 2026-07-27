// Package application computes the P0 phase-exit metrics (E15-S07; plan
// §22 gates) from the analytics pipeline: funnel rates, fire attendance
// against the active cohort, and the regret trend. Queries are read-only
// and pseudonymous throughout.
package application

import (
	"context"
	"time"
)

// EventCounts is the analytics read port.
type EventCounts interface {
	CountEvents(ctx context.Context, name string, since time.Time) (int, error)
	CountDistinctSubjects(ctx context.Context, name string, since time.Time) (int, error)
}

// ActiveCohorts reports the active cohort size (identity read port).
type ActiveCohorts interface {
	CountActive(ctx context.Context) (int, error)
}

// FunnelReport is the P0 exit-metric snapshot (plan §22).
type FunnelReport struct {
	WindowDays         int       `json:"windowDays"`
	PodsHeardRate      float64   `json:"podsHeardRate"`
	SeedToSproutRate   float64   `json:"seedToSproutRate"`
	SproutToRoomRate   float64   `json:"sproutToRoomRate"`
	FireAttendeeCount  int       `json:"fireAttendeeCount"`
	FireAttendanceRate float64   `json:"fireAttendanceRate"`
	RegretCount        int       `json:"regretCount"`
	RegretTrend        string    `json:"regretTrend"`
	ComputedAt         time.Time `json:"computedAt"`
}

// MetricsService computes funnel reports.
type MetricsService struct {
	counts  EventCounts
	cohorts ActiveCohorts
	now     func() time.Time
}

func NewMetricsService(counts EventCounts, cohorts ActiveCohorts, now func() time.Time) MetricsService {
	return MetricsService{counts: counts, cohorts: cohorts, now: now}
}

// Funnel computes the report for a trailing window in days.
func (service MetricsService) Funnel(ctx context.Context, windowDays int) (FunnelReport, error) {
	if windowDays < 1 {
		windowDays = 30
	}
	now := service.now().UTC()
	since := now.Add(-time.Duration(windowDays) * 24 * time.Hour)
	report := FunnelReport{WindowDays: windowDays, ComputedAt: now}

	sown, err := service.counts.CountEvents(ctx, "epono.seed_sown", since)
	if err != nil {
		return report, err
	}
	heard, err := service.counts.CountEvents(ctx, "epono.pod_heard", since)
	if err != nil {
		return report, err
	}
	sprouted, err := service.counts.CountEvents(ctx, "epono.sprout_opened", since)
	if err != nil {
		return report, err
	}
	rooms, err := service.counts.CountEvents(ctx, "epono.room_opened", since)
	if err != nil {
		return report, err
	}
	report.PodsHeardRate = rate(heard, sown)
	report.SeedToSproutRate = rate(sprouted, sown)
	report.SproutToRoomRate = rate(rooms, sprouted)

	attendees, err := service.counts.CountDistinctSubjects(ctx, "gyaase.fire_attended", now.Add(-7*24*time.Hour))
	if err != nil {
		return report, err
	}
	report.FireAttendeeCount = attendees
	active, err := service.cohorts.CountActive(ctx)
	if err != nil {
		return report, err
	}
	report.FireAttendanceRate = rate(attendees, active)

	regrets, err := service.counts.CountEvents(ctx, "wellbeing.regret_reported", since)
	if err != nil {
		return report, err
	}
	report.RegretCount = regrets
	prior, err := service.counts.CountEvents(ctx, "wellbeing.regret_reported", since.Add(-time.Duration(windowDays)*24*time.Hour))
	if err != nil {
		return report, err
	}
	report.RegretTrend = trend(regrets, prior-regrets)

	return report, nil
}

func rate(part, whole int) float64 {
	if whole <= 0 {
		return 0
	}
	return float64(part) / float64(whole)
}

// trend compares the current window count to the prior window count.
func trend(current, previous int) string {
	switch {
	case current > previous:
		return "up"
	case current < previous:
		return "down"
	default:
		return "flat"
	}
}
