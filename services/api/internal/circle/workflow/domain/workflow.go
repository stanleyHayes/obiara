// Package domain models circle admission and moderation workflows without
// storing invitation secrets.
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

type InviteStatus string
type RequestStatus string
type Action string

const (
	InviteActive   InviteStatus = "active"
	InviteRedeemed InviteStatus = "redeemed"
	InviteExpired  InviteStatus = "expired"

	RequestPending  RequestStatus = "pending"
	RequestApproved RequestStatus = "approved"
	RequestDeclined RequestStatus = "declined"
	RequestExpelled RequestStatus = "expelled"

	ActionInviteCreated  Action = "invite_created"
	ActionInviteRedeemed Action = "invite_redeemed"
	ActionInviteExpired  Action = "invite_expired"
	ActionRequested      Action = "requested"
	ActionApproved       Action = "approved"
	ActionDeclined       Action = "declined"
	ActionExpelled       Action = "expelled"
)

var (
	ErrInvalidWorkflow   = errors.New("invalid circle workflow")
	ErrInviteExpired     = errors.New("circle invite expired")
	ErrInviteUsed        = errors.New("circle invite already used")
	ErrInvalidTransition = errors.New("invalid circle workflow transition")
	ErrStaleRevision     = errors.New("stale circle workflow revision")
	ErrCommandMismatch   = errors.New("circle workflow command replay mismatch")
)

var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Command struct {
	ID               string
	ActorID          string
	ExpectedRevision uint64
	ReasonCode       string
	At               time.Time
}

type Event struct {
	Sequence   uint64
	CommandID  string
	ActorID    string
	Action     Action
	ReasonCode string
	At         time.Time
}

type AppliedCommand struct {
	ID          string
	Fingerprint string
	Revision    uint64
}

type Invite struct {
	id          string
	circleID    string
	tokenDigest string
	status      InviteStatus
	expiresAt   time.Time
	revision    uint64
	events      []Event
	commands    []AppliedCommand
}

type InviteState struct {
	ID          string
	CircleID    string
	TokenDigest string
	Status      InviteStatus
	ExpiresAt   time.Time
	Revision    uint64
	Events      []Event
	Commands    []AppliedCommand
}

func NewInvite(id, circleID, tokenDigest string, expiresAt time.Time, command Command) (Invite, error) {
	if !validOpaque(id) || !validOpaque(circleID) || !validDigest(tokenDigest) ||
		expiresAt.IsZero() || !expiresAt.After(command.At) || expiresAt.After(command.At.Add(7*24*time.Hour)) ||
		command.ExpectedRevision != 0 {
		return Invite{}, ErrInvalidWorkflow
	}
	invite := Invite{id: id, circleID: circleID, tokenDigest: tokenDigest, status: InviteActive, expiresAt: expiresAt.UTC()}
	return invite.transition(ActionInviteCreated, command)
}

