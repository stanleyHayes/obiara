// Package domain models immutable, consented vouch attestations and a bounded
// non-transferable reputation commitment.
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

const (
	StatusAwaitingConsent Status = "awaiting_consent"
	StatusActive          Status = "active"
	StatusExpired         Status = "expired"
	StatusRevoked         Status = "revoked"

	ActionProposed  Action = "proposed"
	ActionConsented Action = "consented"
	ActionExpired   Action = "expired"
	ActionRevoked   Action = "revoked"

	MaxStakeUnits uint8 = 100
)

var (
	ErrInvalidAttestation = errors.New("invalid vouch attestation")
	ErrInvalidTransition  = errors.New("invalid vouch attestation transition")
	ErrAttestationExpired = errors.New("vouch attestation expired")
	ErrStaleRevision      = errors.New("stale vouch attestation revision")
	ErrCommandMismatch    = errors.New("vouch attestation command replay mismatch")
)

var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	keyPattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type SubjectScope struct {
	Kind string
	Key  string
}

type Provenance struct {
	VoucherKey    string
	PolicyVersion string
	ConsentedAt   time.Time
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

type Attestation struct {
	id            string
	subjectKey    string
	voucherKey    string
	scope         SubjectScope
	stakeUnits    uint8
	policyVersion string
	status        Status
	expiresAt     time.Time
	provenance    *Provenance
	endedAt       *time.Time
	revision      uint64
	events        []Event
	commands      []AppliedCommand
}

type State struct {
	ID            string
	SubjectKey    string
	VoucherKey    string
	Scope         SubjectScope
	StakeUnits    uint8
	PolicyVersion string
	Status        Status
	ExpiresAt     time.Time
	Provenance    *Provenance
	EndedAt       *time.Time
	Revision      uint64
	Events        []Event
	Commands      []AppliedCommand
}

func Propose(id, subjectKey, voucherKey string, scope SubjectScope, stakeUnits uint8, policyVersion string, expiresAt time.Time, command Command) (Attestation, error) {
	if !validOpaque(id) || !validKey(subjectKey) || !validKey(voucherKey) || subjectKey == voucherKey ||
		!validScope(scope) || stakeUnits == 0 || stakeUnits > MaxStakeUnits ||
		!validOpaque(policyVersion) || expiresAt.IsZero() || !expiresAt.After(command.At) ||
		expiresAt.After(command.At.Add(365*24*time.Hour)) || command.ExpectedRevision != 0 {
		return Attestation{}, ErrInvalidAttestation
	}
	attestation := Attestation{
		id: id, subjectKey: subjectKey, voucherKey: voucherKey, scope: scope,
		stakeUnits: stakeUnits, policyVersion: policyVersion, expiresAt: expiresAt.UTC(),
	}
	return attestation.transition(ActionProposed, command)
}

func Rehydrate(state State) (Attestation, error) {
	attestation := Attestation{
		id: state.ID, subjectKey: state.SubjectKey, voucherKey: state.VoucherKey,
		scope: state.Scope, stakeUnits: state.StakeUnits, policyVersion: state.PolicyVersion,
		status: state.Status, expiresAt: state.ExpiresAt.UTC(),
		provenance: cloneProvenance(state.Provenance), endedAt: cloneTime(state.EndedAt),
		revision: state.Revision, events: append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if !validOpaque(attestation.id) || !validKey(attestation.subjectKey) || !validKey(attestation.voucherKey) ||
		!validScope(attestation.scope) || attestation.stakeUnits == 0 || attestation.stakeUnits > MaxStakeUnits ||
		!validOpaque(attestation.policyVersion) || attestation.revision == 0 ||
		uint64(len(attestation.events)) != attestation.revision ||
		uint64(len(attestation.commands)) != attestation.revision {
		return Attestation{}, ErrInvalidAttestation
	}
	status := Status("")
	var provenance *Provenance
	var endedAt *time.Time
	seen := make(map[string]struct{}, len(attestation.commands))
	var previousAt time.Time
	for index, event := range attestation.events {
		applied := attestation.commands[index]
		if event.Sequence != uint64(index+1) || applied.Revision != event.Sequence ||
			applied.ID != event.CommandID || !validEvent(event) ||
			(!previousAt.IsZero() && event.At.Before(previousAt)) {
			return Attestation{}, ErrInvalidAttestation
		}
		if _, duplicate := seen[applied.ID]; duplicate {
			return Attestation{}, ErrInvalidAttestation
		}
		seen[applied.ID] = struct{}{}
		command := Command{
			ID: event.CommandID, ActorKey: event.ActorKey, ExpectedRevision: uint64(index),
			ReasonCode: event.ReasonCode, At: event.At,
		}
		if applied.Fingerprint != fingerprint(attestation.id, event.Action, command) {
			return Attestation{}, ErrInvalidAttestation
		}
		switch event.Action {
		case ActionProposed:
			if index != 0 {
				return Attestation{}, ErrInvalidAttestation
			}
			status = StatusAwaitingConsent
		case ActionConsented:
			if status != StatusAwaitingConsent || event.ActorKey != attestation.voucherKey {
				return Attestation{}, ErrInvalidAttestation
			}
			status = StatusActive
			provenance = &Provenance{
				VoucherKey: attestation.voucherKey, PolicyVersion: attestation.policyVersion,
				ConsentedAt: event.At,
			}
		case ActionExpired:
			if status != StatusAwaitingConsent && status != StatusActive {
				return Attestation{}, ErrInvalidAttestation
			}
			status = StatusExpired
			value := event.At
			endedAt = &value
		case ActionRevoked:
			if status != StatusActive {
				return Attestation{}, ErrInvalidAttestation
			}
			status = StatusRevoked
			value := event.At
			endedAt = &value
		default:
			return Attestation{}, ErrInvalidAttestation
		}
		previousAt = event.At
	}
	if status != attestation.status || !equalProvenance(provenance, attestation.provenance) ||
		!equalTime(endedAt, attestation.endedAt) {
		return Attestation{}, ErrInvalidAttestation
	}
	return attestation, nil
}

func (attestation Attestation) Consent(command Command) (Attestation, error) {
	if command.ActorKey != attestation.voucherKey {
		return Attestation{}, ErrInvalidTransition
	}
	if !command.At.Before(attestation.expiresAt) {
		return Attestation{}, ErrAttestationExpired
	}
	return attestation.change(ActionConsented, command, StatusAwaitingConsent)
}

func (attestation Attestation) Revoke(command Command) (Attestation, error) {
	return attestation.change(ActionRevoked, command, StatusActive)
}

func (attestation Attestation) Expire(command Command) (Attestation, error) {
	if replayed, err := attestation.replay(command, ActionExpired); replayed || err != nil {
		return attestation, err
	}
	if command.At.Before(attestation.expiresAt) ||
		(attestation.status != StatusAwaitingConsent && attestation.status != StatusActive) {
		return Attestation{}, ErrInvalidTransition
	}
	return attestation.transition(ActionExpired, command)
}

func (attestation Attestation) change(action Action, command Command, required Status) (Attestation, error) {
	if replayed, err := attestation.replay(command, action); replayed || err != nil {
		return attestation, err
	}
	if attestation.status != required {
		return Attestation{}, ErrInvalidTransition
	}
	return attestation.transition(action, command)
}

func (attestation Attestation) transition(action Action, command Command) (Attestation, error) {
	if err := validateCommand(command, attestation.revision); err != nil {
		return Attestation{}, err
	}
	if len(attestation.events) > 0 && command.At.Before(attestation.events[len(attestation.events)-1].At) {
		return Attestation{}, ErrInvalidAttestation
	}
	attestation.revision++
	event := Event{
		Sequence: attestation.revision, CommandID: command.ID, ActorKey: command.ActorKey,
		Action: action, ReasonCode: command.ReasonCode, At: command.At.UTC(),
	}
	attestation.events = append(attestation.events, event)
	attestation.commands = append(attestation.commands, AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(attestation.id, action, command),
		Revision: attestation.revision,
	})
	switch action {
	case ActionProposed:
		attestation.status = StatusAwaitingConsent
	case ActionConsented:
		attestation.status = StatusActive
		attestation.provenance = &Provenance{
			VoucherKey: attestation.voucherKey, PolicyVersion: attestation.policyVersion,
			ConsentedAt: event.At,
		}
	case ActionExpired:
		attestation.status = StatusExpired
		value := event.At
		attestation.endedAt = &value
	case ActionRevoked:
		attestation.status = StatusRevoked
		value := event.At
		attestation.endedAt = &value
	}
	return attestation, nil
}

func (attestation Attestation) replay(command Command, action Action) (bool, error) {
	expected := fingerprint(attestation.id, action, command)
	for _, applied := range attestation.commands {
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
		return ErrInvalidAttestation
	}
	if command.ExpectedRevision != revision {
		return ErrStaleRevision
	}
	return nil
}

func fingerprint(id string, action Action, command Command) string {
	value := id + "\x00" + string(action) + "\x00" + command.ActorKey + "\x00" +
		command.ReasonCode + "\x00" + strconv.FormatUint(command.ExpectedRevision, 10)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validOpaque(value string) bool { return opaquePattern.MatchString(strings.TrimSpace(value)) }
func validKey(value string) bool    { return keyPattern.MatchString(value) }
func validScope(scope SubjectScope) bool {
	switch scope.Kind {
	case "circle", "introduction", "sowing":
		return validKey(scope.Key)
	default:
		return false
	}
}
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
func cloneProvenance(value *Provenance) *Provenance {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.ConsentedAt = cloned.ConsentedAt.UTC()
	return &cloned
}
func equalTime(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}
func equalProvenance(left, right *Provenance) bool {
	return left == nil && right == nil || left != nil && right != nil &&
		left.VoucherKey == right.VoucherKey && left.PolicyVersion == right.PolicyVersion &&
		left.ConsentedAt.Equal(right.ConsentedAt)
}

func (a Attestation) ID() string                 { return a.id }
func (a Attestation) SubjectKey() string         { return a.subjectKey }
func (a Attestation) VoucherKey() string         { return a.voucherKey }
func (a Attestation) Scope() SubjectScope        { return a.scope }
func (a Attestation) StakeUnits() uint8          { return a.stakeUnits }
func (a Attestation) PolicyVersion() string      { return a.policyVersion }
func (a Attestation) Status() Status             { return a.status }
func (a Attestation) ExpiresAt() time.Time       { return a.expiresAt }
func (a Attestation) Provenance() *Provenance    { return cloneProvenance(a.provenance) }
func (a Attestation) EndedAt() *time.Time        { return cloneTime(a.endedAt) }
func (a Attestation) Revision() uint64           { return a.revision }
func (a Attestation) Events() []Event            { return append([]Event(nil), a.events...) }
func (a Attestation) Commands() []AppliedCommand { return append([]AppliedCommand(nil), a.commands...) }
func (a Attestation) HasCommand(id string) bool {
	for _, command := range a.commands {
		if command.ID == id {
			return true
		}
	}
	return false
}
