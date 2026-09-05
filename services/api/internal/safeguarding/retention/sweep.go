// Package retention drives the safeguarding purge to completion.
//
// Assess persists the block and attempts one purge inline, then returns. If
// that single attempt fails — a transaction lost, Mongo briefly unreachable —
// the restriction stays pending and, without this, nothing would ever try
// again. The 24-hour SLA in the domain is a promise about data being gone, and
// a promise nothing retries is not one.
//
// It runs inside the API rather than the worker for the same reason erasure
// does: CompletePurge writes the proof through the aggregate, and a sweep
// deleting rows directly would destroy the data and lose the record that it
// had been destroyed.
package retention

import (
	"context"
	"log/slog"
	"time"
)

// batchSize bounds one pass, so a backlog is worked through steadily rather
// than as one long run of deletes that looks like an incident.
const batchSize = 100

// Purger completes the purges that are already due.
type Purger interface {
	PurgePending(ctx context.Context, dueBefore time.Time, limit int) (int, error)
}

type Sweeper struct {
	purger Purger
	now    func() time.Time
	log    *slog.Logger
}

func NewSweeper(purger Purger, now func() time.Time, log *slog.Logger) Sweeper {
	if now == nil {
		now = time.Now
	}
	if log == nil {
		log = slog.Default()
	}
	return Sweeper{purger: purger, now: now, log: log}
}

// Once completes one batch of purges due before the end of the next pass.
//
// The horizon reaches one interval ahead rather than stopping at now, so a job
// is retried before its deadline instead of exactly on it. A purge that is
// only attempted at the moment it expires has no attempts left.
func (sweeper Sweeper) Once(ctx context.Context, horizon time.Duration) (int, error) {
	return sweeper.purger.PurgePending(ctx, sweeper.now().UTC().Add(horizon), batchSize)
}

// Run sweeps on an interval until the context ends.
func (sweeper Sweeper) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			completed, err := sweeper.Once(ctx, interval)
			if err != nil {
				// Deliberately not fatal. A pending purge is a compliance
				// deadline, not a broken service, and the next pass retries
				// it — but it must be visible while it is outstanding.
				sweeper.log.WarnContext(ctx, "safeguarding purge sweep incomplete",
					slog.String("reason", err.Error()))
			}
			if completed > 0 {
				sweeper.log.InfoContext(ctx, "safeguarding purges completed",
					slog.Int("count", completed))
			}
		}
	}
}
