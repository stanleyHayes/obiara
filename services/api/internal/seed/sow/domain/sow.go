package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalid           = errors.New("invalid sow")
	ErrNotConfirmed      = errors.New("deliberate confirmation is required")
	ErrScreeningRejected = errors.New("sow failed media screening")
	ErrCommandMismatch   = errors.New("command id reused with different input")
)

// Status is where a sow stands. It exists because sows are screened by a
// human before delivery: a sow awaiting that decision is neither delivered
// nor refused, and before this the aggregate had no way to say so — the
// review path came back as an error and the member was told the service was
// unavailable.
type Status string

const (
	// StatusPendingReview is a sow whose seed has been spent and whose
	// delivery is waiting on a person. The seed is spent on the way in
	// because a sow nobody could send for free is the point of the
	// allowance; it is refunded if the review refuses it (M4-ABUSE-01).
	StatusPendingReview Status = "pending_review"
	StatusDelivered     Status = "delivered"
	StatusRejected      Status = "rejected"
)

var ErrNotPending = errors.New("sow is not awaiting review")

type Media struct {
	Key          string
	ScreeningKey string
}

type Sow struct {
	ID             string
	ActorKey       string
	Body           string
	Media          []Media
	CommandID      string
	Fingerprint    string
	AllowanceUnits int64
	Status         Status
	// ScreeningRef ties the sow to the screening decision or review that
	// produced its status, so a release or a refusal can be traced back to
	// the judgement that caused it.
	ScreeningRef string
	AcceptedAt   time.Time
	DecidedAt    *time.Time
}

func Accept(id, actorKey, body string, media []Media, commandID, fingerprint string, units int64, status Status, screeningRef string, at time.Time) (Sow, error) {
	if strings.TrimSpace(id) == "" || actorKey == "" || strings.TrimSpace(body) == "" || commandID == "" || fingerprint == "" || units <= 0 {
		return Sow{}, ErrInvalid
	}
	// A sow may only be created in a state a screening decision can produce.
	// Rejected is not one: a refused sow is one that was held and then
	// refused, and it must pass through pending so its seed can be refunded.
	if status != StatusDelivered && status != StatusPendingReview {
		return Sow{}, ErrInvalid
	}
	if strings.TrimSpace(screeningRef) == "" {
		return Sow{}, ErrInvalid
	}
	if len(media) > 4 {
		return Sow{}, ErrInvalid
	}
	for _, item := range media {
		if item.Key == "" || item.ScreeningKey == "" {
			return Sow{}, ErrInvalid
		}
	}
	return Sow{
		ID: id, ActorKey: actorKey, Body: strings.TrimSpace(body),
		Media: append([]Media(nil), media...), CommandID: commandID,
		Fingerprint: fingerprint, AllowanceUnits: units,
		Status: status, ScreeningRef: screeningRef, AcceptedAt: at.UTC(),
	}, nil
}

// Reconstitute rebuilds a stored sow without re-running the invariants that
// were enforced on the way in. It is separate from Accept because Accept
// refuses to create a rejected sow — correctly, since a rejection is a
// decision and not a starting state — and a repository still has to be able
// to read one back.
func Reconstitute(
	id, actorKey, body string, media []Media, commandID, fingerprint string,
	units int64, status Status, screeningRef string, acceptedAt time.Time, decidedAt *time.Time,
) Sow {
	return Sow{
		ID: id, ActorKey: actorKey, Body: body,
		Media: append([]Media(nil), media...), CommandID: commandID,
		Fingerprint: fingerprint, AllowanceUnits: units,
		Status: status, ScreeningRef: screeningRef,
		AcceptedAt: acceptedAt, DecidedAt: decidedAt,
	}
}

// Release delivers a held sow after a reviewer allowed it.
func (sow Sow) Release(reference string, at time.Time) (Sow, error) {
	return sow.decide(StatusDelivered, reference, at)
}

// Refuse rejects a held sow. The caller is responsible for refunding the
// seed: the aggregate records that it is owed, not that it was paid.
func (sow Sow) Refuse(reference string, at time.Time) (Sow, error) {
	return sow.decide(StatusRejected, reference, at)
}

func (sow Sow) decide(status Status, reference string, at time.Time) (Sow, error) {
	if sow.Status != StatusPendingReview {
		// Deciding a sow twice would refund a seed twice, or deliver one
		// that was already refused.
		return Sow{}, ErrNotPending
	}
	if strings.TrimSpace(reference) == "" || at.IsZero() {
		return Sow{}, ErrInvalid
	}
	decided := at.UTC()
	next := sow
	next.Media = append([]Media(nil), sow.Media...)
	next.Status = status
	next.ScreeningRef = strings.TrimSpace(reference)
	next.DecidedAt = &decided
	return next, nil
}
