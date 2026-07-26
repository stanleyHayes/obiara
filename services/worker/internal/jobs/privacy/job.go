// Package privacy wires the data-subject request processor into the
// worker scheduler (S4-005). The job executes open export and deletion
// requests in statutory order (FR-106 clocks: 72 h export, 30 d deletion).
package privacy

import (
	"context"
	"time"

	privacyapplication "github.com/stanleyHayes/obiara/internal/privacy/application"
	jobsapplication "github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

// NewProcessorJob builds the scheduled privacy processor job.
func NewProcessorJob(processor privacyapplication.Processor, batchSize int, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "privacy.processor",
		Interval:    interval,
		Timeout:     2 * time.Minute,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			return processor.RunBatch(ctx, batchSize)
		},
	}
}
