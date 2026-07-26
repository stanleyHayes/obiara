package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"time"
)

const MaxRecipients = 25

type Status string
type Action string

const (
	StatusActive  Status = "active"
	StatusRevoked Status = "revoked"
	StatusExpired Status = "expired"
	ActionCreated Action = "created"
	ActionPlayed  Action = "played"
	ActionRevoked Action = "revoked"
	ActionExpired Action = "expired"
)

var (
	ErrInvalidPod        = errors.New("invalid seed pod")
	ErrInvalidTransition = errors.New("invalid seed pod transition")
	ErrStaleRevision     = errors.New("stale seed pod revision")
	ErrCommandMismatch   = errors.New("seed pod command replay mismatch")
	ErrRecipientDenied   = errors.New("seed pod recipient denied")
)
var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)

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
type Pod struct {
	id, ownerKey, mediaKey string
	recipientKeys          []string
	status                 Status
	expiresAt              time.Time
	endedAt                *time.Time
	revision               uint64
	events                 []Event
	commands               []AppliedCommand
}
type State struct {
	ID, OwnerKey, MediaKey string
	RecipientKeys          []string
	Status                 Status
	ExpiresAt              time.Time
	EndedAt                *time.Time
	Revision               uint64
	Events                 []Event
	Commands               []AppliedCommand
}

