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

type Action string

const (
	ActionGrant  Action = "grant"
	ActionRevoke Action = "revoke"
)

var (
	ErrInvalid         = errors.New("invalid honesty ribbon")
	ErrDenied          = errors.New("honesty ribbon unavailable")
	ErrStaleRevision   = errors.New("stale honesty ribbon revision")
	ErrCommandMismatch = errors.New("honesty command replay mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)

type Command struct {
	ID, ActorKey     string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence            uint64
	CommandID, ActorKey string
	Action              Action
	At                  time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Ribbon struct {
	id              string
	members, grants []string
	revision        uint64
	events          []Event
	commands        []Applied
}
type State struct {
	ID              string
	Members, Grants []string
	Revision        uint64
	Events          []Event
	Commands        []Applied
}

func New(id string, members []string) (Ribbon, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) {
		return Ribbon{}, ErrInvalid
	}
	return Ribbon{id: id, members: m}, nil
}
func Rehydrate(s State) (Ribbon, error) {
	r, e := New(s.ID, s.Members)
	if e != nil {
		return Ribbon{}, e
	}
	r.grants = append([]string(nil), s.Grants...)
	slices.Sort(r.grants)
	r.grants = slices.Compact(r.grants)
	r.revision = s.Revision
	r.events = append([]Event(nil), s.Events...)
	r.commands = append([]Applied(nil), s.Commands...)
	if len(r.grants) > 2 || len(r.events) != int(r.revision) || len(r.commands) != int(r.revision) {
		return Ribbon{}, ErrInvalid
	}
	return r, nil
}
func (r Ribbon) Grant(c Command) (Ribbon, error)  { return r.apply(ActionGrant, c) }
func (r Ribbon) Revoke(c Command) (Ribbon, error) { return r.apply(ActionRevoke, c) }
func (r Ribbon) apply(a Action, c Command) (Ribbon, error) {
	if !slices.Contains(r.members, c.ActorKey) || !idPattern.MatchString(c.ID) || c.At.IsZero() {
		return Ribbon{}, ErrDenied
	}
	fp := fingerprint(r.id, a, c)
	for _, x := range r.commands {
		if x.ID == c.ID {
			if x.Fingerprint != fp {
				return Ribbon{}, ErrCommandMismatch
			}
			return r, nil
		}
	}
	if c.ExpectedRevision != r.revision {
		return Ribbon{}, ErrStaleRevision
	}
	has := slices.Contains(r.grants, c.ActorKey)
	if a == ActionGrant && !has {
		r.grants = append(r.grants, c.ActorKey)
		slices.Sort(r.grants)
	} else if a == ActionRevoke && has {
		r.grants = deleteValue(r.grants, c.ActorKey)
	} else {
		return Ribbon{}, ErrDenied
	}
	r.revision++
	r.events = append(r.events, Event{r.revision, c.ID, c.ActorKey, a, c.At.UTC()})
	r.commands = append(r.commands, Applied{c.ID, fp, r.revision})
	return r, nil
}
func deleteValue(values []string, target string) []string {
	out := values[:0]
	for _, v := range values {
		if v != target {
			out = append(out, v)
		}
	}
	return out
}
func fingerprint(id string, a Action, c Command) string {
	x := sha256.Sum256([]byte(id + "\x00" + string(a) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(x[:])
}
func (r Ribbon) Visible() bool       { return len(r.grants) == 2 }
func (r Ribbon) ID() string          { return r.id }
func (r Ribbon) Members() []string   { return append([]string(nil), r.members...) }
func (r Ribbon) Grants() []string    { return append([]string(nil), r.grants...) }
func (r Ribbon) Revision() uint64    { return r.revision }
func (r Ribbon) Events() []Event     { return append([]Event(nil), r.events...) }
func (r Ribbon) Commands() []Applied { return append([]Applied(nil), r.commands...) }
