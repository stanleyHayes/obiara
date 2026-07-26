package domain

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalid = errors.New("invalid community audit case")
	ErrClosed  = errors.New("community audit case closed")
	ErrStale   = errors.New("stale community audit version")
	ErrCommand = errors.New("community audit command conflict")
)

type Kind string

const (
	KindCircle Kind = "circle_legitimacy"
	KindVouch  Kind = "vouch"
)

type Status string

const (
	StatusQueued   Status = "queued"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

var op = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)
var reason = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,63}$`)

type Audit struct {
	command, actor, reason, correlation string
	from, to                            Status
	at                                  time.Time
}

func (a Audit) Command() string     { return a.command }
func (a Audit) Actor() string       { return a.actor }
func (a Audit) Reason() string      { return a.reason }
func (a Audit) Correlation() string { return a.correlation }
func (a Audit) From() Status        { return a.from }
func (a Audit) To() Status          { return a.to }
func (a Audit) At() time.Time       { return a.at }
func NewAudit(command, actor, why, correlation string, from, to Status, at time.Time) (Audit, error) {
	if !op.MatchString(command) || !key(actor) || !reason.MatchString(why) || !key(correlation) || at.IsZero() {
		return Audit{}, ErrInvalid
	}
	return Audit{command, actor, why, correlation, from, to, at.UTC()}, nil
}

type Case struct {
	id                                   string
	kind                                 Kind
	subjectRef, evidenceRef, summaryCode string
	status                               Status
	version                              uint64
	created                              time.Time
	audit                                []Audit
}

func New(id string, k Kind, subject, evidence, summary string, at time.Time) (Case, error) {
	if !op.MatchString(id) || !op.MatchString(subject) || !op.MatchString(evidence) || !reason.MatchString(summary) || at.IsZero() || (k != KindCircle && k != KindVouch) {
		return Case{}, ErrInvalid
	}
	return Case{id: id, kind: k, subjectRef: subject, evidenceRef: evidence, summaryCode: summary, status: StatusQueued, version: 1, created: at.UTC()}, nil
}
func Rehydrate(id string, k Kind, s, e, sum string, status Status, v uint64, created time.Time, a []Audit) (Case, error) {
	c, x := New(id, k, s, e, sum, created)
	if x != nil || v == 0 || uint64(len(a))+1 != v {
		return Case{}, ErrInvalid
	}
	c.status = status
	c.version = v
	c.audit = append([]Audit(nil), a...)
	return c, nil
}
func (c Case) Decide(approve bool, command, actor, why, correlation string, at time.Time, expected uint64) (Case, error) {
	target := StatusRejected
	if approve {
		target = StatusApproved
	}
	for _, a := range c.audit {
		if a.command == command {
			if a.to == target && a.actor == actor && a.reason == why {
				return c, nil
			}
			return Case{}, ErrCommand
		}
	}
	if c.status != StatusQueued {
		return Case{}, ErrClosed
	}
	if expected != c.version {
		return Case{}, ErrStale
	}
	a, e := NewAudit(command, actor, why, correlation, c.status, target, at)
	if e != nil {
		return Case{}, e
	}
	c.status = target
	c.version++
	c.audit = append(c.Audit(), a)
	return c, nil
}
func (c Case) ID() string          { return c.id }
func (c Case) Kind() Kind          { return c.kind }
func (c Case) SubjectRef() string  { return c.subjectRef }
func (c Case) EvidenceRef() string { return c.evidenceRef }
func (c Case) SummaryCode() string { return c.summaryCode }
func (c Case) Status() Status      { return c.status }
func (c Case) Version() uint64     { return c.version }
func (c Case) Created() time.Time  { return c.created }
func (c Case) Audit() []Audit      { return append([]Audit(nil), c.audit...) }
func key(s string) bool            { return len(s) == 64 && regexp.MustCompile(`^[a-f0-9]+$`).MatchString(s) }
