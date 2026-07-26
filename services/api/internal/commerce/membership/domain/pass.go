package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid membership pass")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

const MaxGrace = 14 * 24 * time.Hour

type Status string

const (
	StatusActive        Status = "active"
	StatusGrace         Status = "grace"
	StatusExpired       Status = "expired"
	StatusRefundPending Status = "refund_pending"
	StatusRefunded      Status = "refunded"
)

type EventKind string

const (
	EventGranted         EventKind = "granted"
	EventCancelled       EventKind = "cancelled"
	EventRefundRequested EventKind = "refund_requested"
	EventRefundConfirmed EventKind = "refund_confirmed"
)

type Event struct {
	Sequence  uint64
	Kind      EventKind
	CommandID string
	At        time.Time
	Reference string
}

type Grant struct {
	ID            string
	MemberKey     string
	PassID        string
	PassVersion   uint64
	ReceiptRef    string
	GrantedAt     time.Time
	PaidThrough   time.Time
	GraceUntil    time.Time
	GraceDuration time.Duration
}

type State struct {
	Grant              `bson:",inline"`
	CancelledAt        time.Time
	RefundRequestRef   string
	RefundConfirmedRef string
	Revision           uint64
	Events             []Event
	AppliedIDs         []string
}

type Pass struct{ state State }

func New(grant Grant, commandID string) (Pass, error) {
	grant.GrantedAt = grant.GrantedAt.UTC()
	grant.PaidThrough = grant.PaidThrough.UTC()
	grant.GraceUntil = grant.GraceUntil.UTC()
	if !validGrant(grant) || !token.MatchString(commandID) {
		return Pass{}, ErrInvalid
	}
	state := State{Grant: grant, Revision: 1, AppliedIDs: []string{commandID}}
	state.Events = []Event{{Sequence: 1, Kind: EventGranted, CommandID: commandID, At: grant.GrantedAt, Reference: grant.ReceiptRef}}
	return Pass{state: state}, nil
}

func Rehydrate(state State) (Pass, error) {
	state.Events = append([]Event(nil), state.Events...)
	state.AppliedIDs = append([]string(nil), state.AppliedIDs...)
	if !validState(state) {
		return Pass{}, ErrInvalid
	}
	return Pass{state: state}, nil
}

func (pass Pass) Cancel(commandID string, at time.Time) (Pass, error) {
	if !pass.state.CancelledAt.IsZero() || at.IsZero() || at.Before(pass.state.GrantedAt) {
		return Pass{}, ErrInvalid
	}
	next, err := pass.transition(EventCancelled, commandID, "", at)
	if err == nil {
		next.state.CancelledAt = at.UTC()
	}
	return next, err
}

func (pass Pass) RequestRefund(commandID, requestRef string, at time.Time) (Pass, error) {
	if pass.state.CancelledAt.IsZero() || pass.state.RefundRequestRef != "" ||
		!opaque.MatchString(requestRef) {
		return Pass{}, ErrInvalid
	}
	next, err := pass.transition(EventRefundRequested, commandID, requestRef, at)
	if err == nil {
		next.state.RefundRequestRef = requestRef
	}
	return next, err
}

func (pass Pass) ConfirmRefund(commandID, requestRef, providerRef string, at time.Time) (Pass, error) {
	if pass.state.RefundRequestRef == "" || pass.state.RefundRequestRef != requestRef ||
		pass.state.RefundConfirmedRef != "" || !opaque.MatchString(providerRef) {
		return Pass{}, ErrInvalid
	}
	next, err := pass.transition(EventRefundConfirmed, commandID, providerRef, at)
	if err == nil {
		next.state.RefundConfirmedRef = providerRef
	}
	return next, err
}

func (pass Pass) Status(at time.Time) Status {
	if pass.state.RefundConfirmedRef != "" {
		return StatusRefunded
	}
	if pass.state.RefundRequestRef != "" {
		return StatusRefundPending
	}
	if at.Before(pass.state.PaidThrough) || at.Equal(pass.state.PaidThrough) {
		return StatusActive
	}
	if at.Before(pass.state.GraceUntil) || at.Equal(pass.state.GraceUntil) {
		return StatusGrace
	}
	return StatusExpired
}

func (pass Pass) transition(kind EventKind, commandID, reference string, at time.Time) (Pass, error) {
	if !token.MatchString(commandID) || at.IsZero() || slices.Contains(pass.state.AppliedIDs, commandID) {
		return Pass{}, ErrInvalid
	}
	next := pass.State()
	next.Revision++
	next.AppliedIDs = append(next.AppliedIDs, commandID)
	next.Events = append(next.Events, Event{
		Sequence: next.Revision, Kind: kind, CommandID: commandID,
		At: at.UTC(), Reference: reference,
	})
	return Pass{state: next}, nil
}

func (pass Pass) State() State {
	state := pass.state
	state.Events = append([]Event(nil), pass.state.Events...)
	state.AppliedIDs = append([]string(nil), pass.state.AppliedIDs...)
	return state
}
func (pass Pass) ID() string       { return pass.state.ID }
func (pass Pass) Revision() uint64 { return pass.state.Revision }

func validGrant(grant Grant) bool {
	return opaque.MatchString(grant.ID) && opaque.MatchString(grant.MemberKey) &&
		token.MatchString(grant.PassID) && grant.PassVersion > 0 &&
		opaque.MatchString(grant.ReceiptRef) && !grant.GrantedAt.IsZero() &&
		grant.PaidThrough.After(grant.GrantedAt) &&
		grant.GraceDuration >= 0 && grant.GraceDuration <= MaxGrace &&
		grant.GraceUntil.Equal(grant.PaidThrough.Add(grant.GraceDuration))
}

func validState(state State) bool {
	if !validGrant(state.Grant) || state.Revision == 0 ||
		len(state.Events) != int(state.Revision) ||
		len(state.AppliedIDs) != int(state.Revision) {
		return false
	}
	for index, event := range state.Events {
		if event.Sequence != uint64(index+1) || event.CommandID != state.AppliedIDs[index] ||
			!token.MatchString(event.CommandID) || event.At.IsZero() {
			return false
		}
	}
	if state.RefundConfirmedRef != "" &&
		(state.RefundRequestRef == "" || !opaque.MatchString(state.RefundConfirmedRef)) {
		return false
	}
	return true
}
