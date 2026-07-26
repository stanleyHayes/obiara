// Package ritual wires the notification ritual dispatcher into the worker
// scheduler (E13-S02): calendar rituals (dawn, Monday, Sunday) and fire
// heralds dispatch through preferences and the durable outbox, exactly
// once per member per day.
package ritual

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/ritual/application"
	jobsapplication "github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

// NewCalendarJob builds the scheduled calendar-ritual dispatch job.
func NewCalendarJob(dispatcher application.Dispatcher, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "ritual.calendar",
		Interval:    interval,
		Timeout:     2 * time.Minute,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			return dispatcher.DispatchCalendar(ctx)
		},
	}
}

// NewHeraldJob builds the scheduled fire-herald dispatch job.
func NewHeraldJob(dispatcher application.Dispatcher, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "ritual.fire_herald",
		Interval:    interval,
		Timeout:     2 * time.Minute,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			return dispatcher.DispatchHeralds(ctx)
		},
	}
}
