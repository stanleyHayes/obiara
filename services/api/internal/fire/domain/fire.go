// Package domain models fires: scheduled voice gatherings with bounded
// capacity (FR-401: entry Tier 1+, capacity/waitlist/eject host-controllable
// in real time; agent_plan.md §5 P0 weekly fires v1). Capacity and waitlist
// transitions are server-authoritative and race-safe.
package domain

import (
	"errors"
	"strings"
	"time"
)

// FireStatus is the fire lifecycle.
type FireStatus string

const (
	StatusScheduled FireStatus = "scheduled"
	StatusLive      FireStatus = "live"
	StatusEnded     FireStatus = "ended"
	StatusCancelled FireStatus = "cancelled"
)

// CapacityBounds keep P0 fires inside the tested envelope (E09 exit:
// 150-seat acceptance target before scale rollout).
const (
	MinCapacity = 1
	MaxCapacity = 500
)

var (
	ErrFireIDRequired  = errors.New("fire id is required")
	ErrHostRequired    = errors.New("fire host is required")
	ErrTitleRequired   = errors.New("fire title is required")
	ErrInvalidCapacity = errors.New("fire capacity must be 1-500")
	ErrInvalidStart    = errors.New("fire start time is required")
	ErrFireNotOpen     = errors.New("fire is not open for RSVP")
	ErrTierTooLow      = errors.New("fire entry requires verification tier 1")
)

// Fire is one scheduled gathering.
type Fire struct {
	id         string
	hostID     string
	circleID   string
	title      string
	startsAt   time.Time
	capacity   int
	goingCount int
	status     FireStatus
	version    int64
	createdAt  time.Time
}

func NewFire(id, hostID, circleID, title string, startsAt time.Time, capacity int, now time.Time) (Fire, error) {
	if strings.TrimSpace(id) == "" {
		return Fire{}, ErrFireIDRequired
	}
	if strings.TrimSpace(hostID) == "" {
		return Fire{}, ErrHostRequired
	}
	if strings.TrimSpace(title) == "" {
		return Fire{}, ErrTitleRequired
	}
	if capacity < MinCapacity || capacity > MaxCapacity {
		return Fire{}, ErrInvalidCapacity
	}
	if startsAt.IsZero() {
		return Fire{}, ErrInvalidStart
	}
	return Fire{
		id:        id,
		hostID:    hostID,
		circleID:  circleID,
		title:     title,
		startsAt:  startsAt.UTC(),
		capacity:  capacity,
		status:    StatusScheduled,
		version:   1,
		createdAt: now.UTC(),
	}, nil
}

// ReconstituteFire rebuilds a stored fire without policy checks.
func ReconstituteFire(id, hostID, circleID, title string, startsAt time.Time, capacity, goingCount int, status FireStatus, version int64, createdAt time.Time) Fire {
	return Fire{
		id:         id,
		hostID:     hostID,
		circleID:   circleID,
		title:      title,
		startsAt:   startsAt.UTC(),
		capacity:   capacity,
		goingCount: goingCount,
		status:     status,
		version:    version,
		createdAt:  createdAt,
	}
}

// Admit decides an RSVP against current capacity and returns the attendance
// record. waitlistDepth is the current number of waitlisted members. The
// going counter moves with the decision and is persisted in the same
// transaction as the attendance record.
func (fire *Fire) Admit(memberID string, waitlistDepth int, now time.Time) (RSVP, error) {
	if fire.status != StatusScheduled && fire.status != StatusLive {
		return RSVP{}, ErrFireNotOpen
	}
	fire.version++
	if fire.goingCount < fire.capacity {
		fire.goingCount++
		return RSVP{fireID: fire.id, memberID: memberID, status: RSVPGoing, version: 1, createdAt: now.UTC()}, nil
	}
	return RSVP{fireID: fire.id, memberID: memberID, status: RSVPWaitlisted, position: waitlistDepth + 1, version: 1, createdAt: now.UTC()}, nil
}

// Release frees a going seat. The waitlist promotion happens in the same
// transaction via Promote.
func (fire *Fire) Release() {
	if fire.goingCount > 0 {
		fire.goingCount--
	}
	fire.version++
}

// Touch records an attendance change that does not move the counter
// (waitlisted cancellation) so version pinning stays uniform.
func (fire *Fire) Touch() {
	fire.version++
}

// Promote moves a waitlisted RSVP into the freed seat.
func (fire *Fire) Promote(rsvp *RSVP, now time.Time) {
	fire.goingCount++
	rsvp.status = RSVPGoing
	rsvp.position = 0
	rsvp.version++
}

func (fire Fire) ID() string           { return fire.id }
func (fire Fire) HostID() string       { return fire.hostID }
func (fire Fire) CircleID() string     { return fire.circleID }
func (fire Fire) Title() string        { return fire.title }
func (fire Fire) StartsAt() time.Time  { return fire.startsAt }
func (fire Fire) Capacity() int        { return fire.capacity }
func (fire Fire) GoingCount() int      { return fire.goingCount }
func (fire Fire) Status() FireStatus   { return fire.status }
func (fire Fire) Version() int64       { return fire.version }
func (fire Fire) CreatedAt() time.Time { return fire.createdAt }

// RSVPStatus is the attendance state.
type RSVPStatus string

const (
	RSVPGoing      RSVPStatus = "going"
	RSVPWaitlisted RSVPStatus = "waitlisted"
)

// RSVP is one member's attendance record for a fire.
type RSVP struct {
	fireID    string
	memberID  string
	status    RSVPStatus
	position  int
	version   int64
	createdAt time.Time
}

// ReconstituteRSVP rebuilds a stored RSVP without policy checks.
func ReconstituteRSVP(fireID, memberID string, status RSVPStatus, position int, version int64, createdAt time.Time) RSVP {
	return RSVP{fireID: fireID, memberID: memberID, status: status, position: position, version: version, createdAt: createdAt}
}

func (rsvp RSVP) FireID() string       { return rsvp.fireID }
func (rsvp RSVP) MemberID() string     { return rsvp.memberID }
func (rsvp RSVP) Status() RSVPStatus   { return rsvp.status }
func (rsvp RSVP) Position() int        { return rsvp.position }
func (rsvp RSVP) Version() int64       { return rsvp.version }
func (rsvp RSVP) CreatedAt() time.Time { return rsvp.createdAt }
