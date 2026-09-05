package domain

import (
	"errors"
	"fmt"
	"time"
)

// Action is a T&S ladder action (Doc 09 §2 action ladder).
type Action string

const (
	ActionWarning    Action = "warning"
	ActionSuspend14d Action = "suspend_14d"
	ActionSuspend30d Action = "suspend_30d"
	ActionSuspend90d Action = "suspend_90d"
	ActionBan        Action = "ban"
)

var ErrActionNotOnLadder = errors.New("action is not on the ladder for this tier")

// SuspensionDuration maps suspension actions to their length (Doc 09 §2:
// suspensions run 14-90 days).
func SuspensionDuration(action Action) (time.Duration, bool) {
	switch action {
	case ActionSuspend14d:
		return 14 * 24 * time.Hour, true
	case ActionSuspend30d:
		return 30 * 24 * time.Hour, true
	case ActionSuspend90d:
		return 90 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

// IsSuspension reports whether the action is a timed suspension.
func IsSuspension(action Action) bool {
	_, ok := SuspensionDuration(action)
	return ok
}

// CheckLadder validates an action against the case tier and the subject's
// prior action count (Doc 09 §2):
//   - Tier A: immediate ban only; Tier A never rehabilitates on-platform.
//   - Tier B: first action is a 14-90 day suspension; repeat is a ban.
//   - Tier C: first action is a warning; repeat escalates to a suspension.
func CheckLadder(tier Tier, action Action, priorActions int) error {
	switch tier {
	case TierA:
		if action != ActionBan {
			return fmt.Errorf("%w: tier A allows immediate ban only", ErrActionNotOnLadder)
		}
		return nil
	case TierB:
		if priorActions == 0 {
			if !IsSuspension(action) {
				return fmt.Errorf("%w: first tier-B action must be a suspension", ErrActionNotOnLadder)
			}
			return nil
		}
		if action != ActionBan {
			return fmt.Errorf("%w: repeat tier-B action must be a ban", ErrActionNotOnLadder)
		}
		return nil
	case TierC:
		if priorActions == 0 {
			if action != ActionWarning {
				return fmt.Errorf("%w: first tier-C action must be a warning", ErrActionNotOnLadder)
			}
			return nil
		}
		if !IsSuspension(action) {
			return fmt.Errorf("%w: repeat tier-C escalates to a suspension", ErrActionNotOnLadder)
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown tier", ErrActionNotOnLadder)
	}
}

// ActionRecord is the immutable audit entry for one applied action
// (agent_plan.md §4: all privileged and irreversible actions are
// auditable).
type ActionRecord struct {
	ID string
	// CommandID is the operator's request. It is on the record because the
	// log is what CountForSubject reads, and a double-submitted action would
	// otherwise count twice and escalate the subject's next one — a member
	// warned once would be treated as a repeat offender.
	CommandID string
	CaseID    string
	SubjectID string
	Action    Action
	ActorID   string
	Priors    int
	CreatedAt time.Time
}
