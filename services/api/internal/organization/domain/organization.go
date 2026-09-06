// Package domain owns what an organization is.
//
// An organization is a body Obiara has a relationship with — a university, an
// employer, a church, a partner business. It is not a member and never
// becomes one: it has no voice, no tier, and no way into any member surface.
// It exists so that a discount code has an issuer to hang off and an audit
// trail has somebody to name (agent_plan.md §41).
//
// Codes are issued by Obiara staff on an organization's behalf, so there is
// deliberately no organization principal and no organization sign-in here.
// Adding one later is additive; building one now would be a new trust
// boundary to defend for nobody.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type Status string
type Action string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"

	ActionRegistered Action = "registered"
	ActionSuspended  Action = "suspended"
	ActionRestored   Action = "restored"
	ActionRenamed    Action = "renamed"
)

var (
	ErrInvalidOrganization = errors.New("invalid organization")
	ErrInvalidTransition   = errors.New("invalid organization transition")
	ErrStaleRevision       = errors.New("stale organization revision")
	ErrCommandMismatch     = errors.New("organization command replay mismatch")
)

var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	// actorPattern is a keyed operator. Who acted is recorded as a one-way
	// digest for the same reason it is everywhere else: an audit trail has to
	// prove that somebody acted without being a directory of who.
	actorPattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

// MaxNameLength is generous. An organization's name is whatever it calls
// itself, and this is a bound against abuse rather than a style guide.
const MaxNameLength = 200

type Command struct {
	ID, ActorKey, ReasonCode string
	ExpectedRevision         uint64
	At                       time.Time
}

type Event struct {
	Sequence                        uint64
	CommandID, ActorKey, ReasonCode string
	Action                          Action
	At                              time.Time
}

type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}

// Organization is the aggregate.
//
// BillingEmail is stored as written rather than keyed: it is how somebody is
// invoiced, so it has to be readable by the person doing the invoicing. It is
// an organization's address, not a member's, which is why the rule differs
// from everywhere else in this codebase.
type Organization struct {
	id, name, billingEmail string
	status                 Status
	revision               uint64
	registeredAt           time.Time
	events                 []Event
	commands               []AppliedCommand
}

type State struct {
	ID, Name, BillingEmail string
	Status                 Status
	Revision               uint64
	RegisteredAt           time.Time
	Events                 []Event
	Commands               []AppliedCommand
}

// Register records an organization for the first time.
func Register(id, name, billingEmail string, command Command) (Organization, error) {
	name = strings.TrimSpace(name)
	billingEmail = strings.ToLower(strings.TrimSpace(billingEmail))
	if !opaquePattern.MatchString(strings.TrimSpace(id)) ||
		name == "" || len([]rune(name)) > MaxNameLength ||
		!validEmail(billingEmail) || !command.valid() {
		return Organization{}, ErrInvalidOrganization
	}
	organization := Organization{
		id: strings.TrimSpace(id), name: name, billingEmail: billingEmail,
		status: StatusActive, registeredAt: command.At.UTC(),
	}
	return organization.apply(command, ActionRegistered, StatusActive)
}

// Suspend stops an organization being used as an issuer.
//
// It does not touch codes already issued. Whether a suspended issuer's codes
// still redeem is the promotion context's rule, not this one's: an
// organization that stops paying its invoice is a different thing from a code
// that should stop working, and conflating them here would decide both at
// once.
func (organization Organization) Suspend(command Command) (Organization, error) {
	return organization.change(command, ActionSuspended, StatusActive, StatusSuspended)
}

// Restore returns a suspended organization to use.
func (organization Organization) Restore(command Command) (Organization, error) {
	return organization.change(command, ActionRestored, StatusSuspended, StatusActive)
}

// Rename records a change of name, which organizations do.
//
// Kept as its own action rather than a silent field update, because the audit
// trail is the point of this aggregate: a code issued last year was issued by
// whatever the organization was called then, and the trail has to say so.
func (organization Organization) Rename(name string, command Command) (Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > MaxNameLength {
		return Organization{}, ErrInvalidOrganization
	}
	renamed, err := organization.change(command, ActionRenamed, StatusActive, StatusActive)
	if err != nil {
		return Organization{}, err
	}
	if renamed.revision == organization.revision {
		// A replay. The rename already happened and the name it set is
		// already on the aggregate; setting it again from this command's
		// argument would let a retry carrying a different name rewrite it.
		return renamed, nil
	}
	renamed.name = name
	return renamed, nil
}

