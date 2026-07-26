// Package domain models server-verified listening eligibility (FR-202,
// E06-S03): sowing requires at least RequiredSeconds of unique cumulative
// playback of the target's Voice of Introduction, derived only from
// server-merged playback telemetry. Partial-listen events are never
// exposed to sowers (FR-205) — this aggregate stays inside the seed
// context and is read by the sow boundary only.
package domain

import (
	"errors"
	"sort"
	"time"
)

// RequiredSeconds is the sow-arming threshold (FR-202).
const RequiredSeconds = 20.0

var (
	ErrListenerRequired = errors.New("listener id is required")
	ErrAssetRequired    = errors.New("voice asset id is required")
	ErrInvalidDuration  = errors.New("asset duration must be positive")
	ErrInvalidRange     = errors.New("playback range must satisfy 0 <= start < end")
	ErrPlaybackNotFound = errors.New("no playback record for this listener and asset")
	ErrStalePlayback    = errors.New("playback record changed concurrently")
)

// Interval is a half-open [start, end) playback range in seconds.
type Interval struct {
	Start float64 `bson:"start"`
	End   float64 `bson:"end"`
}

// Playback is the merged listening record of one listener for one voice
// asset. Intervals are stored merged, sorted and disjoint, so duplicate,
// out-of-order and replayed heartbeats never double-count (TP-E06-S03-01).
type Playback struct {
	listenerID string
	assetID    string
	duration   float64
	intervals  []Interval
	version    int64
	updatedAt  time.Time
}

// NewPlayback starts a record for a listener/asset pair. duration clamps
// telemetry to the real asset length once persistence-backed asset
// metadata lands (media adapters); until then it is caller-asserted and
// only ever narrows ranges.
func NewPlayback(listenerID, assetID string, duration float64) (Playback, error) {
	if listenerID == "" {
		return Playback{}, ErrListenerRequired
	}
	if assetID == "" {
		return Playback{}, ErrAssetRequired
	}
	if duration <= 0 {
		return Playback{}, ErrInvalidDuration
	}
	return Playback{listenerID: listenerID, assetID: assetID, duration: duration}, nil
}

// ReconstitutePlayback rebuilds a stored record without policy checks.
func ReconstitutePlayback(listenerID, assetID string, duration float64, intervals []Interval, version int64, updatedAt time.Time) Playback {
	return Playback{
		listenerID: listenerID,
		assetID:    assetID,
		duration:   duration,
		intervals:  intervals,
		version:    version,
		updatedAt:  updatedAt,
	}
}

// Record merges one heartbeat range without versioning; callers commit a
// batch with Commit. Ranges are clamped to the asset duration; the union
// semantics make the operation idempotent.
func (playback *Playback) Record(start, end float64) error {
	if start < 0 || end <= start {
		return ErrInvalidRange
	}
	if end > playback.duration {
		end = playback.duration
	}
	if end <= start {
		return ErrInvalidRange
	}
	playback.intervals = merge(append(playback.intervals, Interval{Start: start, End: end}))
	return nil
}

// Commit seals a batch of Record calls: one version increment and a fresh
// timestamp, keeping optimistic concurrency batch-atomic.
func (playback *Playback) Commit(now time.Time) {
	playback.version++
	playback.updatedAt = now.UTC()
}

// TotalSeconds is the unique cumulative listening time.
func (playback Playback) TotalSeconds() float64 {
	total := 0.0
	for _, interval := range playback.intervals {
		total += interval.End - interval.Start
	}
	return total
}

// Eligible reports whether the listener may sow (FR-202).
func (playback Playback) Eligible() bool {
	return playback.TotalSeconds() >= RequiredSeconds
}

func (playback Playback) ListenerID() string    { return playback.listenerID }
func (playback Playback) AssetID() string       { return playback.assetID }
func (playback Playback) Duration() float64     { return playback.duration }
func (playback Playback) Intervals() []Interval { return playback.intervals }
func (playback Playback) Version() int64        { return playback.version }
func (playback Playback) UpdatedAt() time.Time  { return playback.updatedAt }

// merge unions overlapping or adjacent intervals into a sorted disjoint set.
func merge(intervals []Interval) []Interval {
	if len(intervals) == 0 {
		return nil
	}
	sorted := make([]Interval, len(intervals))
	copy(sorted, intervals)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start == sorted[j].Start {
			return sorted[i].End < sorted[j].End
		}
		return sorted[i].Start < sorted[j].Start
	})

	merged := []Interval{sorted[0]}
	for _, current := range sorted[1:] {
		last := &merged[len(merged)-1]
		if current.Start <= last.End {
			if current.End > last.End {
				last.End = current.End
			}
			continue
		}
		merged = append(merged, current)
	}
	return merged
}
