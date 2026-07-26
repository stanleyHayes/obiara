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

type Purpose string
type Action string

const (
	PurposeReflection Purpose = "reflection"
	PurposeArchive    Purpose = "community_archive"
	PurposeMemory     Purpose = "shared_memory"
	MaxRetention              = 30 * 24 * time.Hour
	ActionOpened      Action  = "opened"
	ActionProposed    Action  = "proposed"
	ActionOptedIn     Action  = "opted_in"
	ActionRevoked     Action  = "revoked"
	ActionJoined      Action  = "joined"
	ActionLeft        Action  = "left"
	ActionStarted     Action  = "started"
	ActionStopped     Action  = "stopped"
)

var (
	ErrInvalid         = errors.New("invalid recording policy")
	ErrDenied          = errors.New("recording unavailable")
	ErrConsentRequired = errors.New("all-current consent required")
	ErrStaleRevision   = errors.New("stale recording revision")
	ErrCommandMismatch = errors.New("recording command mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Proposal struct {
	Purpose   Purpose
	Retention time.Duration
}
type Command struct {
	ID, ActorKey, Fingerprint string
	ExpectedRevision          uint64
	At                        time.Time
}
type Event struct {
	Sequence                                     uint64
	Action                                       Action
	CommandID, ActorKey, SubjectKey, Fingerprint string
	Purpose                                      Purpose
	Retention                                    time.Duration
	At                                           time.Time
}
type State struct {
	ID, FireKey, HostKey   string
	Participants, Consents []string
	Proposal               *Proposal
	Active                 bool
	RecordingRef           string
	Revision               uint64
	Events                 []Event
}
type Policy struct{ state State }

func Open(id, fireKey, hostKey string, participants []string, c Command) (Policy, error) {
	normalized, ok := set(participants)
	if !ok || len(normalized) == 0 || len(normalized) > 500 || !slices.Contains(normalized, hostKey) || !idPattern.MatchString(id) || !keyPattern.MatchString(fireKey) || c.ActorKey != hostKey || c.ExpectedRevision != 0 {
		return Policy{}, ErrInvalid
	}
	p := Policy{State{ID: id, FireKey: fireKey, HostKey: hostKey, Participants: normalized}}
	return p.apply(ActionOpened, hostKey, nil, "", c)
}
func Rehydrate(s State) (Policy, error) {
	participants, ok := set(s.Participants)
	consents, ok2 := set(s.Consents)
	if !ok || !ok2 || !idPattern.MatchString(s.ID) || !keyPattern.MatchString(s.FireKey) || !keyPattern.MatchString(s.HostKey) || !slices.Contains(participants, s.HostKey) || s.Revision == 0 || len(s.Events) != int(s.Revision) || s.Active && !keyPattern.MatchString(s.RecordingRef) {
		return Policy{}, ErrInvalid
	}
	for _, c := range consents {
		if !slices.Contains(participants, c) {
			return Policy{}, ErrInvalid
		}
	}
	s.Participants = participants
	s.Consents = consents
	s.Events = append([]Event(nil), s.Events...)
	return Policy{s}, nil
}
func (p Policy) Propose(purpose Purpose, retention time.Duration, c Command) (Policy, error) {
	if c.ActorKey != p.state.HostKey || !validPurpose(purpose) || retention <= 0 || retention > MaxRetention {
		return p, ErrDenied
	}
	proposal := &Proposal{purpose, retention}
	return p.apply(ActionProposed, c.ActorKey, proposal, "", c)
}
func (p Policy) OptIn(c Command) (Policy, error) {
	if !slices.Contains(p.state.Participants, c.ActorKey) || p.state.Proposal == nil {
		return p, ErrDenied
	}
	return p.apply(ActionOptedIn, c.ActorKey, nil, "", c)
}
func (p Policy) Revoke(c Command) (Policy, error) {
	if !slices.Contains(p.state.Participants, c.ActorKey) {
		return p, ErrDenied
	}
	return p.apply(ActionRevoked, c.ActorKey, nil, "", c)
}
func (p Policy) Join(member string, c Command) (Policy, error) {
	if c.ActorKey != p.state.HostKey || !keyPattern.MatchString(member) || slices.Contains(p.state.Participants, member) {
		return p, ErrDenied
	}
	return p.apply(ActionJoined, member, nil, "", c)
}
func (p Policy) Leave(member string, c Command) (Policy, error) {
	if c.ActorKey != p.state.HostKey || member == p.state.HostKey || !slices.Contains(p.state.Participants, member) {
		return p, ErrDenied
	}
	return p.apply(ActionLeft, member, nil, "", c)
}
func (p Policy) Start(ref string, c Command) (Policy, error) {
	if c.ActorKey != p.state.HostKey || p.state.Proposal == nil || p.state.Active || !allConsented(p.state.Participants, p.state.Consents) || !keyPattern.MatchString(ref) {
		return p, ErrConsentRequired
	}
	return p.apply(ActionStarted, c.ActorKey, nil, ref, c)
}
func (p Policy) Stop(c Command) (Policy, error) {
	if c.ActorKey != p.state.HostKey || !p.state.Active {
		return p, ErrDenied
	}
	return p.apply(ActionStopped, c.ActorKey, nil, "", c)
}
func (p Policy) apply(action Action, subject string, proposal *Proposal, ref string, c Command) (Policy, error) {
	for _, e := range p.state.Events {
		if e.CommandID == c.ID {
			if e.Fingerprint != c.Fingerprint {
				return p, ErrCommandMismatch
			}
			return p, nil
		}
	}
	if !idPattern.MatchString(c.ID) || !keyPattern.MatchString(c.ActorKey) || !keyPattern.MatchString(subject) || !keyPattern.MatchString(c.Fingerprint) || c.At.IsZero() {
		return Policy{}, ErrInvalid
	}
	if c.ExpectedRevision != p.state.Revision {
		return Policy{}, ErrStaleRevision
	}
	p.state.Participants = append([]string(nil), p.state.Participants...)
	p.state.Consents = append([]string(nil), p.state.Consents...)
	switch action {
	case ActionProposed:
		copy := *proposal
		p.state.Proposal = &copy
		p.state.Consents = nil
		p.state.Active = false
		p.state.RecordingRef = ""
	case ActionOptedIn:
		p.state.Consents = add(p.state.Consents, subject)
	case ActionRevoked:
		p.state.Consents = remove(p.state.Consents, subject)
		p.state.Active = false
		p.state.RecordingRef = ""
	case ActionJoined:
		p.state.Participants = add(p.state.Participants, subject)
		p.state.Active = false
		p.state.RecordingRef = ""
	case ActionLeft:
		p.state.Participants = remove(p.state.Participants, subject)
		p.state.Consents = remove(p.state.Consents, subject)
		p.state.Active = false
		p.state.RecordingRef = ""
	case ActionStarted:
		p.state.Active = true
		p.state.RecordingRef = ref
	case ActionStopped:
		p.state.Active = false
		p.state.RecordingRef = ""
	}
	p.state.Revision++
	event := Event{Sequence: p.state.Revision, Action: action, CommandID: c.ID, ActorKey: c.ActorKey, SubjectKey: subject, Fingerprint: c.Fingerprint, At: c.At.UTC()}
	if proposal != nil {
		event.Purpose = proposal.Purpose
		event.Retention = proposal.Retention
	}
	p.state.Events = append(append([]Event(nil), p.state.Events...), event)
	return p, nil
}
func Fingerprint(id string, action Action, actor, subject string, purpose Purpose, retention time.Duration, revision uint64) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + actor + "\x00" + subject + "\x00" + string(purpose) + "\x00" + retention.String() + "\x00" + strconv.FormatUint(revision, 10)))
	return hex.EncodeToString(sum[:])
}
func validPurpose(p Purpose) bool {
	return p == PurposeReflection || p == PurposeArchive || p == PurposeMemory
}
func set(v []string) ([]string, bool) {
	x := append([]string(nil), v...)
	for _, k := range x {
		if !keyPattern.MatchString(k) {
			return nil, false
		}
	}
	slices.Sort(x)
	for i := 1; i < len(x); i++ {
		if x[i] == x[i-1] {
			return nil, false
		}
	}
	return x, true
}
func add(v []string, x string) []string {
	if !slices.Contains(v, x) {
		v = append(v, x)
		slices.Sort(v)
	}
	return v
}
func remove(v []string, x string) []string {
	i := slices.Index(v, x)
	if i < 0 {
		return v
	}
	return append(v[:i:i], v[i+1:]...)
}
func allConsented(p, c []string) bool { return slices.Equal(p, c) }
func (p Policy) State() State {
	s := p.state
	s.Participants = append([]string(nil), s.Participants...)
	s.Consents = append([]string(nil), s.Consents...)
	s.Events = append([]Event(nil), s.Events...)
	if s.Proposal != nil {
		x := *s.Proposal
		s.Proposal = &x
	}
	return s
}
func (p Policy) ID() string       { return p.state.ID }
func (p Policy) Revision() uint64 { return p.state.Revision }