func RehydrateInvite(state InviteState) (Invite, error) {
	invite := Invite{
		id: state.ID, circleID: state.CircleID, tokenDigest: state.TokenDigest,
		status: state.Status, expiresAt: state.ExpiresAt.UTC(), revision: state.Revision,
		events: append([]Event(nil), state.Events...), commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if !validOpaque(invite.id) || !validOpaque(invite.circleID) || !validDigest(invite.tokenDigest) ||
		invite.expiresAt.IsZero() || invite.revision == 0 ||
		uint64(len(invite.events)) != invite.revision || uint64(len(invite.commands)) != invite.revision {
		return Invite{}, ErrInvalidWorkflow
	}
	status := InviteStatus("")
	seenCommands := make(map[string]struct{}, len(invite.commands))
	var previousAt time.Time
	for index, event := range invite.events {
		command := invite.commands[index]
		if event.Sequence != uint64(index+1) || command.Revision != event.Sequence ||
			event.CommandID != command.ID || !validEvent(event) ||
			(!previousAt.IsZero() && event.At.Before(previousAt)) {
			return Invite{}, ErrInvalidWorkflow
		}
		if _, duplicate := seenCommands[command.ID]; duplicate {
			return Invite{}, ErrInvalidWorkflow
		}
		seenCommands[command.ID] = struct{}{}
		expected := fingerprint(invite.id, event.Action, Command{
			ID: command.ID, ActorID: event.ActorID, ExpectedRevision: uint64(index),
			ReasonCode: event.ReasonCode, At: event.At,
		})
		if command.Fingerprint != expected {
			return Invite{}, ErrInvalidWorkflow
		}
		previousAt = event.At
		switch event.Action {
		case ActionInviteCreated:
			if index != 0 {
				return Invite{}, ErrInvalidWorkflow
			}
			status = InviteActive
		case ActionInviteRedeemed:
			if status != InviteActive {
				return Invite{}, ErrInvalidWorkflow
			}
			status = InviteRedeemed
		case ActionInviteExpired:
			if status != InviteActive {
				return Invite{}, ErrInvalidWorkflow
			}
			status = InviteExpired
		default:
			return Invite{}, ErrInvalidWorkflow
		}
	}
	if status != invite.status {
		return Invite{}, ErrInvalidWorkflow
	}
	return invite, nil
}

func (invite Invite) Redeem(command Command) (Invite, error) {
	if replay, err := invite.replay(command, ActionInviteRedeemed); replay || err != nil {
		return invite, err
	}
	if invite.status != InviteActive {
		return Invite{}, ErrInviteUsed
	}
	if !command.At.Before(invite.expiresAt) {
		return Invite{}, ErrInviteExpired
	}
	return invite.transition(ActionInviteRedeemed, command)
}

func (invite Invite) Expire(command Command) (Invite, error) {
	if replay, err := invite.replay(command, ActionInviteExpired); replay || err != nil {
		return invite, err
	}
	if invite.status != InviteActive || command.At.Before(invite.expiresAt) {
		return Invite{}, ErrInvalidTransition
	}
	return invite.transition(ActionInviteExpired, command)
}

func (invite Invite) transition(action Action, command Command) (Invite, error) {
	if err := validateCommand(command, invite.revision); err != nil {
		return Invite{}, err
	}
	if len(invite.events) > 0 && command.At.Before(invite.events[len(invite.events)-1].At) {
		return Invite{}, ErrInvalidWorkflow
	}
	fingerprint := fingerprint(invite.id, action, command)
	invite.revision++
	invite.events = append(invite.events, Event{
		Sequence: invite.revision, CommandID: command.ID, ActorID: command.ActorID,
		Action: action, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	})
	invite.commands = append(invite.commands, AppliedCommand{ID: command.ID, Fingerprint: fingerprint, Revision: invite.revision})
	switch action {
	case ActionInviteCreated:
		invite.status = InviteActive
	case ActionInviteRedeemed:
		invite.status = InviteRedeemed
	case ActionInviteExpired:
		invite.status = InviteExpired
	}
	return invite, nil
}

func (invite Invite) replay(command Command, action Action) (bool, error) {
	return replay(invite.id, invite.commands, command, action)
}

type Request struct {
	id       string
	circleID string
	memberID string
	source   string
	status   RequestStatus
	revision uint64
	events   []Event
	commands []AppliedCommand
}

type RequestState struct {
	ID       string
	CircleID string
	MemberID string
	Source   string
	Status   RequestStatus
	Revision uint64
	Events   []Event
	Commands []AppliedCommand
}

func NewRequest(id, circleID, memberID, source string, command Command) (Request, error) {
	if !validOpaque(id) || !validOpaque(circleID) || !validOpaque(memberID) ||
		(source != "direct" && source != "invite") || command.ExpectedRevision != 0 {
		return Request{}, ErrInvalidWorkflow
	}
	request := Request{id: id, circleID: circleID, memberID: memberID, source: source}
	return request.transition(ActionRequested, command)
}

func RehydrateRequest(state RequestState) (Request, error) {
	request := Request{
		id: state.ID, circleID: state.CircleID, memberID: state.MemberID,
		source: state.Source, status: state.Status, revision: state.Revision,
		events: append([]Event(nil), state.Events...), commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if !validOpaque(request.id) || !validOpaque(request.circleID) || !validOpaque(request.memberID) ||
		(request.source != "direct" && request.source != "invite") || request.revision == 0 ||
		uint64(len(request.events)) != request.revision || uint64(len(request.commands)) != request.revision {
		return Request{}, ErrInvalidWorkflow
	}
	status := RequestStatus("")
	seenCommands := make(map[string]struct{}, len(request.commands))
	var previousAt time.Time
	for index, event := range request.events {
		command := request.commands[index]
		if event.Sequence != uint64(index+1) || command.Revision != event.Sequence ||
			event.CommandID != command.ID || !validEvent(event) ||
			(!previousAt.IsZero() && event.At.Before(previousAt)) {
			return Request{}, ErrInvalidWorkflow
		}
		if _, duplicate := seenCommands[command.ID]; duplicate {
			return Request{}, ErrInvalidWorkflow
		}
		seenCommands[command.ID] = struct{}{}
		expected := fingerprint(request.id, event.Action, Command{
			ID: command.ID, ActorID: event.ActorID, ExpectedRevision: uint64(index),
			ReasonCode: event.ReasonCode, At: event.At,
		})
		if command.Fingerprint != expected {
			return Request{}, ErrInvalidWorkflow
		}
		previousAt = event.At
		switch event.Action {
		case ActionRequested:
			if index != 0 {
				return Request{}, ErrInvalidWorkflow
			}
			status = RequestPending
		case ActionApproved:
			if status != RequestPending {
				return Request{}, ErrInvalidWorkflow
			}
			status = RequestApproved
		case ActionDeclined:
			if status != RequestPending {
				return Request{}, ErrInvalidWorkflow
			}
			status = RequestDeclined
		case ActionExpelled:
			if status != RequestApproved {
				return Request{}, ErrInvalidWorkflow
			}
			status = RequestExpelled
		default:
			return Request{}, ErrInvalidWorkflow
		}
	}
	if status != request.status {
		return Request{}, ErrInvalidWorkflow
	}
	return request, nil
}

func (request Request) Approve(command Command) (Request, error) {
	return request.change(ActionApproved, command, RequestPending)
}
func (request Request) Decline(command Command) (Request, error) {
	return request.change(ActionDeclined, command, RequestPending)
}
func (request Request) Expel(command Command) (Request, error) {
	return request.change(ActionExpelled, command, RequestApproved)
}

func (request Request) change(action Action, command Command, required RequestStatus) (Request, error) {
	if replayed, err := replay(request.id, request.commands, command, action); replayed || err != nil {
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
		return Request{}, ErrInvalidWorkflow
	}
	request.revision++
	request.events = append(request.events, Event{
		Sequence: request.revision, CommandID: command.ID, ActorID: command.ActorID,
		Action: action, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	})
	request.commands = append(request.commands, AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(request.id, action, command), Revision: request.revision,
	})
	switch action {
	case ActionRequested:
		request.status = RequestPending
	case ActionApproved:
		request.status = RequestApproved
	case ActionDeclined:
		request.status = RequestDeclined
	case ActionExpelled:
		request.status = RequestExpelled
	}
	return request, nil
}

func validateCommand(command Command, revision uint64) error {
	if !validOpaque(strings.TrimSpace(command.ID)) || !validOpaque(strings.TrimSpace(command.ActorID)) ||
		!reasonPattern.MatchString(strings.TrimSpace(command.ReasonCode)) || command.At.IsZero() {
		return ErrInvalidWorkflow
	}
	if command.ExpectedRevision != revision {
		return ErrStaleRevision
	}
	return nil
}

func replay(aggregateID string, commands []AppliedCommand, command Command, action Action) (bool, error) {
	expected := fingerprint(aggregateID, action, command)
	for _, applied := range commands {
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

func fingerprint(aggregateID string, action Action, command Command) string {
	value := aggregateID + "\x00" + string(action) + "\x00" + command.ActorID + "\x00" +
		command.ReasonCode + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validOpaque(value string) bool { return opaquePattern.MatchString(strings.TrimSpace(value)) }
func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
func validEvent(event Event) bool {
	return event.Sequence > 0 && validOpaque(event.CommandID) && validOpaque(event.ActorID) &&
		reasonPattern.MatchString(event.ReasonCode) && !event.At.IsZero()
}

func (invite Invite) ID() string           { return invite.id }
func (invite Invite) CircleID() string     { return invite.circleID }
func (invite Invite) TokenDigest() string  { return invite.tokenDigest }
func (invite Invite) Status() InviteStatus { return invite.status }
func (invite Invite) ExpiresAt() time.Time { return invite.expiresAt }
func (invite Invite) Revision() uint64     { return invite.revision }
func (invite Invite) Events() []Event      { return append([]Event(nil), invite.events...) }
func (invite Invite) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), invite.commands...)
}
func (invite Invite) HasCommand(id string) bool {
	for _, command := range invite.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}
func (request Request) ID() string            { return request.id }
func (request Request) CircleID() string      { return request.circleID }
func (request Request) MemberID() string      { return request.memberID }
func (request Request) Source() string        { return request.source }
func (request Request) Status() RequestStatus { return request.status }
func (request Request) Revision() uint64      { return request.revision }
func (request Request) Events() []Event       { return append([]Event(nil), request.events...) }
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
