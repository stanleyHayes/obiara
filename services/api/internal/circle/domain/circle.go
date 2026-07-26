package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidCircle      = errors.New("invalid circle")
	ErrInvalidTransition  = errors.New("invalid membership transition")
	ErrMembershipNotFound = errors.New("circle membership not found")
	ErrStaleRevision      = errors.New("stale circle revision")
	ErrCommandMismatch    = errors.New("circle command replay mismatch")
	ErrAccessDenied       = errors.New("circle access denied")
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Type string

const (
	TypeCommunity    Type = "community"
	TypeCampus       Type = "campus"
	TypeProfessional Type = "professional"
	TypeInterest     Type = "interest"
	TypeSupport      Type = "support"
)

type Visibility string

const (
	VisibilityPrivate      Visibility = "private"
	VisibilityDiscoverable Visibility = "discoverable"
)

type MembershipState string

const (
	StateRequested MembershipState = "requested"
	StateMember    MembershipState = "member"
	StateHost      MembershipState = "host"
	StateOwner     MembershipState = "owner"
	StateExpelled  MembershipState = "expelled"
	StateLeft      MembershipState = "left"
)

type Capability string

const (
	CapabilityDiscover Capability = "discover"
	CapabilityView     Capability = "view"
	CapabilityPost     Capability = "post"
	CapabilityManage   Capability = "manage"
)

type Membership struct {
	memberID  string
	state     MembershipState
	updatedAt time.Time
}

func NewMembership(memberID string, state MembershipState, updatedAt time.Time) (Membership, error) {
	memberID = strings.TrimSpace(memberID)
	if !opaquePattern.MatchString(memberID) || updatedAt.IsZero() || !validState(state) {
		return Membership{}, ErrInvalidCircle
	}
	return Membership{memberID: memberID, state: state, updatedAt: updatedAt.UTC()}, nil
}

func (membership Membership) MemberID() string       { return membership.memberID }
func (membership Membership) State() MembershipState { return membership.state }
func (membership Membership) UpdatedAt() time.Time   { return membership.updatedAt }

type Transition struct {
	revision  uint64
	commandID string
	actorID   string
	memberID  string
	from      string
	to        MembershipState
	at        time.Time
}

func NewTransition(revision uint64, commandID, actorID, memberID, from string, to MembershipState, at time.Time) (Transition, error) {
	commandID, actorID, memberID = strings.TrimSpace(commandID), strings.TrimSpace(actorID), strings.TrimSpace(memberID)
	if revision == 0 || !opaquePattern.MatchString(commandID) || !opaquePattern.MatchString(actorID) ||
		!opaquePattern.MatchString(memberID) || !validState(to) || at.IsZero() {
		return Transition{}, ErrInvalidCircle
	}
	if from != "" && !validState(MembershipState(from)) {
		return Transition{}, ErrInvalidCircle
	}
	return Transition{
		revision: revision, commandID: commandID, actorID: actorID, memberID: memberID,
		from: from, to: to, at: at.UTC(),
	}, nil
}

func (event Transition) Revision() uint64    { return event.revision }
func (event Transition) CommandID() string   { return event.commandID }
func (event Transition) ActorID() string     { return event.actorID }
func (event Transition) MemberID() string    { return event.memberID }
func (event Transition) From() string        { return event.from }
func (event Transition) To() MembershipState { return event.to }
func (event Transition) At() time.Time       { return event.at }

type AppliedCommand struct {
	id          string
	fingerprint string
	revision    uint64
}

func NewAppliedCommand(id, fingerprint string, revision uint64) (AppliedCommand, error) {
	id, fingerprint = strings.TrimSpace(id), strings.TrimSpace(fingerprint)
	if !opaquePattern.MatchString(id) || len(fingerprint) != sha256.Size*2 || revision == 0 {
		return AppliedCommand{}, ErrInvalidCircle
	}
	if _, err := hex.DecodeString(fingerprint); err != nil {
		return AppliedCommand{}, ErrInvalidCircle
	}
	return AppliedCommand{id: id, fingerprint: fingerprint, revision: revision}, nil
}

