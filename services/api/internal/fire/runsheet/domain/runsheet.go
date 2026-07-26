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

type SegmentType string

const (
	TypeTalk  SegmentType = "talk"
	TypeBreak SegmentType = "break"
	TypeGame  SegmentType = "game"
	TypeClose SegmentType = "close"
)

type Status string

const (
	StatusReady     Status = "ready"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
)
const MaxSegments = 20

var (
	ErrInvalid    = errors.New("invalid run sheet")
	ErrTransition = errors.New("invalid run sheet transition")
	ErrStale      = errors.New("stale run sheet revision")
	ErrMismatch   = errors.New("run sheet command replay mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)
var title = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{1,47}$`)
var allowedGames = map[string]bool{
	"1111111111111111111111111111111111111111111111111111111111111111": true,
	"2222222222222222222222222222222222222222222222222222222222222222": true,
	"3333333333333333333333333333333333333333333333333333333333333333": true,
}

type Segment struct {
	Type            SegmentType   `bson:"type"`
	TitleCode       string        `bson:"titleCode"`
	PlannedDuration time.Duration `bson:"plannedDuration"`
	CapabilityRef   string        `bson:"capabilityRef,omitempty"`
}
type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence  uint64        `bson:"sequence"`
	CommandID string        `bson:"commandId"`
	Action    string        `bson:"action"`
	Index     int           `bson:"index"`
	Delta     time.Duration `bson:"delta"`
	At        time.Time     `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type Projection struct {
	ID                            string
	Version                       uint64
	Status                        Status
	CurrentIndex                  int
	Current                       *Segment
	StartedAt, EndsAt, ServerTime time.Time
	Remaining                     time.Duration
}
type RunSheet struct {
	id, fireKey, hostKey string
	version              uint64
	segments             []Segment
	status               Status
	current              int
	startedAt, endsAt    time.Time
	revision             uint64
	events               []Event
	commands             []Applied
}
type State struct {
	ID, FireKey, HostKey string
	Version              uint64
	Segments             []Segment
	Status               Status
	Current              int
	StartedAt, EndsAt    time.Time
	Revision             uint64
	Events               []Event
	Commands             []Applied
}

func Create(id, fireKey, hostKey string, version uint64, segments []Segment, c Command) (RunSheet, error) {
	r := RunSheet{id: id, fireKey: fireKey, hostKey: hostKey, version: version, segments: clone(segments), status: StatusReady, current: -1}
	if !opaque.MatchString(id) || !key.MatchString(fireKey) || !key.MatchString(hostKey) || version == 0 || !validSegments(segments) || c.ExpectedRevision != 0 {
		return RunSheet{}, ErrInvalid
	}
	if e := r.apply(c, "create", -1, 0); e != nil {
		return RunSheet{}, e
	}
	return r, nil
}
func Rehydrate(s State) (RunSheet, error) {
	r := RunSheet{id: s.ID, fireKey: s.FireKey, hostKey: s.HostKey, version: s.Version, segments: clone(s.Segments), status: s.Status, current: s.Current, startedAt: s.StartedAt.UTC(), endsAt: s.EndsAt.UTC(), revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...)}
	if !opaque.MatchString(r.id) || !key.MatchString(r.fireKey) || !key.MatchString(r.hostKey) || r.version == 0 || !validSegments(r.segments) || len(r.events) != int(r.revision) || len(r.commands) != int(r.revision) || r.revision == 0 {
		return RunSheet{}, ErrInvalid
	}
	return r, nil
}
func (r RunSheet) Start(now time.Time, c Command) (RunSheet, error) {
	if replay, e := r.replay(c, "start", 0, 0); replay || e != nil {
		return r, e
	}
	if r.status != StatusReady {
		return RunSheet{}, ErrTransition
	}
	r.status, r.current, r.startedAt = StatusRunning, 0, now.UTC()
	r.endsAt = r.startedAt.Add(r.segments[0].PlannedDuration)
	return r, r.apply(c, "start", 0, 0)
}
func (r RunSheet) Advance(now time.Time, c Command) (RunSheet, error) {
	return r.move(now, c, "advance")
}
func (r RunSheet) Skip(now time.Time, c Command) (RunSheet, error) { return r.move(now, c, "skip") }
func (r RunSheet) move(now time.Time, c Command, action string) (RunSheet, error) {
	if replay, e := r.replay(c, action, r.current, 0); replay || e != nil {
		return r, e
	}
	if r.status != StatusRunning {
		return RunSheet{}, ErrTransition
	}
	if r.current+1 >= len(r.segments) {
		r.status = StatusCompleted
		r.current = len(r.segments)
		r.startedAt, r.endsAt = now.UTC(), now.UTC()
	} else {
		r.current++
		r.startedAt = now.UTC()
		r.endsAt = r.startedAt.Add(r.segments[r.current].PlannedDuration)
	}
	return r, r.apply(c, action, r.current, 0)
}
func (r RunSheet) Extend(delta time.Duration, c Command) (RunSheet, error) {
	if replay, e := r.replay(c, "extend", r.current, delta); replay || e != nil {
		return r, e
	}
	if r.status != StatusRunning || delta <= 0 || delta > time.Hour {
		return RunSheet{}, ErrTransition
	}
	r.endsAt = r.endsAt.Add(delta)
	return r, r.apply(c, "extend", r.current, delta)
}
func (r RunSheet) Project(now time.Time) Projection {
	p := Projection{ID: r.id, Version: r.version, Status: r.status, CurrentIndex: r.current, StartedAt: r.startedAt, EndsAt: r.endsAt, ServerTime: now.UTC()}
	if r.status == StatusRunning && r.current >= 0 && r.current < len(r.segments) {
		x := r.segments[r.current]
		p.Current = &x
		p.Remaining = r.endsAt.Sub(now.UTC())
		if p.Remaining < 0 {
			p.Remaining = 0
		}
	}
	return p
}
func (r *RunSheet) apply(c Command, action string, index int, delta time.Duration) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != r.revision {
		return ErrStale
	}
	f := fingerprint(r.id, c, action, index, delta)
	r.revision++
	r.events = append(r.events, Event{r.revision, c.ID, action, index, delta, c.At.UTC()})
	r.commands = append(r.commands, Applied{c.ID, f, r.revision})
	return nil
}
func (r RunSheet) replay(c Command, action string, index int, delta time.Duration) (bool, error) {
	f := fingerprint(r.id, c, action, index, delta)
	for _, x := range r.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func validSegments(v []Segment) bool {
	if len(v) == 0 || len(v) > MaxSegments {
		return false
	}
	for _, s := range v {
		if !title.MatchString(s.TitleCode) || s.PlannedDuration < time.Minute || s.PlannedDuration > 4*time.Hour {
			return false
		}
		switch s.Type {
		case TypeTalk, TypeBreak, TypeClose:
			if s.CapabilityRef != "" {
				return false
			}
		case TypeGame:
			if !allowedGames[s.CapabilityRef] {
				return false
			}
		default:
			return false
		}
	}
	return true
}
func clone(v []Segment) []Segment { return append([]Segment(nil), v...) }
func fingerprint(id string, c Command, a string, i int, d time.Duration) string {
	s := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strconv.Itoa(i) + "\x00" + strconv.FormatInt(int64(d), 10) + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(s[:])
}
func AllowedGameCapabilities() []string {
	out := make([]string, 0, len(allowedGames))
	for v := range allowedGames {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}
func (r RunSheet) ID() string           { return r.id }
func (r RunSheet) FireKey() string      { return r.fireKey }
func (r RunSheet) HostKey() string      { return r.hostKey }
func (r RunSheet) Version() uint64      { return r.version }
func (r RunSheet) Segments() []Segment  { return clone(r.segments) }
func (r RunSheet) Status() Status       { return r.status }
func (r RunSheet) Current() int         { return r.current }
func (r RunSheet) StartedAt() time.Time { return r.startedAt }
func (r RunSheet) EndsAt() time.Time    { return r.endsAt }
func (r RunSheet) Revision() uint64     { return r.revision }
func (r RunSheet) Events() []Event      { return append([]Event(nil), r.events...) }
func (r RunSheet) Commands() []Applied  { return append([]Applied(nil), r.commands...) }