// change applies one transition, and is the only place that does.
//
// The replay check comes first, before the state guard, because a retry of the
// command that produced the current state must be answered with that state.
// The other order reports "already suspended" to an operator whose first
// request succeeded and whose connection dropped — so they suspend again to
// find out, which is exactly what idempotency exists to make unnecessary.
func (organization Organization) change(
	command Command, action Action, from, to Status,
) (Organization, error) {
	if !command.valid() {
		return Organization{}, ErrInvalidOrganization
	}
	if applied, found := organization.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, action, to) {
			return Organization{}, ErrCommandMismatch
		}
		return organization, nil
	}
	if organization.status != from {
		return Organization{}, ErrInvalidTransition
	}
	if command.ExpectedRevision != organization.revision {
		return Organization{}, ErrStaleRevision
	}
	return organization.apply(command, action, to)
}

func (organization Organization) apply(
	command Command, action Action, status Status,
) (Organization, error) {
	next := organization
	next.status = status
	next.revision = organization.revision + 1
	next.events = append(append([]Event(nil), organization.events...), Event{
		Sequence: next.revision, CommandID: command.ID, ActorKey: command.ActorKey,
		ReasonCode: command.ReasonCode, Action: action, At: command.At.UTC(),
	})
	next.commands = append(append([]AppliedCommand(nil), organization.commands...), AppliedCommand{
		ID: command.ID, Fingerprint: fingerprint(command, action, status), Revision: next.revision,
	})
	return next, nil
}

func (organization Organization) applied(commandID string) (AppliedCommand, bool) {
	for _, applied := range organization.commands {
		if applied.ID == commandID {
			return applied, true
		}
	}
	return AppliedCommand{}, false
}

func (command Command) valid() bool {
	return opaquePattern.MatchString(strings.TrimSpace(command.ID)) &&
		actorPattern.MatchString(command.ActorKey) &&
		reasonPattern.MatchString(command.ReasonCode) &&
		!command.At.IsZero()
}

// fingerprint binds a command id to what it did, so a retry carrying the same
// id and different intent is refused rather than answered with the first one.
func fingerprint(command Command, action Action, status Status) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		command.ID, command.ActorKey, command.ReasonCode, string(action), string(status),
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}

// validEmail is deliberately loose. This is a billing contact somebody typed,
// and a stricter rule here would reject real addresses to no benefit — what
// makes it correct is that an invoice reaches somebody, which no regular
// expression can establish.
func validEmail(address string) bool {
	if address == "" || len(address) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(address)
	return err == nil && parsed.Address == address
}

// Rehydrate rebuilds an organization from storage.
func Rehydrate(state State) (Organization, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(state.ID)) ||
		state.Revision == 0 ||
		uint64(len(state.Events)) != state.Revision ||
		uint64(len(state.Commands)) != state.Revision {
		return Organization{}, ErrInvalidOrganization
	}
	if state.Status != StatusActive && state.Status != StatusSuspended {
		return Organization{}, ErrInvalidOrganization
	}
	return Organization{
		id: strings.TrimSpace(state.ID), name: state.Name, billingEmail: state.BillingEmail,
		status: state.Status, revision: state.Revision, registeredAt: state.RegisteredAt.UTC(),
		events:   append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}, nil
}

func (organization Organization) ID() string           { return organization.id }
func (organization Organization) Name() string         { return organization.name }
func (organization Organization) BillingEmail() string { return organization.billingEmail }
func (organization Organization) Status() Status       { return organization.status }
func (organization Organization) Revision() uint64     { return organization.revision }
func (organization Organization) RegisteredAt() time.Time {
	return organization.registeredAt
}
func (organization Organization) Events() []Event {
	return append([]Event(nil), organization.events...)
}
func (organization Organization) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), organization.commands...)
}

// Issuing reports whether this organization may have a code issued for it.
//
// A suspended organization is not a live relationship, so nothing new is
// issued in its name.
func (organization Organization) Issuing() bool { return organization.status == StatusActive }
