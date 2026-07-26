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

var (
	ErrInvalid    = errors.New("invalid retention record")
	ErrStale      = errors.New("stale retention command")
	ErrMismatch   = errors.New("retention command mismatch")
	ErrHeld       = errors.New("legal hold prevents erasure")
	ErrTransition = errors.New("invalid retention transition")
)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Policy struct {
	DataClass   string        `bson:"dataClass"`
	Purpose     string        `bson:"purpose"`
	Version     uint64        `bson:"version"`
	RetainFor   time.Duration `bson:"retainFor"`
	EffectiveAt time.Time     `bson:"effectiveAt"`
}

func NewPolicy(class, purpose string, version uint64, retain time.Duration, at time.Time) (Policy, error) {
	p := Policy{class, purpose, version, retain, at.UTC()}
	if !token.MatchString(class) || !token.MatchString(purpose) || version == 0 || retain < 24*time.Hour || retain > 10*365*24*time.Hour || at.IsZero() {
		return Policy{}, ErrInvalid
	}
	return p, nil
}

type Status string

const (
	StatusRetained         Status = "retained"
	StatusErasureRequested Status = "erasure_requested"
	StatusErased           Status = "erased"
)

type Hold struct {
	CaseKey    string    `bson:"caseKey"`
	Active     bool      `bson:"active"`
	PlacedAt   time.Time `bson:"placedAt"`
	ReleasedAt time.Time `bson:"releasedAt,omitempty"`
}
type Command struct {
	ID               string    `bson:"id"`
	ExpectedRevision uint64    `bson:"expectedRevision"`
	At               time.Time `bson:"at"`
}
type Event struct {
	Sequence  uint64    `bson:"sequence"`
	CommandID string    `bson:"commandId"`
	Action    string    `bson:"action"`
	At        time.Time `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type Counters struct {
	Holds            uint64 `bson:"holds"`
	ErasureRequests  uint64 `bson:"erasureRequests"`
	VerifiedErasures uint64 `bson:"verifiedErasures"`
}
type Record struct {
	id, subjectKey                 string
	policy                         Policy
	status                         Status
	createdAt, expiresAt, erasedAt time.Time
	hold                           Hold
	verificationKey                string
	revision                       uint64
	events                         []Event
	commands                       []Applied
	counters                       Counters
}
type State struct {
	ID, SubjectKey                 string
	Policy                         Policy
	Status                         Status
	CreatedAt, ExpiresAt, ErasedAt time.Time
	Hold                           Hold
	VerificationKey                string
	Revision                       uint64
	Events                         []Event
	Commands                       []Applied
	Counters                       Counters
}

func Create(id, subject string, p Policy, c Command) (Record, error) {
	r := Record{id: id, subjectKey: subject, policy: p, status: StatusRetained, createdAt: c.At.UTC(), expiresAt: c.At.UTC().Add(p.RetainFor)}
	if !opaque.MatchString(id) || !key.MatchString(subject) || c.ExpectedRevision != 0 || c.At.Before(p.EffectiveAt) {
		return Record{}, ErrInvalid
	}
	if e := r.apply(c, "create"); e != nil {
		return Record{}, e
	}
	return r, nil
}
func Rehydrate(s State) (Record, error) {
	r := Record{id: s.ID, subjectKey: s.SubjectKey, policy: s.Policy, status: s.Status, createdAt: s.CreatedAt, expiresAt: s.ExpiresAt, erasedAt: s.ErasedAt, hold: s.Hold, verificationKey: s.VerificationKey, revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...), counters: s.Counters}
	if !opaque.MatchString(r.id) || !key.MatchString(r.subjectKey) || r.revision == 0 || len(r.events) != int(r.revision) || len(r.commands) != int(r.revision) {
		return Record{}, ErrInvalid
	}
	return r, nil
}
func (r Record) PlaceHold(caseKey string, c Command) (Record, error) {
	if replay, e := r.replay(c, "hold", caseKey); replay || e != nil {
		return r, e
	}
	if !key.MatchString(caseKey) || r.status == StatusErased || r.hold.Active {
		return Record{}, ErrTransition
	}
	r = r.clone()
	r.hold = Hold{CaseKey: caseKey, Active: true, PlacedAt: c.At.UTC()}
	r.counters.Holds++
	return r, r.apply(c, "hold", caseKey)
}
func (r Record) ReleaseHold(caseKey string, c Command) (Record, error) {
	if replay, e := r.replay(c, "release", caseKey); replay || e != nil {
		return r, e
	}
	if !r.hold.Active || r.hold.CaseKey != caseKey {
		return Record{}, ErrTransition
	}
	r = r.clone()
	r.hold.Active = false
	r.hold.ReleasedAt = c.At.UTC()
	return r, r.apply(c, "release", caseKey)
}
func (r Record) RequestErasure(c Command) (Record, error) {
	if replay, e := r.replay(c, "request-erasure"); replay || e != nil {
		return r, e
	}
	if r.hold.Active {
		return Record{}, ErrHeld
	}
	if r.status != StatusRetained {
		return Record{}, ErrTransition
	}
	r = r.clone()
	r.status = StatusErasureRequested
	r.counters.ErasureRequests++
	return r, r.apply(c, "request-erasure")
}
func (r Record) CompleteErasure(verification string, c Command) (Record, error) {
	if replay, e := r.replay(c, "complete-erasure", verification); replay || e != nil {
		return r, e
	}
	if r.hold.Active {
		return Record{}, ErrHeld
	}
	if r.status != StatusErasureRequested || !key.MatchString(verification) {
		return Record{}, ErrTransition
	}
	r = r.clone()
	r.status = StatusErased
	r.erasedAt = c.At.UTC()
	r.verificationKey = verification
	r.counters.VerifiedErasures++
	return r, r.apply(c, "complete-erasure", verification)
}
func (r Record) clone() Record {
	r.events = append([]Event(nil), r.events...)
	r.commands = append([]Applied(nil), r.commands...)
	return r
}
func (r *Record) apply(c Command, a string, v ...string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != r.revision {
		return ErrStale
	}
	f := fingerprint(r.id, c, a, v...)
	r.revision++
	r.events = append(r.events, Event{r.revision, c.ID, a, c.At.UTC()})
	r.commands = append(r.commands, Applied{c.ID, f, r.revision})
	return nil
}
func (r Record) replay(c Command, a string, v ...string) (bool, error) {
	f := fingerprint(r.id, c, a, v...)
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
func fingerprint(id string, c Command, a string, v ...string) string {
	s := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strings.Join(v, "\x00") + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(s[:])
}
func (r Record) ID() string              { return r.id }
func (r Record) SubjectKey() string      { return r.subjectKey }
func (r Record) Policy() Policy          { return r.policy }
func (r Record) Status() Status          { return r.status }
func (r Record) CreatedAt() time.Time    { return r.createdAt }
func (r Record) ExpiresAt() time.Time    { return r.expiresAt }
func (r Record) ErasedAt() time.Time     { return r.erasedAt }
func (r Record) Hold() Hold              { return r.hold }
func (r Record) VerificationKey() string { return r.verificationKey }
func (r Record) Revision() uint64        { return r.revision }
func (r Record) Events() []Event         { return append([]Event(nil), r.events...) }
func (r Record) Commands() []Applied     { return append([]Applied(nil), r.commands...) }
func (r Record) Counters() Counters      { return r.counters }
