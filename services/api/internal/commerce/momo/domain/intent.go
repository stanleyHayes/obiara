package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid momo collection intent")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-zA-Z0-9._:-]{3,128}$`)

const MaxAmountPesewas uint64 = 5_000_000

type Status string

const (
	AwaitingMemberConfirmation   Status = "awaiting_member_confirmation"
	AwaitingProviderConfirmation Status = "awaiting_provider_confirmation"
	Succeeded                    Status = "succeeded"
	Failed                       Status = "failed"
)

type EventKind string

const (
	Created           EventKind = "created"
	MemberConfirmed   EventKind = "member_confirmed"
	ProviderRequested EventKind = "provider_requested"
	ProviderSucceeded EventKind = "provider_succeeded"
	ProviderFailed    EventKind = "provider_failed"
)

type Event struct {
	Sequence    uint64
	Kind        EventKind
	Status      Status
	CommandID   string
	ProviderRef string
	At          time.Time
}

type State struct {
	ID, MemberKey, PhoneRef         string
	AmountPesewas                   uint64
	Status                          Status
	ProviderRequestRef, ProviderRef string
	Revision                        uint64
	Events                          []Event
	AppliedIDs                      []string
}

type Intent struct{ state State }

func Create(id, memberKey, phoneRef string, amount uint64, commandID string, at time.Time) (Intent, error) {
	if !opaque.MatchString(id) || !opaque.MatchString(memberKey) || !opaque.MatchString(phoneRef) ||
		amount == 0 || amount > MaxAmountPesewas || !token.MatchString(commandID) || at.IsZero() {
		return Intent{}, ErrInvalid
	}
	state := State{ID: id, MemberKey: memberKey, PhoneRef: phoneRef, AmountPesewas: amount,
		Status: AwaitingMemberConfirmation, Revision: 1, AppliedIDs: []string{commandID}}
	state.Events = []Event{{Sequence: 1, Kind: Created, Status: state.Status, CommandID: commandID, At: at.UTC()}}
	return Intent{state}, nil
}

func Rehydrate(state State) (Intent, error) {
	state.Events = append([]Event(nil), state.Events...)
	state.AppliedIDs = append([]string(nil), state.AppliedIDs...)
	if !validState(state) {
		return Intent{}, ErrInvalid
	}
	return Intent{state}, nil
}

func (i Intent) Confirm(commandID string, at time.Time) (Intent, error) {
	if i.state.Status != AwaitingMemberConfirmation {
		return Intent{}, ErrInvalid
	}
	return i.transition(MemberConfirmed, AwaitingProviderConfirmation, commandID, "", at)
}

func (i Intent) MarkRequested(commandID, requestRef string, at time.Time) (Intent, error) {
	if i.state.Status != AwaitingProviderConfirmation || i.state.ProviderRequestRef != "" || !opaque.MatchString(requestRef) {
		return Intent{}, ErrInvalid
	}
	next, err := i.transition(ProviderRequested, AwaitingProviderConfirmation, commandID, requestRef, at)
	if err == nil {
		next.state.ProviderRequestRef = requestRef
	}
	return next, err
}

func (i Intent) ApplyProvider(commandID, providerRef string, success bool, at time.Time) (Intent, error) {
	if i.state.Status != AwaitingProviderConfirmation || i.state.ProviderRequestRef == "" || !opaque.MatchString(providerRef) {
		return Intent{}, ErrInvalid
	}
	kind, status := ProviderFailed, Failed
	if success {
		kind, status = ProviderSucceeded, Succeeded
	}
	next, err := i.transition(kind, status, commandID, providerRef, at)
	if err == nil {
		next.state.ProviderRef = providerRef
	}
	return next, err
}

func (i Intent) transition(kind EventKind, status Status, commandID, ref string, at time.Time) (Intent, error) {
	if !token.MatchString(commandID) || at.IsZero() || slices.Contains(i.state.AppliedIDs, commandID) {
		return Intent{}, ErrInvalid
	}
	next := i.State()
	next.Revision++
	next.Status = status
	next.AppliedIDs = append(next.AppliedIDs, commandID)
	next.Events = append(next.Events, Event{Sequence: next.Revision, Kind: kind, Status: status, CommandID: commandID, ProviderRef: ref, At: at.UTC()})
	return Intent{next}, nil
}

func (i Intent) State() State {
	s := i.state
	s.Events = append([]Event(nil), s.Events...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	return s
}
func (i Intent) ID() string       { return i.state.ID }
func (i Intent) Revision() uint64 { return i.state.Revision }

func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.MemberKey) || !opaque.MatchString(s.PhoneRef) || s.AmountPesewas == 0 || s.AmountPesewas > MaxAmountPesewas || s.Revision == 0 || len(s.Events) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	for x, e := range s.Events {
		if e.Sequence != uint64(x+1) || e.At.IsZero() || !token.MatchString(e.CommandID) || e.CommandID != s.AppliedIDs[x] {
			return false
		}
	}
	if s.Status == Succeeded || s.Status == Failed {
		return opaque.MatchString(s.ProviderRef)
	}
	return true
}
