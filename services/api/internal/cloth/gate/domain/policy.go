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

const VersionV1 = "gate.v1"

type Action string

const (
	ActionOpened  Action = "opened"
	ActionGranted Action = "granted"
	ActionRevoked Action = "revoked"
)

var (
	ErrInvalid         = errors.New("invalid cloth gate policy")
	ErrUnknownVersion  = errors.New("unknown cloth gate version")
	ErrNotMember       = errors.New("gate actor unavailable")
	ErrStaleRevision   = errors.New("stale gate revision")
	ErrCommandMismatch = errors.New("gate command replay mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Capability struct{ ReviewerKey, QuestionKey, MaterialKey string }
type MemberGrant struct {
	MemberKey  string
	Capability Capability
}
type Command struct {
	ID, ActorKey, Fingerprint string
	ExpectedRevision          uint64
	At                        time.Time
}
type Event struct {
	Sequence                         uint64
	Action                           Action
	CommandID, ActorKey, Fingerprint string
	Capability                       Capability
	At                               time.Time
}
type State struct {
	ID, Version string
	Members     [2]string
	Grants      []MemberGrant
	Revision    uint64
	Events      []Event
}
type Policy struct{ state State }

func Open(id, version string, members [2]string, c Command) (Policy, error) {
	if version != VersionV1 {
		return Policy{}, ErrUnknownVersion
	}
	if !idPattern.MatchString(id) || !validMembers(members) || c.ExpectedRevision != 0 || (c.ActorKey != members[0] && c.ActorKey != members[1]) {
		return Policy{}, ErrInvalid
	}
	if members[1] < members[0] {
		members[0], members[1] = members[1], members[0]
	}
	p := Policy{State{ID: id, Version: version, Members: members}}
	return p.apply(ActionOpened, Capability{}, c)
}
func Rehydrate(s State) (Policy, error) {
	if s.Version != VersionV1 {
		return Policy{}, ErrUnknownVersion
	}
	if !idPattern.MatchString(s.ID) || !validMembers(s.Members) || s.Members[0] >= s.Members[1] || s.Revision == 0 || len(s.Events) != int(s.Revision) {
		return Policy{}, ErrInvalid
	}
	p := Policy{state: s}
	p.state.Grants = append([]MemberGrant(nil), s.Grants...)
	p.state.Events = append([]Event(nil), s.Events...)
	computed := []MemberGrant{}
	for i, event := range p.state.Events {
		if event.Sequence != uint64(i+1) || !idPattern.MatchString(event.CommandID) || !keyPattern.MatchString(event.ActorKey) || !keyPattern.MatchString(event.Fingerprint) || (event.ActorKey != s.Members[0] && event.ActorKey != s.Members[1]) {
			return Policy{}, ErrInvalid
		}
		switch event.Action {
		case ActionOpened:
			if i != 0 {
				return Policy{}, ErrInvalid
			}
		case ActionGranted:
			if !validCapability(event.Capability) {
				return Policy{}, ErrInvalid
			}
			computed = setGrant(computed, event.ActorKey, event.Capability, true)
		case ActionRevoked:
			if !validCapability(event.Capability) {
				return Policy{}, ErrInvalid
			}
			computed = setGrant(computed, event.ActorKey, event.Capability, false)
		default:
			return Policy{}, ErrInvalid
		}
	}
	sortGrants(computed)
	stored := append([]MemberGrant(nil), s.Grants...)
	sortGrants(stored)
	if !slices.Equal(computed, stored) {
		return Policy{}, ErrInvalid
	}
	p.state.Grants = stored
	return p, nil
}
func (p Policy) Grant(cap Capability, c Command) (Policy, error) {
	return p.change(ActionGranted, cap, c)
}
func (p Policy) Revoke(cap Capability, c Command) (Policy, error) {
	return p.change(ActionRevoked, cap, c)
}
func (p Policy) change(action Action, cap Capability, c Command) (Policy, error) {
	for _, event := range p.state.Events {
		if event.CommandID == c.ID {
			if event.Fingerprint != c.Fingerprint {
				return p, ErrCommandMismatch
			}
			return p, nil
		}
	}
	if c.ActorKey != p.state.Members[0] && c.ActorKey != p.state.Members[1] {
		return p, ErrNotMember
	}
	if !validCapability(cap) {
		return p, ErrInvalid
	}
	return p.apply(action, cap, c)
}
func (p Policy) apply(action Action, cap Capability, c Command) (Policy, error) {
	if !idPattern.MatchString(c.ID) || !keyPattern.MatchString(c.ActorKey) || !keyPattern.MatchString(c.Fingerprint) || c.At.IsZero() {
		return Policy{}, ErrInvalid
	}
	if c.ExpectedRevision != p.state.Revision {
		return Policy{}, ErrStaleRevision
	}
	p.state.Grants = append([]MemberGrant(nil), p.state.Grants...)
	if action == ActionGranted {
		p.state.Grants = setGrant(p.state.Grants, c.ActorKey, cap, true)
	} else if action == ActionRevoked {
		p.state.Grants = setGrant(p.state.Grants, c.ActorKey, cap, false)
	}
	sortGrants(p.state.Grants)
	p.state.Revision++
	p.state.Events = append(append([]Event(nil), p.state.Events...), Event{p.state.Revision, action, c.ID, c.ActorKey, c.Fingerprint, cap, c.At.UTC()})
	return p, nil
}
func (p Policy) Allows(cap Capability) bool {
	if !validCapability(cap) {
		return false
	}
	return hasGrant(p.state.Grants, p.state.Members[0], cap) && hasGrant(p.state.Grants, p.state.Members[1], cap)
}
func Fingerprint(id string, action Action, actor string, cap Capability, revision uint64) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + actor + "\x00" + cap.ReviewerKey + "\x00" + cap.QuestionKey + "\x00" + cap.MaterialKey + "\x00" + strconv.FormatUint(revision, 10)))
	return hex.EncodeToString(sum[:])
}
func validMembers(m [2]string) bool {
	return keyPattern.MatchString(m[0]) && keyPattern.MatchString(m[1]) && m[0] != m[1]
}
func validCapability(c Capability) bool {
	return keyPattern.MatchString(c.ReviewerKey) && keyPattern.MatchString(c.QuestionKey) && keyPattern.MatchString(c.MaterialKey)
}
func hasGrant(gs []MemberGrant, m string, c Capability) bool {
	return slices.Contains(gs, MemberGrant{m, c})
}
func setGrant(gs []MemberGrant, m string, c Capability, on bool) []MemberGrant {
	g := MemberGrant{m, c}
	index := slices.Index(gs, g)
	if on && index < 0 {
		return append(gs, g)
	}
	if !on && index >= 0 {
		return append(gs[:index:index], gs[index+1:]...)
	}
	return gs
}
func sortGrants(gs []MemberGrant) {
	slices.SortFunc(gs, func(a, b MemberGrant) int {
		left := a.MemberKey + a.Capability.ReviewerKey + a.Capability.QuestionKey + a.Capability.MaterialKey
		right := b.MemberKey + b.Capability.ReviewerKey + b.Capability.QuestionKey + b.Capability.MaterialKey
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
		return 0
	})
}
func (p Policy) State() State {
	s := p.state
	s.Grants = append([]MemberGrant(nil), s.Grants...)
	s.Events = append([]Event(nil), s.Events...)
	return s
}
func (p Policy) ID() string       { return p.state.ID }
func (p Policy) Revision() uint64 { return p.state.Revision }
