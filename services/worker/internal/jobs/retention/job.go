// Package retention wires the retention runner into the worker scheduler
// (E15-S08).
package retention

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/retention"
	jobsapplication "github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

// NewJob builds the scheduled retention job.
func NewJob(runner *retention.Runner, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "retention.runner",
		Interval:    interval,
		Timeout:     5 * time.Minute,
		MaxAttempts: 3,
		Run: func(ctx context.Context) error {
			_, err := runner.RunOnce(ctx)
			return err
		},
	}
}