func (command AppliedCommand) ID() string          { return command.id }
func (command AppliedCommand) Fingerprint() string { return command.fingerprint }
func (command AppliedCommand) Revision() uint64    { return command.revision }

type Circle struct {
	id          string
	kind        Type
	visibility  Visibility
	memberships map[string]Membership
	history     []Transition
	commands    []AppliedCommand
	revision    uint64
	updatedAt   time.Time
}

type State struct {
	ID          string
	Type        Type
	Visibility  Visibility
	Memberships []Membership
	History     []Transition
	Commands    []AppliedCommand
	Revision    uint64
	UpdatedAt   time.Time
}

type Command struct {
	ID               string
	ExpectedRevision uint64
	ActorID          string
	Kind             string
	Payload          string
	RecordedAt       time.Time
}

func Create(id string, kind Type, ownerID string, command Command) (Circle, error) {
	id, ownerID = strings.TrimSpace(id), strings.TrimSpace(ownerID)
	if !opaquePattern.MatchString(id) || !opaquePattern.MatchString(ownerID) ||
		!validType(kind) || command.ExpectedRevision != 0 {
		return Circle{}, ErrInvalidCircle
	}
	circle := Circle{
		id: id, kind: kind, visibility: VisibilityPrivate,
		memberships: make(map[string]Membership),
	}
	return circle.transition(ownerID, StateOwner, command, true)
}

func Rehydrate(state State) (Circle, error) {
	state.ID = strings.TrimSpace(state.ID)
	if !opaquePattern.MatchString(state.ID) || !validType(state.Type) || !validVisibility(state.Visibility) ||
		state.Revision == 0 || state.UpdatedAt.IsZero() ||
		uint64(len(state.History)) != state.Revision || uint64(len(state.Commands)) != state.Revision {
		return Circle{}, ErrInvalidCircle
	}
	circle := Circle{
		id: state.ID, kind: state.Type, visibility: state.Visibility,
		memberships: make(map[string]Membership, len(state.Memberships)),
		history:     append([]Transition(nil), state.History...),
		commands:    append([]AppliedCommand(nil), state.Commands...),
		revision:    state.Revision, updatedAt: state.UpdatedAt.UTC(),
	}
	owners := 0
	for _, membership := range state.Memberships {
		if _, duplicate := circle.memberships[membership.memberID]; duplicate {
			return Circle{}, ErrInvalidCircle
		}
		circle.memberships[membership.memberID] = membership
		if membership.state == StateOwner {
			owners++
		}
	}
	if owners != 1 {
		return Circle{}, ErrInvalidCircle
	}
	seenCommands := map[string]struct{}{}
	replayed := make(map[string]MembershipState, len(state.Memberships))
	for index := range state.History {
		event, command := state.History[index], state.Commands[index]
		if event.revision != uint64(index+1) || command.revision != uint64(index+1) ||
			event.commandID != command.id || event.at.After(state.UpdatedAt) {
			return Circle{}, ErrInvalidCircle
		}
		if _, duplicate := seenCommands[command.id]; duplicate {
			return Circle{}, ErrInvalidCircle
		}
		seenCommands[command.id] = struct{}{}
		current, exists := replayed[event.memberID]
		expectedFrom := ""
		if exists {
			expectedFrom = string(current)
		}
		ownerSetting := exists && current == StateOwner && event.to == StateOwner
		if event.from != expectedFrom ||
			(!ownerSetting && !allowedTransition(current, exists, event.to, index == 0)) {
			return Circle{}, ErrInvalidCircle
		}
		if index == 0 {
			if event.to != StateOwner || event.actorID != event.memberID {
				return Circle{}, ErrInvalidCircle
			}
		} else if event.to == StateRequested {
			if event.actorID != event.memberID {
				return Circle{}, ErrInvalidCircle
			}
		} else if event.to == StateLeft {
			if event.actorID != event.memberID {
				return Circle{}, ErrInvalidCircle
			}
		} else {
			actorState := replayed[event.actorID]
			if actorState != StateOwner && actorState != StateHost {
				return Circle{}, ErrInvalidCircle
			}
		}
		replayed[event.memberID] = event.to
	}
	if len(replayed) != len(circle.memberships) {
		return Circle{}, ErrInvalidCircle
	}
	for memberID, membership := range circle.memberships {
		if replayed[memberID] != membership.state {
			return Circle{}, ErrInvalidCircle
		}
	}
	return circle, nil
}

