// Package domain models bounded, explicit introduction-source requests.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SourceType string
type Status string
type Action string

const (
	SourceCircle    SourceType = "consented_circle"
	SourceTrust     SourceType = "consented_trust"
	SourceHost      SourceType = "consented_host"
	SourceCohort    SourceType = "policy_cohort"
	StatusOpen      Status     = "open"
	StatusWithdrawn Status     = "withdrawn"
	StatusExpired   Status     = "expired"
	ActionOpened    Action     = "opened"
	ActionWithdrawn Action     = "withdrawn"
	ActionExpired   Action     = "expired"
	MaxCandidates              = 50
)

var (
	ErrInvalidRequest    = errors.New("invalid introduction source request")
	ErrInvalidTransition = errors.New("invalid introduction source transition")
	ErrStaleRevision     = errors.New("stale introduction source revision")
	ErrCommandMismatch   = errors.New("introduction source command replay mismatch")
)
var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	keyPattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Source struct {
	Type SourceType
	Key  string
}
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
type Request struct {
	id           string
	requesterKey string
	source       Source
	candidateIDs []string
	status       Status
	expiresAt    time.Time
	endedAt      *time.Time
	revision     uint64
	events       []Event
	commands     []AppliedCommand
}
type State struct {
	ID           string
	RequesterKey string
	Source       Source
	CandidateIDs []string
	Status       Status
	ExpiresAt    time.Time
	EndedAt      *time.Time
	Revision     uint64
	Events       []Event
	Commands     []AppliedCommand
}

