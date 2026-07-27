// Package application computes per-channel delivery statistics (E13-S08):
// what was attempted, what landed, and what failed, per channel and
// template. Numbers come only from the delivery logs — never message
// content.
package application

import (
	"context"
	"time"
)

// DeliveryCounts is the cross-channel read port.
type DeliveryCounts interface {
	CountByStatus(ctx context.Context, collection string, since time.Time) (map[string]int, error)
}

// ChannelStats is one channel's delivery outcome.
type ChannelStats struct {
	Attempted   int     `json:"attempted"`
	Sent        int     `json:"sent"`
	Delivered   int     `json:"delivered"`
	Failed      int     `json:"failed"`
	SuccessRate float64 `json:"successRate"`
}

// StatsReport aggregates channels over a window.
type StatsReport struct {
	WindowDays int                     `json:"windowDays"`
	Channels   map[string]ChannelStats `json:"channels"`
	ComputedAt time.Time               `json:"computedAt"`
}

// collections maps channels to their delivery-log collections.
var collections = map[string]string{
	"whatsapp": "whatsapp_deliveries",
	"email":    "email_deliveries",
}

// StatsService computes delivery statistics.
type StatsService struct {
	counts DeliveryCounts
	now    func() time.Time
}

func NewStatsService(counts DeliveryCounts, now func() time.Time) StatsService {
	return StatsService{counts: counts, now: now}
}

// Stats aggregates the window.
func (service StatsService) Stats(ctx context.Context, windowDays int) (StatsReport, error) {
	if windowDays < 1 {
		windowDays = 30
	}
	since := service.now().UTC().Add(-time.Duration(windowDays) * 24 * time.Hour)
	report := StatsReport{
		WindowDays: windowDays,
		Channels:   map[string]ChannelStats{},
		ComputedAt: service.now().UTC(),
	}

	for channel, collection := range collections {
		byStatus, err := service.counts.CountByStatus(ctx, collection, since)
		if err != nil {
			return report, err
		}
		stats := ChannelStats{
			Sent:      byStatus["sent"],
			Delivered: byStatus["delivered"],
			Failed:    byStatus["failed"] + byStatus["bounced"] + byStatus["complained"],
		}
		stats.Attempted = stats.Sent + stats.Delivered + stats.Failed
		if stats.Attempted > 0 {
			stats.SuccessRate = float64(stats.Sent+stats.Delivered) / float64(stats.Attempted)
		}
		report.Channels[channel] = stats
	}
	return report, nil
}