func (circle Circle) Request(memberID string, command Command) (Circle, error) {
	return circle.transition(memberID, StateRequested, command, false)
}

func (circle Circle) Approve(memberID string, command Command) (Circle, error) {
	return circle.transition(memberID, StateMember, command, false)
}

func (circle Circle) PromoteHost(memberID string, command Command) (Circle, error) {
	return circle.transition(memberID, StateHost, command, false)
}

func (circle Circle) Leave(memberID string, command Command) (Circle, error) {
	return circle.transition(memberID, StateLeft, command, false)
}

func (circle Circle) Expel(memberID string, command Command) (Circle, error) {
	return circle.transition(memberID, StateExpelled, command, false)
}

func (circle Circle) transition(memberID string, target MembershipState, command Command, creating bool) (Circle, error) {
	memberID = strings.TrimSpace(memberID)
	command.ID, command.ActorID, command.Kind, command.Payload = strings.TrimSpace(command.ID), strings.TrimSpace(command.ActorID), strings.TrimSpace(command.Kind), strings.TrimSpace(command.Payload)
	fingerprint := commandFingerprint(circle.id, command)
	for _, applied := range circle.commands {
		if applied.id != command.ID {
			continue
		}
		if applied.fingerprint != fingerprint {
			return Circle{}, ErrCommandMismatch
		}
		return circle, nil
	}
	if !opaquePattern.MatchString(memberID) || !opaquePattern.MatchString(command.ID) ||
		!opaquePattern.MatchString(command.ActorID) || !opaquePattern.MatchString(command.Kind) ||
		command.Payload == "" || command.RecordedAt.IsZero() ||
		(!circle.updatedAt.IsZero() && command.RecordedAt.UTC().Before(circle.updatedAt)) {
		return Circle{}, ErrInvalidCircle
	}
	if command.ExpectedRevision != circle.revision {
		return Circle{}, ErrStaleRevision
	}
	current, exists := circle.memberships[memberID]
	from := ""
	if exists {
		from = string(current.state)
	}
	if !allowedTransition(current.state, exists, target, creating) {
		return Circle{}, ErrInvalidTransition
	}
	if creating && command.ActorID != memberID {
		return Circle{}, ErrAccessDenied
	}
	if !creating && target != StateRequested && !circle.canManage(command.ActorID) && command.ActorID != memberID {
		return Circle{}, ErrAccessDenied
	}
	if target == StateRequested && command.ActorID != memberID {
		return Circle{}, ErrAccessDenied
	}
	next := circle.clone()
	revision := circle.revision + 1
	membership, _ := NewMembership(memberID, target, command.RecordedAt)
	event, _ := NewTransition(revision, command.ID, command.ActorID, memberID, from, target, command.RecordedAt)
	applied, _ := NewAppliedCommand(command.ID, fingerprint, revision)
	next.memberships[memberID] = membership
	next.history = append(next.history, event)
	next.commands = append(next.commands, applied)
	next.revision, next.updatedAt = revision, command.RecordedAt.UTC()
	return next, nil
}

func (circle Circle) SetVisibility(visibility Visibility, command Command) (Circle, error) {
	command.ID, command.ActorID, command.Kind = strings.TrimSpace(command.ID), strings.TrimSpace(command.ActorID), strings.TrimSpace(command.Kind)
	command.Payload = string(visibility)
	fingerprint := commandFingerprint(circle.id, command)
	for _, applied := range circle.commands {
		if applied.id != command.ID {
			continue
		}
		if applied.fingerprint != fingerprint {
			return Circle{}, ErrCommandMismatch
		}
		return circle, nil
	}
	if !validVisibility(visibility) || !circle.canManage(command.ActorID) {
		return Circle{}, ErrAccessDenied
	}
	if !opaquePattern.MatchString(command.ID) || !opaquePattern.MatchString(command.Kind) ||
		command.RecordedAt.IsZero() || command.ExpectedRevision != circle.revision ||
		command.RecordedAt.UTC().Before(circle.updatedAt) {
		if command.ExpectedRevision != circle.revision {
			return Circle{}, ErrStaleRevision
		}
		return Circle{}, ErrInvalidCircle
	}
	// Settings changes are audited as an owner self-transition so every
	// revision retains one immutable event and command pair.
	next := circle.clone()
	revision := circle.revision + 1
	owner := circle.ownerID()
	event, _ := NewTransition(revision, command.ID, command.ActorID, owner, string(StateOwner), StateOwner, command.RecordedAt)
	applied, _ := NewAppliedCommand(command.ID, fingerprint, revision)
	next.history, next.commands = append(next.history, event), append(next.commands, applied)
	next.revision, next.updatedAt, next.visibility = revision, command.RecordedAt.UTC(), visibility
	return next, nil
}

