// Package domain models the manual-assisted P0 vouch lifecycle.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Status string
type Action string
type Decision string

const (
	StatusAwaitingConsent Status = "awaiting_consent"
	StatusConsented       Status = "consented"
	StatusApproved        Status = "approved"
	StatusDeclined        Status = "declined"
	StatusWithdrawn       Status = "withdrawn"
	StatusExpired         Status = "expired"

	ActionRequested Action = "requested"
	ActionConsented Action = "consented"
	ActionApproved  Action = "approved"
	ActionDeclined  Action = "declined"
	ActionWithdrawn Action = "withdrawn"
	ActionExpired   Action = "expired"

	DecisionApprove Decision = "approve"
	DecisionDecline Decision = "decline"
)

var (
	ErrInvalidRequest    = errors.New("invalid assisted vouch request")
	ErrInvalidTransition = errors.New("invalid assisted vouch transition")
	ErrRequestExpired    = errors.New("assisted vouch request expired")
	ErrStaleRevision     = errors.New("stale assisted vouch revision")
	ErrCommandMismatch   = errors.New("assisted vouch command replay mismatch")
)

var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
	digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Command struct {
	ID               string
	ActorKey         string
	ExpectedRevision uint64
	ReasonCode       string
	At               time.Time
}

type Event struct {
	Sequence   uint64
	CommandID  string
	ActorKey   string
	Action     Action
	ReasonCode string
	At         time.Time
}

type AppliedCommand struct {
	ID          string
	Fingerprint string
	Revision    uint64
}

type Outcome struct {
	Decision    Decision
	ReasonCode  string
	OperatorKey string
	Provenance  string
	DecidedAt   time.Time
}

type Request struct {
	id           string
	subjectKey   string
	requesterKey string
	voucherKey   string
	status       Status
	expiresAt    time.Time
	consentedAt  *time.Time
	outcome      *Outcome
	revision     uint64
	events       []Event
	commands     []AppliedCommand
}

type State struct {
	ID           string
	SubjectKey   string
	RequesterKey string
	VoucherKey   string
	Status       Status
	ExpiresAt    time.Time
	ConsentedAt  *time.Time
	Outcome      *Outcome
	Revision     uint64
	Events       []Event
	Commands     []AppliedCommand
}

func NewRequest(id, subjectKey, requesterKey, voucherKey string, expiresAt time.Time, command Command) (Request, error) {
	if !validOpaque(id) || !validKey(subjectKey) || !validKey(requesterKey) || !validKey(voucherKey) ||
		subjectKey == voucherKey || expiresAt.IsZero() || !expiresAt.After(command.At) ||
		expiresAt.After(command.At.Add(14*24*time.Hour)) || command.ExpectedRevision != 0 {
		return Request{}, ErrInvalidRequest
	}
	request := Request{
		id: id, subjectKey: subjectKey, requesterKey: requesterKey,
		voucherKey: voucherKey, expiresAt: expiresAt.UTC(),
	}
	return request.transition(ActionRequested, command)
}

