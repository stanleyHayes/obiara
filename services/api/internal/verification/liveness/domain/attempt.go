// Package domain models a privacy-minimal voice-and-face liveness attempt.
// Biometric media and free-form provider payloads never enter this aggregate.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidAttempt   = errors.New("invalid liveness attempt")
	ErrAttemptNotOpen   = errors.New("liveness attempt is not open")
	ErrManualReviewOnly = errors.New("liveness attempt is not awaiting manual review")
	ErrStaleVersion     = errors.New("stale liveness attempt version")
)

type Status string

const (
	StatusPending      Status = "pending"
	StatusPassed       Status = "passed"
	StatusFailed       Status = "failed"
	StatusQueuedManual Status = "queued_manual"
)

type Reason string

const (
	ReasonProviderLive        Reason = "provider_live"
	ReasonProviderNotLive     Reason = "provider_not_live"
	ReasonProviderUncertain   Reason = "provider_uncertain"
	ReasonProviderUnavailable Reason = "provider_unavailable"
	ReasonManualPass          Reason = "manual_pass"
	ReasonManualFail          Reason = "manual_fail"
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)

type Event struct {
	status     Status
	reason     Reason
	actorKey   string
	occurredAt time.Time
	version    uint64
}

type EventParams struct {
	Status     Status
	Reason     Reason
	ActorKey   string
	OccurredAt time.Time
	Version    uint64
}

func NewEvent(params EventParams) (Event, error) {
	if !validStatusReason(params.Status, params.Reason) ||
		!validDigest(strings.TrimSpace(params.ActorKey)) ||
		params.OccurredAt.IsZero() || params.Version < 2 {
		return Event{}, ErrInvalidAttempt
	}
	return Event{
		status: params.Status, reason: params.Reason,
		actorKey: params.ActorKey, occurredAt: params.OccurredAt.UTC(),
		version: params.Version,
	}, nil
}

func (event Event) Status() Status        { return event.status }
func (event Event) Reason() Reason        { return event.reason }
func (event Event) ActorKey() string      { return event.actorKey }
func (event Event) OccurredAt() time.Time { return event.occurredAt }
func (event Event) Version() uint64       { return event.version }

// Attempt retains only keyed subject/input proof and a bounded transition
// history. Voice/face artifact references live in the temporary review queue.
type Attempt struct {
	id          string
	commandID   string
	subjectKey  string
	inputKey    string
	status      Status
	reason      Reason
	providerRef string
	createdAt   time.Time
	decidedAt   time.Time
	version     uint64
	events      []Event
}

func NewAttempt(id, commandID, subjectKey, inputKey string, now time.Time) (Attempt, error) {
	id = strings.TrimSpace(id)
	commandID = strings.TrimSpace(commandID)
	subjectKey = strings.TrimSpace(subjectKey)
	inputKey = strings.TrimSpace(inputKey)
	if !validOpaque(id) || !validOpaque(commandID) || !validDigest(subjectKey) ||
		!validDigest(inputKey) || now.IsZero() {
		return Attempt{}, ErrInvalidAttempt
	}
	return Attempt{
		id: id, commandID: commandID, subjectKey: subjectKey, inputKey: inputKey,
		status: StatusPending, createdAt: now.UTC(), version: 1,
	}, nil
}

func Reconstitute(
	id, commandID, subjectKey, inputKey string,
	status Status, reason Reason, providerRef string,
	createdAt, decidedAt time.Time, version uint64, events []Event,
) (Attempt, error) {
	attempt, err := NewAttempt(id, commandID, subjectKey, inputKey, createdAt)
	if err != nil {
		return Attempt{}, err
	}
	if status == StatusPending {
		if reason != "" || providerRef != "" || !decidedAt.IsZero() || version != 1 || len(events) != 0 {
			return Attempt{}, ErrInvalidAttempt
		}
		return attempt, nil
	}
	if version < 2 || len(events) == 0 || events[len(events)-1].Version() != version ||
		events[len(events)-1].Status() != status || events[len(events)-1].Reason() != reason ||
		decidedAt.IsZero() || !validStatusReason(status, reason) {
		return Attempt{}, ErrInvalidAttempt
	}
	for index, event := range events {
		if event.Version() != uint64(index+2) ||
			(index > 0 && event.OccurredAt().Before(events[index-1].OccurredAt())) {
			return Attempt{}, ErrInvalidAttempt
		}
	}
	providerDecision := reason == ReasonProviderLive || reason == ReasonProviderNotLive
	if (providerDecision && !validOpaque(providerRef)) || (!providerDecision && providerRef != "") {
		return Attempt{}, ErrInvalidAttempt
	}
	attempt.status = status
	attempt.reason = reason
	attempt.providerRef = providerRef
	attempt.decidedAt = decidedAt.UTC()
	attempt.version = version
	attempt.events = append([]Event(nil), events...)
	return attempt, nil
}

