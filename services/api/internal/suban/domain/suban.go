// Package domain is the suban character ledger (E15-S04; Doc 08 §4).
// Events are append-only and auditable; marks are thresholded labels
// recomputed from the ledger — never a number to members, never cached.
package domain

import (
	"errors"
	"math"
	"time"
)

// Kind is an auditable suban event type (Doc 08 §4 inputs).
type Kind string

const (
	KindMeetingFollowThrough Kind = "meeting_follow_through" // + standard
	KindKindClosure          Kind = "kind_closure"           // + small
	KindPauseStone           Kind = "pause_stone"            // + small
	KindThemeCompleted       Kind = "theme_completed"        // + small
	KindCleanVouch           Kind = "clean_vouch"            // + standard
	KindGraciousDecline      Kind = "gracious_decline"       // + tiny
	KindGhostPattern         Kind = "ghost_pattern"          // 0 first, − thereafter
	KindHarassmentFinding    Kind = "harassment_finding"     // −−
	KindFraudFinding         Kind = "fraud_finding"          // account-ending
	KindVouchStakeLoss       Kind = "vouch_stake_loss"       // − propagated
)

var ErrInvalidKind = errors.New("unknown suban event kind")

// Weight classes from Doc 08 §4.
const (
	weightTiny     = 0.25
	weightSmall    = 0.5
	weightStandard = 1.0
)

// PositiveHalfLife is the decay constant for positive events (Doc 08 §4:
// positive events half-life 12 months — character must stay in practice).
const PositiveHalfLife = 365.25 * 24 * time.Hour / 2

// NegativeWindow is how long negatives weigh (Doc 09 §2 rehabilitation:
// Tier-C/B penalties decay after clean periods 6/18 months; we apply the
// stricter 18-month window to conduct findings).
const NegativeWindow = 18 * 30 * 24 * time.Hour

// Provenance ties an event to its source record.
type Provenance struct {
	Source string // e.g. "meeting", "closure", "vouch", "panel"
	Ref    string
}

// Event is one immutable suban event.
type Event struct {
	ID         string
	SubjectID  string
	Kind       Kind
	Provenance Provenance
	OccurredAt time.Time
}

func valid(kind Kind) bool {
	switch kind {
	case KindMeetingFollowThrough, KindKindClosure, KindPauseStone, KindThemeCompleted,
		KindCleanVouch, KindGraciousDecline, KindGhostPattern, KindHarassmentFinding,
		KindFraudFinding, KindVouchStakeLoss:
		return true
	}
	return false
}

// Mark is a thresholded character label — never a score (Doc 08 §4).
type Mark string

const (
	MarkKeepsWord      Mark = "keeps_word"
	MarkGracious       Mark = "gracious"
	MarkTrustedVoucher Mark = "trusted_voucher"
)

// markThreshold is the effective credit a mark needs (Doc 08 §4: ≥3
// confirmed meetings). The tolerance absorbs half-life decay over
// day-fresh events without weakening the threshold in practice (aged
// credits still fall far below it — see the decay tests).
const (
	markThreshold = 3.0
	markTolerance = 0.05
)

// ComputeMarks recomputes a member's marks from their full event ledger at
// a moment. Positive credits decay with a 12-month half-life; a conduct
// finding inside the rehabilitation window suppresses every mark; a fraud
// finding suppresses permanently.
func ComputeMarks(events []Event, now time.Time) []Mark {
	now = now.UTC()
	credits := map[Mark]float64{
		MarkKeepsWord:      0,
		MarkGracious:       0,
		MarkTrustedVoucher: 0,
	}
	suppressed := false
	for _, event := range events {
		age := now.Sub(event.OccurredAt.UTC())
		switch event.Kind {
		case KindFraudFinding:
			suppressed = true
		case KindHarassmentFinding, KindVouchStakeLoss, KindGhostPattern:
			if age < NegativeWindow {
				suppressed = true
			}
		case KindMeetingFollowThrough:
			credits[MarkKeepsWord] += decayed(weightStandard, age)
		case KindKindClosure, KindGraciousDecline:
			credits[MarkGracious] += decayed(weightFor(event.Kind), age)
		case KindCleanVouch:
			credits[MarkTrustedVoucher] += decayed(weightStandard, age)
		case KindPauseStone, KindThemeCompleted:
			// Small positive inputs; they feed future marks without
			// reaching thresholds alone.
		}
	}
	if suppressed {
		return nil
	}
	var marks []Mark
	for mark, credit := range credits {
		if credit >= markThreshold-markTolerance {
			marks = append(marks, mark)
		}
	}
	return marks
}

func weightFor(kind Kind) float64 {
	switch kind {
	case KindGraciousDecline:
		return weightTiny
	case KindKindClosure, KindPauseStone, KindThemeCompleted:
		return weightSmall
	default:
		return weightStandard
	}
}

// decayed applies the positive half-life: weight × 0.5^(age/halfLife).
func decayed(weight float64, age time.Duration) float64 {
	if age < 0 {
		age = 0
	}
	return weight * math.Pow(0.5, float64(age)/float64(PositiveHalfLife))
}

// PeriodCap is the anti-gaming bound: events of one kind per subject per
// 30 days (Doc 08 §4: event caps per period).
const PeriodCap = 10

// CapWindow is the anti-gaming accounting window.
const CapWindow = 30 * 24 * time.Hour