func Rehydrate(state State) (Request, error) {
	request := Request{
		id: state.ID, subjectKey: state.SubjectKey, requesterKey: state.RequesterKey,
		voucherKey: state.VoucherKey, status: state.Status, expiresAt: state.ExpiresAt.UTC(),
		consentedAt: cloneTime(state.ConsentedAt), outcome: cloneOutcome(state.Outcome),
		revision: state.Revision, events: append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if !validOpaque(request.id) || !validKey(request.subjectKey) || !validKey(request.requesterKey) ||
		!validKey(request.voucherKey) || request.revision == 0 ||
		uint64(len(request.events)) != request.revision || uint64(len(request.commands)) != request.revision {
		return Request{}, ErrInvalidRequest
	}
	status := Status("")
	seen := make(map[string]struct{}, len(request.commands))
	var consentedAt *time.Time
	var outcome *Outcome
	var previousAt time.Time
	for index, event := range request.events {
		applied := request.commands[index]
		if event.Sequence != uint64(index+1) || applied.Revision != event.Sequence ||
			applied.ID != event.CommandID || !validEvent(event) ||
			(!previousAt.IsZero() && event.At.Before(previousAt)) {
			return Request{}, ErrInvalidRequest
		}
		if _, duplicate := seen[applied.ID]; duplicate {
			return Request{}, ErrInvalidRequest
		}
		seen[applied.ID] = struct{}{}
		command := Command{
			ID: event.CommandID, ActorKey: event.ActorKey, ExpectedRevision: uint64(index),
			ReasonCode: event.ReasonCode, At: event.At,
		}
		if applied.Fingerprint != fingerprint(request.id, event.Action, command) {
			return Request{}, ErrInvalidRequest
		}
		switch event.Action {
		case ActionRequested:
			if index != 0 {
				return Request{}, ErrInvalidRequest
			}
			status = StatusAwaitingConsent
		case ActionConsented:
			if status != StatusAwaitingConsent {
				return Request{}, ErrInvalidRequest
			}
			status = StatusConsented
			value := event.At
			consentedAt = &value
		case ActionApproved, ActionDeclined:
			if status != StatusConsented {
				return Request{}, ErrInvalidRequest
			}
			status = StatusApproved
			decision := DecisionApprove
			if event.Action == ActionDeclined {
				status, decision = StatusDeclined, DecisionDecline
			}
			outcome = &Outcome{
				Decision: decision, ReasonCode: event.ReasonCode, OperatorKey: event.ActorKey,
				Provenance: "manual_assisted", DecidedAt: event.At,
			}
		case ActionWithdrawn:
			if status != StatusAwaitingConsent && status != StatusConsented {
				return Request{}, ErrInvalidRequest
			}
			status = StatusWithdrawn
		case ActionExpired:
			if status != StatusAwaitingConsent && status != StatusConsented {
				return Request{}, ErrInvalidRequest
			}
			status = StatusExpired
		default:
			return Request{}, ErrInvalidRequest
		}
		previousAt = event.At
	}
	if status != request.status || !equalTime(consentedAt, request.consentedAt) || !equalOutcome(outcome, request.outcome) {
		return Request{}, ErrInvalidRequest
	}
	return request, nil
}

func (request Request) Consent(command Command) (Request, error) {
	if !command.At.Before(request.expiresAt) {
		return Request{}, ErrRequestExpired
	}
	return request.change(ActionConsented, command, StatusAwaitingConsent)
}

func (request Request) Decide(decision Decision, command Command) (Request, error) {
	if !command.At.Before(request.expiresAt) {
		return Request{}, ErrRequestExpired
	}
	action := ActionApproved
	if decision == DecisionDecline {
		action = ActionDeclined
	} else if decision != DecisionApprove {
		return Request{}, ErrInvalidRequest
	}
	return request.change(action, command, StatusConsented)
}

func (request Request) Withdraw(command Command) (Request, error) {
	if replayed, err := request.replay(command, ActionWithdrawn); replayed || err != nil {
		return request, err
	}
	if request.status != StatusAwaitingConsent && request.status != StatusConsented {
		return Request{}, ErrInvalidTransition
	}
	return request.transition(ActionWithdrawn, command)
}

func (request Request) Expire(command Command) (Request, error) {
	if replayed, err := request.replay(command, ActionExpired); replayed || err != nil {
		return request, err
	}
	if command.At.Before(request.expiresAt) ||
		(request.status != StatusAwaitingConsent && request.status != StatusConsented) {
		return Request{}, ErrInvalidTransition
	}
	return request.transition(ActionExpired, command)
}

func (request Request) change(action Action, command Command, required Status) (Request, error) {
	if replayed, err := request.replay(command, action); replayed || err != nil {
		return request, err
	}
	if request.status != required {
		return Request{}, ErrInvalidTransition
	}
	return request.transition(action, command)
}

func (request Request) transition(action Action, command Command) (Request, error) {
	if err := validateCommand(command, request.revision); err != nil {
		return Request{}, err
	}
	if len(request.events) > 0 && command.At.Before(request.events[len(request.events)-1].At) {
		return Request{}, ErrInvalidRequest
	}
	request.revision++
	event := Event{
		Sequence: request.revision, CommandID: command.ID, ActorKey: command.ActorKey,
		Action: action, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	}
	request.events = append(request.events, event)
	request.commands = append(request.commands, AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(request.id, action, command), Revision: request.revision,
	})
	switch action {
	case ActionRequested:
		request.status = StatusAwaitingConsent
	case ActionConsented:
		request.status = StatusConsented
		value := event.At
		request.consentedAt = &value
	case ActionApproved, ActionDeclined:
		request.status = StatusApproved
		decision := DecisionApprove
		if action == ActionDeclined {
			request.status, decision = StatusDeclined, DecisionDecline
		}
		request.outcome = &Outcome{
			Decision: decision, ReasonCode: event.ReasonCode, OperatorKey: event.ActorKey,
			Provenance: "manual_assisted", DecidedAt: event.At,
		}
	case ActionWithdrawn:
		request.status = StatusWithdrawn
	case ActionExpired:
		request.status = StatusExpired
	}
	return request, nil
}

