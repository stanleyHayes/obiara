// Package relay delivers durable outbox records to publishers (E02-S09
// consuming E02-S06). Publishing is at-least-once; consumers dedupe via the
// inbox store. Failed deliveries increment the attempt counter for backoff
// and dead-letter triage.
package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

// Publisher is the outbound port for event publication (notifications,
// projections, provider calls). Implementations must be provider-neutral.
type Publisher interface {
	Publish(context.Context, outbox.Record) error
}

// Store is the outbox surface the relay needs; *outbox.Store satisfies it.
type Store interface {
	Pending(context.Context, int) ([]outbox.Record, error)
	MarkPublished(context.Context, string) error
	MarkAttemptFailed(context.Context, string) error
}

// NewOutboxJob builds the scheduled outbox relay job.
func NewOutboxJob(store Store, publisher Publisher, batchSize int, interval time.Duration) application.Job {
	return application.Job{
		Name:        "outbox.relay",
		Interval:    interval,
		Timeout:     30 * time.Second,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			pending, err := store.Pending(ctx, batchSize)
			if err != nil {
				return fmt.Errorf("read pending outbox records: %w", err)
			}
			for _, record := range pending {
				if err := ctx.Err(); err != nil {
					return err
				}
				if err := publisher.Publish(ctx, record); err != nil {
					_ = store.MarkAttemptFailed(ctx, record.ID)
					return fmt.Errorf("publish outbox record %q: %w", record.ID, err)
				}
				if err := store.MarkPublished(ctx, record.ID); err != nil {
					return fmt.Errorf("mark outbox record %q published: %w", record.ID, err)
				}
			}
			return nil
		},
	}
}
