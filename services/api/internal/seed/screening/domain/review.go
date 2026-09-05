// Package domain models a sow screening review: the lifecycle of one
// decision a person has to make before a sow is delivered.
//
// It carries the decision and not the content. A reviewer needs to read the
// sow's words, and those live in the store beside this record — keeping them
// out of the aggregate means the rules about deciding can be tested without
// a member's writing anywhere near them, and means a decision can be replayed
// from the log after the content has been deleted.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalid    = errors.New("invalid screening review")
	ErrNotPending = errors.New("screening review is already decided")
)

// reference is the shape the screening adapter validates a routed review's
// reference against. A review whose id does not match could be routed and
// then never found again.
var reference = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Status string

const (
	StatusPending  Status = "pending"
	StatusReleased Status = "released"
	StatusRefused  Status = "refused"
)

type Review struct {
	id        string
	reason    string
	status    Status
	routedAt  time.Time
	decidedAt *time.Time
	decidedBy string
	commandID string
}

// Route opens a review for a sow that screening sent to a person.
func Route(id, reason string, at time.Time) (Review, error) {
	id = strings.TrimSpace(id)
	if !reference.MatchString(id) || strings.TrimSpace(reason) == "" || at.IsZero() {
		return Review{}, ErrInvalid
	}
	return Review{id: id, reason: strings.TrimSpace(reason), status: StatusPending, routedAt: at.UTC()}, nil
}

// Decide records a reviewer's judgement.
//
// The actor is required because a decision nobody is accountable for is not a
// review, and the command id is required because deciding twice would release
// a sow that was refused, or refund its seed a second time.
func (review Review) Decide(status Status, actorID, commandID string, at time.Time) (Review, error) {
	if review.status != StatusPending {
		return Review{}, ErrNotPending
	}
	if status != StatusReleased && status != StatusRefused {
		return Review{}, ErrInvalid
	}
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(commandID) == "" || at.IsZero() {
		return Review{}, ErrInvalid
	}
	decided := at.UTC()
	next := review
	next.status = status
	next.decidedAt = &decided
	next.decidedBy = strings.TrimSpace(actorID)
	next.commandID = strings.TrimSpace(commandID)
	return next, nil
}

// Reconstitute rebuilds a stored review without re-running its invariants,
// which is correct only because they were enforced on the way in.
func Reconstitute(id, reason string, status Status, routedAt time.Time, decidedAt *time.Time, decidedBy, commandID string) Review {
	return Review{
		id: id, reason: reason, status: status, routedAt: routedAt,
		decidedAt: decidedAt, decidedBy: decidedBy, commandID: commandID,
	}
}

func (review Review) ID() string            { return review.id }
func (review Review) Reason() string        { return review.reason }
func (review Review) Status() Status        { return review.status }
func (review Review) RoutedAt() time.Time   { return review.routedAt }
func (review Review) DecidedAt() *time.Time { return review.decidedAt }
func (review Review) DecidedBy() string     { return review.decidedBy }
func (review Review) CommandID() string     { return review.commandID }

// Released reports whether the sow behind this review may be delivered.
func (review Review) Released() bool { return review.status == StatusReleased }
