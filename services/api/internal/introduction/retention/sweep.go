// Package retention erases withdrawn and expired Voice of Introduction audio.
//
// It runs inside the API rather than the worker on purpose. Erasure is only
// provable because the aggregate records it — `MarkPurged` appends an audit
// event, and a sweep that wrote to Mongo directly would erase the bytes and
// lose the proof. The aggregate lives here, so the sweep does too.
package retention

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

// batchSize bounds one pass. Erasing an entire backlog in a single sweep
// would hold a long run of calls against storage and, on a first run, look
// indistinguishable from an incident.
const batchSize = 100

// Store finds recordings whose audio should no longer exist.
type Store interface {
	DueForPurge(ctx context.Context, at time.Time, limit int) ([]domain.Introduction, error)
}

// Purger marks the aggregate purged once the bytes are gone.
type Purger interface {
	Purge(ctx context.Context, introductionID, commandID string) (domain.Introduction, error)
}

// Eraser removes the stored audio.
type Eraser interface {
	Erase(ctx context.Context, assetID string) error
}

type Sweeper struct {
	store  Store
	eraser Eraser
	purger Purger
	now    func() time.Time
	log    *slog.Logger
}

func NewSweeper(store Store, eraser Eraser, purger Purger, now func() time.Time, log *slog.Logger) Sweeper {
	if now == nil {
		now = time.Now
	}
	return Sweeper{store: store, eraser: eraser, purger: purger, now: now, log: log}
}

// Once erases one batch and reports how many recordings it finished.
//
// Bytes first, then the aggregate. Marking one purged before the audio is
// gone would record an erasure that did not happen, and nothing afterwards
// would go looking for those bytes again.
func (sweeper Sweeper) Once(ctx context.Context) (int, error) {
	due, err := sweeper.store.DueForPurge(ctx, sweeper.now().UTC(), batchSize)
	if err != nil {
		return 0, err
	}

	purged := 0
	for _, introduction := range due {
		if err := sweeper.eraser.Erase(ctx, introduction.Media().AssetID()); err != nil {
			// One recording that will not erase must not stop the rest. It
			// keeps purge_pending and the next pass tries again — which is
			// also how a legal hold is honoured without special-casing it.
			sweeper.log.WarnContext(ctx, "voice erasure deferred",
				slog.String("introduction", introduction.ID()),
				slog.String("reason", err.Error()))
			continue
		}
		// Derived from the aggregate, so a retried sweep replays the same
		// command rather than opening a second purge.
		commandID := "purge:" + introduction.ID()
		if _, err := sweeper.purger.Purge(ctx, introduction.ID(), commandID); err != nil {
			if errors.Is(err, application.ErrCommandAlreadyUsed) {
				purged++
				continue
			}
			sweeper.log.WarnContext(ctx, "voice purge could not be recorded",
				slog.String("introduction", introduction.ID()),
				slog.String("reason", err.Error()))
			continue
		}
		purged++
	}
	return purged, nil
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
			purged, err := sweeper.Once(ctx)
			if err != nil {
				sweeper.log.WarnContext(ctx, "voice erasure sweep failed",
					slog.String("reason", err.Error()))
				continue
			}
			if purged > 0 {
				sweeper.log.InfoContext(ctx, "voice recordings erased",
					slog.Int("count", purged))
			}
		}
	}
}