func Create(id, ownerKey, mediaKey string, recipients []string, expiresAt time.Time, c Command) (Pod, error) {
	normalized, ok := normalizeRecipients(recipients)
	if !opaquePattern.MatchString(id) || !keyPattern.MatchString(ownerKey) || !keyPattern.MatchString(mediaKey) || !ok || len(normalized) == 0 || c.ExpectedRevision != 0 || !expiresAt.After(c.At) || expiresAt.After(c.At.Add(7*24*time.Hour)) {
		return Pod{}, ErrInvalidPod
	}
	p := Pod{id: id, ownerKey: ownerKey, mediaKey: mediaKey, recipientKeys: normalized, expiresAt: expiresAt.UTC()}
	return p.transition(ActionCreated, c)
}
func Rehydrate(s State) (Pod, error) {
	recipients, ok := normalizeRecipients(s.RecipientKeys)
	p := Pod{id: s.ID, ownerKey: s.OwnerKey, mediaKey: s.MediaKey, recipientKeys: recipients, status: s.Status, expiresAt: s.ExpiresAt.UTC(), endedAt: cloneTime(s.EndedAt), revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]AppliedCommand(nil), s.Commands...)}
	if !ok || len(recipients) == 0 || !opaquePattern.MatchString(p.id) || !keyPattern.MatchString(p.ownerKey) || !keyPattern.MatchString(p.mediaKey) || p.revision == 0 || len(p.events) != int(p.revision) || len(p.commands) != int(p.revision) {
		return Pod{}, ErrInvalidPod
	}
	status := Status("")
	var ended *time.Time
	seen := map[string]bool{}
	for i, e := range p.events {
		a := p.commands[i]
		if e.Sequence != uint64(i+1) || a.Revision != e.Sequence || a.ID != e.CommandID || seen[a.ID] || !validCommand(Command{ID: e.CommandID, ActorKey: e.ActorKey, ReasonCode: e.ReasonCode, ExpectedRevision: uint64(i), At: e.At}) {
			return Pod{}, ErrInvalidPod
		}
		seen[a.ID] = true
		if a.Fingerprint != fingerprint(p.id, e.Action, Command{ID: e.CommandID, ActorKey: e.ActorKey, ReasonCode: e.ReasonCode, ExpectedRevision: uint64(i), At: e.At}) {
			return Pod{}, ErrInvalidPod
		}
		switch e.Action {
		case ActionCreated:
			if i != 0 {
				return Pod{}, ErrInvalidPod
			}
			status = StatusActive
		case ActionPlayed:
			if status != StatusActive {
				return Pod{}, ErrInvalidPod
			}
		case ActionRevoked:
			if status != StatusActive {
				return Pod{}, ErrInvalidPod
			}
			status = StatusRevoked
			v := e.At
			ended = &v
		case ActionExpired:
			if status != StatusActive {
				return Pod{}, ErrInvalidPod
			}
			status = StatusExpired
			v := e.At
			ended = &v
		default:
			return Pod{}, ErrInvalidPod
		}
	}
	if status != p.status || !equalTime(ended, p.endedAt) {
		return Pod{}, ErrInvalidPod
	}
	return p, nil
}
func (p Pod) Play(c Command) (Pod, error) {
	if !p.IsRecipient(c.ActorKey) {
		return Pod{}, ErrRecipientDenied
	}
	return p.change(ActionPlayed, c, false)
}
func (p Pod) Revoke(c Command) (Pod, error) {
	if c.ActorKey != p.ownerKey {
		return Pod{}, ErrRecipientDenied
	}
	return p.change(ActionRevoked, c, false)
}
func (p Pod) Expire(c Command) (Pod, error) { return p.change(ActionExpired, c, true) }
func (p Pod) change(action Action, c Command, expiry bool) (Pod, error) {
	if replay, err := p.replay(action, c); replay || err != nil {
		return p, err
	}
	if p.status != StatusActive || expiry && c.At.Before(p.expiresAt) || action == ActionPlayed && !c.At.Before(p.expiresAt) {
		return Pod{}, ErrInvalidTransition
	}
	return p.transition(action, c)
}
func (p Pod) transition(action Action, c Command) (Pod, error) {
	if !validCommand(c) {
		return Pod{}, ErrInvalidPod
	}
	if c.ExpectedRevision != p.revision {
		return Pod{}, ErrStaleRevision
	}
	p.revision++
	e := Event{Sequence: p.revision, CommandID: c.ID, ActorKey: c.ActorKey, ReasonCode: c.ReasonCode, Action: action, At: c.At.UTC()}
	p.events = append(p.events, e)
	p.commands = append(p.commands, AppliedCommand{ID: c.ID, Fingerprint: fingerprint(p.id, action, c), Revision: p.revision})
	switch action {
	case ActionCreated:
		p.status = StatusActive
	case ActionRevoked:
		p.status = StatusRevoked
		v := e.At
		p.endedAt = &v
	case ActionExpired:
		p.status = StatusExpired
		v := e.At
		p.endedAt = &v
	}
	return p, nil
}
func (p Pod) replay(action Action, c Command) (bool, error) {
	f := fingerprint(p.id, action, c)
	for _, a := range p.commands {
		if a.ID == c.ID {
			if a.Fingerprint != f {
				return false, ErrCommandMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func normalizeRecipients(in []string) ([]string, bool) {
	if len(in) > MaxRecipients {
		return nil, false
	}
	out := append([]string(nil), in...)
	for _, v := range out {
		if !keyPattern.MatchString(v) {
			return nil, false
		}
	}
	slices.Sort(out)
	out = slices.Compact(out)
	return out, true
}
func validCommand(c Command) bool {
	return opaquePattern.MatchString(c.ID) && keyPattern.MatchString(c.ActorKey) && reasonPattern.MatchString(c.ReasonCode) && !c.At.IsZero()
}
func fingerprint(id string, a Action, c Command) string {
	s := sha256.Sum256([]byte(id + "\x00" + string(a) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.ReasonCode + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(s[:])
}
func cloneTime(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	x := v.UTC()
	return &x
}
func equalTime(a, b *time.Time) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Equal(*b)
}
func (p Pod) ID() string                 { return p.id }
func (p Pod) OwnerKey() string           { return p.ownerKey }
func (p Pod) MediaKey() string           { return p.mediaKey }
func (p Pod) RecipientKeys() []string    { return append([]string(nil), p.recipientKeys...) }
func (p Pod) IsRecipient(k string) bool  { _, ok := slices.BinarySearch(p.recipientKeys, k); return ok }
func (p Pod) Status() Status             { return p.status }
func (p Pod) ExpiresAt() time.Time       { return p.expiresAt }
func (p Pod) EndedAt() *time.Time        { return cloneTime(p.endedAt) }
func (p Pod) Revision() uint64           { return p.revision }
func (p Pod) Events() []Event            { return append([]Event(nil), p.events...) }
func (p Pod) Commands() []AppliedCommand { return append([]AppliedCommand(nil), p.commands...) }
func (p Pod) HasCommand(id string) bool {
	for _, c := range p.commands {
		if c.ID == id {
			return true
		}
	}
	return false
}