func (circle Circle) Allows(memberID string, capability Capability) bool {
	memberID = strings.TrimSpace(memberID)
	membership, exists := circle.memberships[memberID]
	active := exists && (membership.state == StateMember || membership.state == StateHost || membership.state == StateOwner)
	switch capability {
	case CapabilityDiscover:
		return circle.visibility == VisibilityDiscoverable || active
	case CapabilityView, CapabilityPost:
		return active
	case CapabilityManage:
		return exists && (membership.state == StateHost || membership.state == StateOwner)
	default:
		return false
	}
}

func (circle Circle) ID() string             { return circle.id }
func (circle Circle) Type() Type             { return circle.kind }
func (circle Circle) Visibility() Visibility { return circle.visibility }
func (circle Circle) Revision() uint64       { return circle.revision }
func (circle Circle) UpdatedAt() time.Time   { return circle.updatedAt }
func (circle Circle) History() []Transition  { return append([]Transition(nil), circle.history...) }
func (circle Circle) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), circle.commands...)
}
func (circle Circle) Memberships() []Membership {
	result := make([]Membership, 0, len(circle.memberships))
	for _, membership := range circle.memberships {
		result = append(result, membership)
	}
	return result
}
func (circle Circle) HasCommand(id string) bool {
	for _, command := range circle.commands {
		if command.id == strings.TrimSpace(id) {
			return true
		}
	}
	return false
}

func (circle Circle) VerifyReplay(command Command) error {
	for _, applied := range circle.commands {
		if applied.id != strings.TrimSpace(command.ID) {
			continue
		}
		if applied.fingerprint != commandFingerprint(circle.id, command) {
			return ErrCommandMismatch
		}
		return nil
	}
	return ErrCommandMismatch
}

func (circle Circle) clone() Circle {
	next := circle
	next.memberships = make(map[string]Membership, len(circle.memberships))
	for id, membership := range circle.memberships {
		next.memberships[id] = membership
	}
	next.history = append([]Transition(nil), circle.history...)
	next.commands = append([]AppliedCommand(nil), circle.commands...)
	return next
}

func (circle Circle) canManage(memberID string) bool {
	return circle.Allows(memberID, CapabilityManage)
}

func (circle Circle) ownerID() string {
	for id, membership := range circle.memberships {
		if membership.state == StateOwner {
			return id
		}
	}
	return ""
}

func validType(kind Type) bool {
	switch kind {
	case TypeCommunity, TypeCampus, TypeProfessional, TypeInterest, TypeSupport:
		return true
	default:
		return false
	}
}

func validVisibility(visibility Visibility) bool {
	return visibility == VisibilityPrivate || visibility == VisibilityDiscoverable
}

func validState(state MembershipState) bool {
	switch state {
	case StateRequested, StateMember, StateHost, StateOwner, StateExpelled, StateLeft:
		return true
	default:
		return false
	}
}

func allowedTransition(current MembershipState, exists bool, target MembershipState, creating bool) bool {
	if creating {
		return !exists && target == StateOwner
	}
	if !exists {
		return target == StateRequested
	}
	switch current {
	case StateRequested:
		return target == StateMember || target == StateExpelled
	case StateMember:
		return target == StateHost || target == StateLeft || target == StateExpelled
	case StateHost:
		return target == StateLeft || target == StateExpelled
	default:
		return false
	}
}

func commandFingerprint(circleID string, command Command) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		circleID, command.ActorID, command.Kind, command.Payload,
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}