func Open(id, requesterKey string, source Source, candidates []string, expiresAt time.Time, command Command) (Request, error) {
	normalized, ok := normalizeCandidates(candidates)
	if !validOpaque(id) || !validKey(requesterKey) || !validSource(source) ||
		!ok || expiresAt.IsZero() || !expiresAt.After(command.At) ||
		expiresAt.After(command.At.Add(time.Hour)) || command.ExpectedRevision != 0 {
		return Request{}, ErrInvalidRequest
	}
	request := Request{id: id, requesterKey: requesterKey, source: source, candidateIDs: normalized, expiresAt: expiresAt.UTC()}
	return request.transition(ActionOpened, command)
}
func Rehydrate(state State) (Request, error) {
	candidates, ok := normalizeCandidates(state.CandidateIDs)
	request := Request{
		id: state.ID, requesterKey: state.RequesterKey, source: state.Source,
		candidateIDs: candidates, status: state.Status, expiresAt: state.ExpiresAt.UTC(),
		endedAt: cloneTime(state.EndedAt), revision: state.Revision,
		events: append([]Event(nil), state.Events...), commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if !ok || !validOpaque(request.id) || !validKey(request.requesterKey) || !validSource(request.source) ||
		request.revision == 0 || uint64(len(request.events)) != request.revision ||
		uint64(len(request.commands)) != request.revision {
		return Request{}, ErrInvalidRequest
	}
	status := Status("")
	var endedAt *time.Time
	seen := map[string]struct{}{}
	var previous time.Time
	for index, event := range request.events {
		applied := request.commands[index]
		if event.Sequence != uint64(index+1) || applied.Revision != event.Sequence ||
			applied.ID != event.CommandID || !validEvent(event) || (!previous.IsZero() && event.At.Before(previous)) {
			return Request{}, ErrInvalidRequest
		}
		if _, duplicate := seen[applied.ID]; duplicate {
			return Request{}, ErrInvalidRequest
		}
		seen[applied.ID] = struct{}{}
		command := Command{ID: event.CommandID, ActorKey: event.ActorKey, ExpectedRevision: uint64(index), ReasonCode: event.ReasonCode, At: event.At}
		if applied.Fingerprint != fingerprint(request.id, event.Action, command) {
			return Request{}, ErrInvalidRequest
		}
		switch event.Action {
		case ActionOpened:
			if index != 0 {
				return Request{}, ErrInvalidRequest
			}
			status = StatusOpen
		case ActionWithdrawn:
			if status != StatusOpen {
				return Request{}, ErrInvalidRequest
			}
			status = StatusWithdrawn
			value := event.At
			endedAt = &value
		case ActionExpired:
			if status != StatusOpen {
				return Request{}, ErrInvalidRequest
			}
			status = StatusExpired
			value := event.At
			endedAt = &value
		default:
			return Request{}, ErrInvalidRequest
		}
		previous = event.At
	}
	if status != request.status || !equalTime(endedAt, request.endedAt) {
		return Request{}, ErrInvalidRequest
	}
	return request, nil
}
func (r Request) Withdraw(command Command) (Request, error) {
	return r.change(ActionWithdrawn, command, false)
}
func (r Request) Expire(command Command) (Request, error) {
	return r.change(ActionExpired, command, true)
}
func (r Request) change(action Action, command Command, requireExpired bool) (Request, error) {
	if replayed, err := r.replay(command, action); replayed || err != nil {
		return r, err
	}
	if r.status != StatusOpen || requireExpired && command.At.Before(r.expiresAt) {
		return Request{}, ErrInvalidTransition
	}
	return r.transition(action, command)
}
func (r Request) transition(action Action, command Command) (Request, error) {
	if err := validateCommand(command, r.revision); err != nil {
		return Request{}, err
	}
	if len(r.events) > 0 && command.At.Before(r.events[len(r.events)-1].At) {
		return Request{}, ErrInvalidRequest
	}
	r.revision++
	event := Event{Sequence: r.revision, CommandID: command.ID, ActorKey: command.ActorKey, Action: action, ReasonCode: command.ReasonCode, At: command.At.UTC()}
	r.events = append(r.events, event)
	r.commands = append(r.commands, AppliedCommand{ID: command.ID, Fingerprint: fingerprint(r.id, action, command), Revision: r.revision})
	switch action {
	case ActionOpened:
		r.status = StatusOpen
	case ActionWithdrawn:
		r.status = StatusWithdrawn
		value := event.At
		r.endedAt = &value
	case ActionExpired:
		r.status = StatusExpired
		value := event.At
		r.endedAt = &value
	}
	return r, nil
}
func (r Request) replay(command Command, action Action) (bool, error) {
	expected := fingerprint(r.id, action, command)
	for _, applied := range r.commands {
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
func normalizeCandidates(values []string) ([]string, bool) {
	if len(values) > MaxCandidates {
		return nil, false
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !validKey(value) {
			return nil, false
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result, true
}
func validateCommand(command Command, revision uint64) error {
	if !validOpaque(command.ID) || !validKey(command.ActorKey) || !reasonPattern.MatchString(command.ReasonCode) || command.At.IsZero() {
		return ErrInvalidRequest
	}
	if command.ExpectedRevision != revision {
		return ErrStaleRevision
	}
	return nil
}
func fingerprint(id string, action Action, command Command) string {
	value := id + "\x00" + string(action) + "\x00" + command.ActorKey + "\x00" + command.ReasonCode + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func validSource(source Source) bool {
	switch source.Type {
	case SourceCircle, SourceTrust, SourceHost, SourceCohort:
		return validKey(source.Key)
	default:
		return false
	}
}
func validOpaque(value string) bool { return opaquePattern.MatchString(strings.TrimSpace(value)) }
func validKey(value string) bool    { return keyPattern.MatchString(value) }
func validEvent(e Event) bool {
	return e.Sequence > 0 && validOpaque(e.CommandID) && validKey(e.ActorKey) && reasonPattern.MatchString(e.ReasonCode) && !e.At.IsZero()
}
func cloneTime(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	c := v.UTC()
	return &c
}
func equalTime(a, b *time.Time) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Equal(*b)
}
func (r Request) ID() string                 { return r.id }
func (r Request) RequesterKey() string       { return r.requesterKey }
func (r Request) Source() Source             { return r.source }
func (r Request) CandidateIDs() []string     { return append([]string(nil), r.candidateIDs...) }
func (r Request) Status() Status             { return r.status }
func (r Request) ExpiresAt() time.Time       { return r.expiresAt }
func (r Request) EndedAt() *time.Time        { return cloneTime(r.endedAt) }
func (r Request) Revision() uint64           { return r.revision }
func (r Request) Events() []Event            { return append([]Event(nil), r.events...) }
func (r Request) Commands() []AppliedCommand { return append([]AppliedCommand(nil), r.commands...) }
func (r Request) HasCommand(id string) bool {
	for _, c := range r.commands {
		if c.ID == id {
			return true
		}
	}
	return false
}
