// Package domain models privacy-preserving identity collision reviews.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type Kind string
type Status string
type Resolution string

const (
	KindSharedDevice Kind = "shared_device"
	KindKnownName    Kind = "known_name"

	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDenied   Status = "denied"

	ResolutionApprove Resolution = "approve"
	ResolutionDeny    Resolution = "deny"
)

var (
	ErrInvalidKind       = errors.New("invalid collision kind")
	ErrKeyRequired       = errors.New("privacy-preserving key is required")
	ErrCaseIDRequired    = errors.New("review case id is required")
	ErrCaseClosed        = errors.New("review case is closed")
	ErrInvalidReason     = errors.New("bounded reason code is required")
	ErrActorRequired     = errors.New("privacy-preserving actor key is required")
	ErrInvalidResolution = errors.New("invalid review resolution")
)

var reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{1,63}$`)

type Case struct {
	id         string
	kind       Kind
	signalKey  string
	subjectKey string
	status     Status
	reasonCode string
	version    int64
	createdAt  time.Time
	resolvedAt *time.Time
}

type AuditEvent struct {
	CaseID     string
	Sequence   int64
	From       Status
	To         Status
	ReasonCode string
	ActorKey   string
	OccurredAt time.Time
}

func ValidKind(kind Kind) bool {
	return kind == KindSharedDevice || kind == KindKnownName
}

func NewCase(id string, kind Kind, signalKey, subjectKey string, now time.Time) (Case, AuditEvent, error) {
	if strings.TrimSpace(id) == "" {
		return Case{}, AuditEvent{}, ErrCaseIDRequired
	}
	if !ValidKind(kind) {
		return Case{}, AuditEvent{}, ErrInvalidKind
	}
	if strings.TrimSpace(signalKey) == "" || strings.TrimSpace(subjectKey) == "" {
		return Case{}, AuditEvent{}, ErrKeyRequired
	}
	createdAt := now.UTC()
	reviewCase := Case{
		id: id, kind: kind, signalKey: signalKey, subjectKey: subjectKey,
		status: StatusPending, version: 1, createdAt: createdAt,
	}
	return reviewCase, AuditEvent{
		CaseID: id, Sequence: 1, To: StatusPending,
		ReasonCode: "collision_detected", ActorKey: "system", OccurredAt: createdAt,
	}, nil
}

func ReconstituteCase(id string, kind Kind, signalKey, subjectKey string, status Status, reasonCode string, version int64, createdAt time.Time, resolvedAt *time.Time) Case {
	return Case{id: id, kind: kind, signalKey: signalKey, subjectKey: subjectKey, status: status, reasonCode: reasonCode, version: version, createdAt: createdAt, resolvedAt: resolvedAt}
}

func (reviewCase *Case) Resolve(resolution Resolution, reasonCode, actorKey string, now time.Time) (AuditEvent, error) {
	if reviewCase.status != StatusPending {
		return AuditEvent{}, ErrCaseClosed
	}
	if !reasonPattern.MatchString(reasonCode) {
		return AuditEvent{}, ErrInvalidReason
	}
	if strings.TrimSpace(actorKey) == "" {
		return AuditEvent{}, ErrActorRequired
	}
	var target Status
	switch resolution {
	case ResolutionApprove:
		target = StatusApproved
	case ResolutionDeny:
		target = StatusDenied
	default:
		return AuditEvent{}, ErrInvalidResolution
	}
	from := reviewCase.status
	resolvedAt := now.UTC()
	reviewCase.status = target
	reviewCase.reasonCode = reasonCode
	reviewCase.version++
	reviewCase.resolvedAt = &resolvedAt
	return AuditEvent{
		CaseID: reviewCase.id, Sequence: reviewCase.version, From: from, To: target,
		ReasonCode: reasonCode, ActorKey: actorKey, OccurredAt: resolvedAt,
	}, nil
}

// Allowed is deliberately conservative: an unresolved shared-device collision
// never grants access. A known-name collision is review-only unless an
// operator explicitly denies it.
func (reviewCase Case) Allowed() bool {
	if reviewCase.status == StatusApproved {
		return true
	}
	if reviewCase.status == StatusDenied {
		return false
	}
	return reviewCase.kind == KindKnownName
}

func (reviewCase Case) ID() string             { return reviewCase.id }
func (reviewCase Case) Kind() Kind             { return reviewCase.kind }
func (reviewCase Case) SignalKey() string      { return reviewCase.signalKey }
func (reviewCase Case) SubjectKey() string     { return reviewCase.subjectKey }
func (reviewCase Case) Status() Status         { return reviewCase.status }
func (reviewCase Case) ReasonCode() string     { return reviewCase.reasonCode }
func (reviewCase Case) Version() int64         { return reviewCase.version }
func (reviewCase Case) CreatedAt() time.Time   { return reviewCase.createdAt }
func (reviewCase Case) ResolvedAt() *time.Time { return reviewCase.resolvedAt }
