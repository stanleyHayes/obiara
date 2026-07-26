package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"time"
)

type Type string
type Status string
type Action string

const (
	TypeCall        Type   = "call"
	TypeMeeting     Type   = "meeting"
	TypeExclusivity Type   = "exclusivity"
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
	StatusWithdrawn Status = "withdrawn"
	ActionCreated   Action = "created"
	ActionAccepted  Action = "accepted"
	ActionRejected  Action = "rejected"
	ActionWithdrawn Action = "withdrawn"
	MaxLifetime            = 30 * 24 * time.Hour
)

var (
	ErrInvalid         = errors.New("invalid private proposal")
	ErrNotParticipant  = errors.New("proposal unavailable")
	ErrTerminal        = errors.New("proposal already decided")
	ErrExpired         = errors.New("proposal expired")
	ErrStaleRevision   = errors.New("stale proposal revision")
	ErrCommandMismatch = errors.New("proposal command replay mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Command struct {
	ID, ActorKey, Fingerprint string
	ExpectedRevision          uint64
	At                        time.Time
}
type Event struct {
	Sequence                         uint64
	Action                           Action
	CommandID, ActorKey, Fingerprint string
	At                               time.Time
}
type State struct {
	ID                                 string
	Kind                               Type
	SenderKey, RecipientKey, DetailKey string
	Status                             Status
	ExpiresAt                          time.Time
	Revision                           uint64
	Events                             []Event
}
type Proposal struct{ state State }

func Create(id string, kind Type, senderKey, recipientKey, detailKey string, expiresAt time.Time, c Command) (Proposal, error) {
	if !idPattern.MatchString(id) || !validType(kind) || !keyPattern.MatchString(senderKey) || !keyPattern.MatchString(recipientKey) || senderKey == recipientKey || !keyPattern.MatchString(detailKey) || c.ActorKey != senderKey || c.ExpectedRevision != 0 || !expiresAt.After(c.At) || expiresAt.After(c.At.Add(MaxLifetime)) {
		return Proposal{}, ErrInvalid
	}
	p := Proposal{State{id, kind, senderKey, recipientKey, detailKey, StatusPending, expiresAt.UTC(), 0, nil}}
	return p.apply(ActionCreated, c)
}
func Rehydrate(s State) (Proposal, error) {
	if !idPattern.MatchString(s.ID) || !validType(s.Kind) || !keyPattern.MatchString(s.SenderKey) || !keyPattern.MatchString(s.RecipientKey) || s.SenderKey == s.RecipientKey || !keyPattern.MatchString(s.DetailKey) || s.Revision == 0 || len(s.Events) != int(s.Revision) {
		return Proposal{}, ErrInvalid
	}
	p := Proposal{s}
	p.state.Events = append([]Event(nil), s.Events...)
	status := Status("")
	for i, e := range p.state.Events {
		if e.Sequence != uint64(i+1) || !idPattern.MatchString(e.CommandID) || !keyPattern.MatchString(e.ActorKey) || !keyPattern.MatchString(e.Fingerprint) {
			return Proposal{}, ErrInvalid
		}
		switch e.Action {
		case ActionCreated:
			if i != 0 || e.ActorKey != s.SenderKey {
				return Proposal{}, ErrInvalid
			}
			status = StatusPending
		case ActionAccepted:
			if status != StatusPending || e.ActorKey != s.RecipientKey {
				return Proposal{}, ErrInvalid
			}
			status = StatusAccepted
		case ActionRejected:
			if status != StatusPending || e.ActorKey != s.RecipientKey {
				return Proposal{}, ErrInvalid
			}
			status = StatusRejected
		case ActionWithdrawn:
			if status != StatusPending || e.ActorKey != s.SenderKey {
				return Proposal{}, ErrInvalid
			}
			status = StatusWithdrawn
		default:
			return Proposal{}, ErrInvalid
		}
	}
	if status != s.Status {
		return Proposal{}, ErrInvalid
	}
	p.state.ExpiresAt = s.ExpiresAt.UTC()
	return p, nil
}
func (p Proposal) Accept(c Command) (Proposal, error) {
	return p.decide(ActionAccepted, c, p.state.RecipientKey)
}
func (p Proposal) Reject(c Command) (Proposal, error) {
	return p.decide(ActionRejected, c, p.state.RecipientKey)
}
func (p Proposal) Withdraw(c Command) (Proposal, error) {
	return p.decide(ActionWithdrawn, c, p.state.SenderKey)
}
func (p Proposal) decide(action Action, c Command, required string) (Proposal, error) {
	for _, e := range p.state.Events {
		if e.CommandID == c.ID {
			if e.Fingerprint != c.Fingerprint {
				return p, ErrCommandMismatch
			}
			return p, nil
		}
	}
	if c.ActorKey != required {
		return p, ErrNotParticipant
	}
	if p.state.Status != StatusPending {
		return p, ErrTerminal
	}
	if !c.At.Before(p.state.ExpiresAt) {
		return p, ErrExpired
	}
	return p.apply(action, c)
}
func (p Proposal) apply(action Action, c Command) (Proposal, error) {
	if !idPattern.MatchString(c.ID) || !keyPattern.MatchString(c.ActorKey) || !keyPattern.MatchString(c.Fingerprint) || c.At.IsZero() {
		return Proposal{}, ErrInvalid
	}
	if c.ExpectedRevision != p.state.Revision {
		return Proposal{}, ErrStaleRevision
	}
	p.state.Revision++
	p.state.Events = append(append([]Event(nil), p.state.Events...), Event{p.state.Revision, action, c.ID, c.ActorKey, c.Fingerprint, c.At.UTC()})
	switch action {
	case ActionCreated:
		p.state.Status = StatusPending
	case ActionAccepted:
		p.state.Status = StatusAccepted
	case ActionRejected:
		p.state.Status = StatusRejected
	case ActionWithdrawn:
		p.state.Status = StatusWithdrawn
	}
	return p, nil
}
func Fingerprint(id string, action Action, actor string, revision uint64) string {
	s := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + actor + "\x00" + strconv.FormatUint(revision, 10)))
	return hex.EncodeToString(s[:])
}
func validType(v Type) bool         { return v == TypeCall || v == TypeMeeting || v == TypeExclusivity }
func (p Proposal) State() State     { s := p.state; s.Events = append([]Event(nil), s.Events...); return s }
func (p Proposal) ID() string       { return p.state.ID }
func (p Proposal) Revision() uint64 { return p.state.Revision }
func (p Proposal) Status() Status   { return p.state.Status }
