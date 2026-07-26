// Package application ingests playback telemetry and answers sow
// eligibility. The listening boundary is server-authoritative: clients
// send ranges, they never assert eligibility (FR-202; agent_plan.md §4:
// clients render state, they never decide it).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
)

// HeartbeatRange is one client-reported playback range in seconds.
type HeartbeatRange struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// ListeningRepository persists playback records with optimistic concurrency.
type ListeningRepository interface {
	Find(context.Context, string, string) (domain.Playback, error)
	// Save creates or updates the record, pinned to its version.
	// Concurrent writers get domain.ErrStalePlayback.
	Save(context.Context, domain.Playback) error
}

// ListeningService records heartbeats and evaluates eligibility.
type ListeningService struct {
	playback ListeningRepository
	now      func() time.Time
}

func NewListeningService(playback ListeningRepository, now func() time.Time) ListeningService {
	return ListeningService{playback: playback, now: now}
}

const maxStaleRetries = 3

// RecordHeartbeats merges a batch of ranges for one listener/asset and
// returns the updated record. Stale-write races (multi-device listening)
// retry against the freshly loaded record.
func (service ListeningService) RecordHeartbeats(ctx context.Context, listenerID, assetID string, assetDuration float64, ranges []HeartbeatRange) (domain.Playback, error) {
	for attempt := 0; ; attempt++ {
		record, err := service.playback.Find(ctx, listenerID, assetID)
		if errors.Is(err, domain.ErrPlaybackNotFound) {
			record, err = domain.NewPlayback(listenerID, assetID, assetDuration)
			if err != nil {
				return domain.Playback{}, err
			}
		} else if err != nil {
			return domain.Playback{}, err
		}

		for _, heartbeat := range ranges {
			if err := record.Record(heartbeat.Start, heartbeat.End); err != nil {
				return domain.Playback{}, err
			}
		}
		if len(ranges) == 0 {
			return record, nil
		}
		record.Commit(service.now())

		err = service.playback.Save(ctx, record)
		if errors.Is(err, domain.ErrStalePlayback) && attempt < maxStaleRetries {
			continue
		}
		if err != nil {
			return domain.Playback{}, err
		}
		return record, nil
	}
}

// Eligibility reports the sow-arming state for the sow boundary (E06-S04).
// It is read-only and never exposed to the sower's counterpart (FR-205).
func (service ListeningService) Eligibility(ctx context.Context, listenerID, assetID string) (eligible bool, totalSeconds float64, err error) {
	record, err := service.playback.Find(ctx, listenerID, assetID)
	if errors.Is(err, domain.ErrPlaybackNotFound) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return record.Eligible(), record.TotalSeconds(), nil
}
