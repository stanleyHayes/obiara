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

type Status string
type Kind string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
	KindMember   Kind   = "member"
	KindInactive Kind   = "inactive"
)

var (
	ErrInvalid         = errors.New("invalid closure")
	ErrDenied          = errors.New("closure unavailable")
	ErrNotDue          = errors.New("inactivity threshold not reached")
	ErrStaleRevision   = errors.New("stale closure revision")
	ErrCommandMismatch = errors.New("closure command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)
)

type Command struct {
	ID, ActorKey     string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence  uint64
	CommandID string
	Kind      Kind
	At        time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Closure struct {
	id                     string
	members                []string
	status                 Status
	lastActivity, closedAt time.Time
	revision               uint64
	event                  *Event
	commands               []Applied
}
type State struct {
	ID                     string
	Members                []string
	Status                 Status
	LastActivity, ClosedAt time.Time
	Revision               uint64
	Event                  *Event
	Commands               []Applied
}

func New(id string, members []string, lastActivity time.Time) (Closure, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) || lastActivity.IsZero() {
		return Closure{}, ErrInvalid
	}
	return Closure{id: id, members: m, status: StatusOpen, lastActivity: lastActivity.UTC()}, nil
}
func Rehydrate(s State) (Closure, error) {
	c, err := New(s.ID, s.Members, s.LastActivity)
	if err != nil {
		return Closure{}, err
	}
	c.status, c.closedAt, c.revision, c.event, c.commands = s.Status, s.ClosedAt, s.Revision, s.Event, append([]Applied(nil), s.Commands...)
	if c.status == StatusOpen && (c.revision != 0 || c.event != nil || !c.closedAt.IsZero()) {
		return Closure{}, ErrInvalid
	}
	if c.status == StatusClosed && (c.revision != 1 || c.event == nil || c.closedAt.IsZero() || len(c.commands) != 1) {
		return Closure{}, ErrInvalid
	}
	return c, nil
}
func (c Closure) Close(cmd Command) (Closure, error) {
	if !slices.Contains(c.members, cmd.ActorKey) {
		return Closure{}, ErrDenied
	}
	return c.apply(KindMember, cmd, 0)
}
func (c Closure) CloseInactive(cmd Command, threshold time.Duration) (Closure, error) {
	if cmd.ActorKey != "" || threshold <= 0 {
		return Closure{}, ErrDenied
	}
	return c.apply(KindInactive, cmd, threshold)
}
func (c Closure) apply(kind Kind, cmd Command, threshold time.Duration) (Closure, error) {
	if !idPattern.MatchString(cmd.ID) || cmd.At.IsZero() {
		return Closure{}, ErrInvalid
	}
	fp := fingerprint(c.id, kind, cmd, threshold)
	for _, seen := range c.commands {
		if seen.ID == cmd.ID {
			if seen.Fingerprint != fp {
				return Closure{}, ErrCommandMismatch
			}
			return c, nil
		}
	}
	if c.status == StatusClosed {
		return Closure{}, ErrDenied
	}
	if cmd.ExpectedRevision != c.revision {
		return Closure{}, ErrStaleRevision
	}
	if kind == KindInactive && cmd.At.Before(c.lastActivity.Add(threshold)) {
		return Closure{}, ErrNotDue
	}
	c.status, c.closedAt, c.revision = StatusClosed, cmd.At.UTC(), c.revision+1
	c.event = &Event{Sequence: c.revision, CommandID: cmd.ID, Kind: kind, At: c.closedAt}
	c.commands = append(c.commands, Applied{cmd.ID, fp, c.revision})
	return c, nil
}
func fingerprint(id string, kind Kind, c Command, threshold time.Duration) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(kind) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + threshold.String()))
	return hex.EncodeToString(sum[:])
}
func (c Closure) ID() string              { return c.id }
func (c Closure) Members() []string       { return append([]string(nil), c.members...) }
func (c Closure) Status() Status          { return c.status }
func (c Closure) LastActivity() time.Time { return c.lastActivity }
func (c Closure) ClosedAt() time.Time     { return c.closedAt }
func (c Closure) Revision() uint64        { return c.revision }
func (c Closure) Event() *Event {
	if c.event == nil {
		return nil
	}
	v := *c.event
	return &v
}
func (c Closure) Commands() []Applied { return append([]Applied(nil), c.commands...) }