func (attempt Attempt) QueueManual(reason Reason, actorKey string, now time.Time, expectedVersion uint64) (Attempt, error) {
	if reason != ReasonProviderUncertain && reason != ReasonProviderUnavailable {
		return Attempt{}, ErrInvalidAttempt
	}
	return attempt.transition(StatusQueuedManual, reason, actorKey, "", now, expectedVersion, false)
}

func (attempt Attempt) ProviderDecision(passed bool, providerRef, actorKey string, now time.Time, expectedVersion uint64) (Attempt, error) {
	status, reason := StatusFailed, ReasonProviderNotLive
	if passed {
		status, reason = StatusPassed, ReasonProviderLive
	}
	return attempt.transition(status, reason, actorKey, providerRef, now, expectedVersion, false)
}

func (attempt Attempt) ManualDecision(passed bool, actorKey string, now time.Time, expectedVersion uint64) (Attempt, error) {
	if attempt.status != StatusQueuedManual {
		return Attempt{}, ErrManualReviewOnly
	}
	status, reason := StatusFailed, ReasonManualFail
	if passed {
		status, reason = StatusPassed, ReasonManualPass
	}
	return attempt.transition(status, reason, actorKey, "", now, expectedVersion, true)
}

func (attempt Attempt) transition(status Status, reason Reason, actorKey, providerRef string, now time.Time, expectedVersion uint64, fromManual bool) (Attempt, error) {
	actorKey = strings.TrimSpace(actorKey)
	providerRef = strings.TrimSpace(providerRef)
	if expectedVersion != attempt.version {
		return Attempt{}, ErrStaleVersion
	}
	if now.IsZero() || !validDigest(actorKey) {
		return Attempt{}, ErrInvalidAttempt
	}
	if fromManual {
		if attempt.status != StatusQueuedManual {
			return Attempt{}, ErrManualReviewOnly
		}
	} else if attempt.status != StatusPending {
		return Attempt{}, ErrAttemptNotOpen
	}
	if (reason == ReasonProviderLive || reason == ReasonProviderNotLive) && !validOpaque(providerRef) {
		return Attempt{}, ErrInvalidAttempt
	}
	decidedAt := now.UTC()
	next := attempt
	next.status = status
	next.reason = reason
	next.providerRef = providerRef
	next.decidedAt = decidedAt
	next.version++
	next.events = append(attempt.Events(), Event{
		status: status, reason: reason, actorKey: actorKey,
		occurredAt: decidedAt, version: next.version,
	})
	return next, nil
}

func (attempt Attempt) ID() string           { return attempt.id }
func (attempt Attempt) CommandID() string    { return attempt.commandID }
func (attempt Attempt) SubjectKey() string   { return attempt.subjectKey }
func (attempt Attempt) InputKey() string     { return attempt.inputKey }
func (attempt Attempt) Status() Status       { return attempt.status }
func (attempt Attempt) Reason() Reason       { return attempt.reason }
func (attempt Attempt) ProviderRef() string  { return attempt.providerRef }
func (attempt Attempt) CreatedAt() time.Time { return attempt.createdAt }
func (attempt Attempt) DecidedAt() time.Time { return attempt.decidedAt }
func (attempt Attempt) Version() uint64      { return attempt.version }
func (attempt Attempt) Events() []Event      { return append([]Event(nil), attempt.events...) }
func (attempt Attempt) Passed() bool         { return attempt.status == StatusPassed }
func (attempt Attempt) Terminal() bool {
	return attempt.status == StatusPassed || attempt.status == StatusFailed
}

func validOpaque(value string) bool {
	return opaquePattern.MatchString(value)
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func validStatusReason(status Status, reason Reason) bool {
	switch status {
	case StatusPassed:
		return reason == ReasonProviderLive || reason == ReasonManualPass
	case StatusFailed:
		return reason == ReasonProviderNotLive || reason == ReasonManualFail
	case StatusQueuedManual:
		return reason == ReasonProviderUncertain || reason == ReasonProviderUnavailable
	default:
		return false
	}
}