func (request Request) replay(command Command, action Action) (bool, error) {
	expected := fingerprint(request.id, action, command)
	for _, applied := range request.commands {
		if applied.ID != command.ID {
			continue
		}
		if applied.Fingerprint != expected {
			return false, ErrCommandMismatch
		}
		return true, nil
	}
	return false, nil
}

func validateCommand(command Command, revision uint64) error {
	if !validOpaque(command.ID) || !validKey(command.ActorKey) ||
		!reasonPattern.MatchString(command.ReasonCode) || command.At.IsZero() {
		return ErrInvalidRequest
	}
	if command.ExpectedRevision != revision {
		return ErrStaleRevision
	}
	return nil
}

func fingerprint(requestID string, action Action, command Command) string {
	value := requestID + "\x00" + string(action) + "\x00" + command.ActorKey + "\x00" +
		command.ReasonCode + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validOpaque(value string) bool { return opaquePattern.MatchString(strings.TrimSpace(value)) }
func validKey(value string) bool    { return digestPattern.MatchString(value) }
func validEvent(event Event) bool {
	return event.Sequence > 0 && validOpaque(event.CommandID) && validKey(event.ActorKey) &&
		reasonPattern.MatchString(event.ReasonCode) && !event.At.IsZero()
}
func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
func cloneOutcome(value *Outcome) *Outcome {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.DecidedAt = cloned.DecidedAt.UTC()
	return &cloned
}
func equalTime(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}
func equalOutcome(left, right *Outcome) bool {
	return left == nil && right == nil || left != nil && right != nil &&
		left.Decision == right.Decision && left.ReasonCode == right.ReasonCode &&
		left.OperatorKey == right.OperatorKey && left.Provenance == right.Provenance &&
		left.DecidedAt.Equal(right.DecidedAt)
}

func (request Request) ID() string              { return request.id }
func (request Request) SubjectKey() string      { return request.subjectKey }
func (request Request) RequesterKey() string    { return request.requesterKey }
func (request Request) VoucherKey() string      { return request.voucherKey }
func (request Request) Status() Status          { return request.status }
func (request Request) ExpiresAt() time.Time    { return request.expiresAt }
func (request Request) ConsentedAt() *time.Time { return cloneTime(request.consentedAt) }
func (request Request) Outcome() *Outcome       { return cloneOutcome(request.outcome) }
func (request Request) Revision() uint64        { return request.revision }
func (request Request) Events() []Event         { return append([]Event(nil), request.events...) }
func (request Request) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), request.commands...)
}
func (request Request) HasCommand(id string) bool {
	for _, command := range request.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}
