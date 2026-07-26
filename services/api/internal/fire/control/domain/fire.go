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

type Role string
type Action string

const (
	RoleHost        Role   = "host"
	RoleCohost      Role   = "cohost"
	RoleParticipant Role   = "participant"
	ActionOpened    Action = "opened"
	ActionPromoted  Action = "promoted"
	ActionDemoted   Action = "demoted"
	ActionMuted     Action = "muted"
	ActionEjected   Action = "ejected"
)

var (
	ErrInvalid         = errors.New("invalid fire control state")
	ErrDenied          = errors.New("fire control unavailable")
	ErrEjected         = errors.New("participant access revoked")
	ErrStaleRevision   = errors.New("stale fire control revision")
	ErrCommandMismatch = errors.New("fire control replay mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var reasonPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,63}$`)

type Member struct {
	Key            string
	Role           Role
	Muted, Ejected bool
}
type Command struct {
	ID, ActorKey, ReasonCode, Fingerprint string
	ExpectedRevision                      uint64
	At                                    time.Time
}
type Event struct {
	Sequence                                                uint64
	Action                                                  Action
	CommandID, ActorKey, TargetKey, ReasonCode, Fingerprint string
	At                                                      time.Time
}
type State struct {
	ID, FireKey string
	Members     []Member
	Revision    uint64
	Events      []Event
}
type Fire struct{ state State }

func Open(id, fireKey, hostKey string, participantKeys []string, c Command) (Fire, error) {
	if !idPattern.MatchString(id) || !keyPattern.MatchString(fireKey) || !keyPattern.MatchString(hostKey) || c.ActorKey != hostKey || c.ExpectedRevision != 0 || len(participantKeys) > 500 {
		return Fire{}, ErrInvalid
	}
	members := []Member{{Key: hostKey, Role: RoleHost}}
	seen := map[string]bool{hostKey: true}
	for _, key := range participantKeys {
		if !keyPattern.MatchString(key) || seen[key] {
			return Fire{}, ErrInvalid
		}
		seen[key] = true
		members = append(members, Member{Key: key, Role: RoleParticipant})
	}
	sortMembers(members)
	f := Fire{State{ID: id, FireKey: fireKey, Members: members}}
	return f.apply(ActionOpened, hostKey, c)
}
func Rehydrate(s State) (Fire, error) {
	if !idPattern.MatchString(s.ID) || !keyPattern.MatchString(s.FireKey) || s.Revision == 0 || len(s.Events) != int(s.Revision) {
		return Fire{}, ErrInvalid
	}
	f := Fire{state: s}
	f.state.Members = append([]Member(nil), s.Members...)
	f.state.Events = append([]Event(nil), s.Events...)
	hostCount := 0
	for _, m := range f.state.Members {
		if !keyPattern.MatchString(m.Key) || m.Ejected && m.Role != RoleParticipant || m.Role != RoleHost && m.Role != RoleCohost && m.Role != RoleParticipant {
			return Fire{}, ErrInvalid
		}
		if m.Role == RoleHost {
			hostCount++
		}
	}
	if hostCount != 1 {
		return Fire{}, ErrInvalid
	}
	for i, e := range f.state.Events {
		if e.Sequence != uint64(i+1) || !idPattern.MatchString(e.CommandID) || !keyPattern.MatchString(e.ActorKey) || !keyPattern.MatchString(e.TargetKey) || !reasonPattern.MatchString(e.ReasonCode) || !keyPattern.MatchString(e.Fingerprint) {
			return Fire{}, ErrInvalid
		}
	}
	return f, nil
}
func (f Fire) Promote(target string, c Command) (Fire, error) {
	return f.change(ActionPromoted, target, c)
}
func (f Fire) Demote(target string, c Command) (Fire, error) {
	return f.change(ActionDemoted, target, c)
}
func (f Fire) Mute(target string, c Command) (Fire, error) { return f.change(ActionMuted, target, c) }
func (f Fire) Eject(target string, c Command) (Fire, error) {
	return f.change(ActionEjected, target, c)
}
func (f Fire) change(action Action, target string, c Command) (Fire, error) {
	for _, event := range f.state.Events {
		if event.CommandID == c.ID {
			if event.Fingerprint != c.Fingerprint {
				return f, ErrCommandMismatch
			}
			return f, nil
		}
	}
	actorIndex, targetIndex := f.index(c.ActorKey), f.index(target)
	if actorIndex < 0 || targetIndex < 0 || f.state.Members[actorIndex].Ejected || f.state.Members[targetIndex].Ejected {
		return f, ErrDenied
	}
	actor, targetMember := f.state.Members[actorIndex], f.state.Members[targetIndex]
	switch action {
	case ActionPromoted:
		if actor.Role != RoleHost || targetMember.Role != RoleParticipant {
			return f, ErrDenied
		}
	case ActionDemoted:
		if actor.Role != RoleHost || targetMember.Role != RoleCohost {
			return f, ErrDenied
		}
	case ActionMuted, ActionEjected:
		if actor.Role != RoleHost && actor.Role != RoleCohost || targetMember.Role != RoleParticipant {
			return f, ErrDenied
		}
	default:
		return f, ErrInvalid
	}
	return f.apply(action, target, c)
}
func (f Fire) apply(action Action, target string, c Command) (Fire, error) {
	if !idPattern.MatchString(c.ID) || !keyPattern.MatchString(c.ActorKey) || !keyPattern.MatchString(target) || !reasonPattern.MatchString(c.ReasonCode) || !keyPattern.MatchString(c.Fingerprint) || c.At.IsZero() {
		return Fire{}, ErrInvalid
	}
	if c.ExpectedRevision != f.state.Revision {
		return Fire{}, ErrStaleRevision
	}
	f.state.Members = append([]Member(nil), f.state.Members...)
	index := f.index(target)
	switch action {
	case ActionPromoted:
		f.state.Members[index].Role = RoleCohost
	case ActionDemoted:
		f.state.Members[index].Role = RoleParticipant
	case ActionMuted:
		f.state.Members[index].Muted = true
	case ActionEjected:
		f.state.Members[index].Ejected = true
		f.state.Members[index].Muted = true
	}
	f.state.Revision++
	f.state.Events = append(append([]Event(nil), f.state.Events...), Event{f.state.Revision, action, c.ID, c.ActorKey, target, c.ReasonCode, c.Fingerprint, c.At.UTC()})
	return f, nil
}
func Fingerprint(id string, action Action, actor, target, reason string, revision uint64) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + actor + "\x00" + target + "\x00" + reason + "\x00" + strconv.FormatUint(revision, 10)))
	return hex.EncodeToString(sum[:])
}
func (f Fire) index(key string) int {
	return slices.IndexFunc(f.state.Members, func(m Member) bool { return m.Key == key })
}
func sortMembers(m []Member) {
	slices.SortFunc(m, func(a, b Member) int {
		if a.Key < b.Key {
			return -1
		}
		if a.Key > b.Key {
			return 1
		}
		return 0
	})
}
func (f Fire) State() State {
	s := f.state
	s.Members = append([]Member(nil), s.Members...)
	s.Events = append([]Event(nil), s.Events...)
	return s
}
func (f Fire) ID() string       { return f.state.ID }
func (f Fire) Revision() uint64 { return f.state.Revision }
