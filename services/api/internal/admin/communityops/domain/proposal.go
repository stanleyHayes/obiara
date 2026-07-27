package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid community operation proposal")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)

const MaxCapacity = 5000

type Operation string

const (
	AssignHost     Operation = "assign_host"
	ReplaceHost    Operation = "replace_host"
	CancelFire     Operation = "cancel_fire"
	RescheduleFire Operation = "reschedule_fire"
)

type Reason string

const (
	ReasonHostUnavailable      Reason = "host_unavailable"
	ReasonCertificationExpired Reason = "certification_expired"
	ReasonSafetyCapacity       Reason = "safety_capacity"
	ReasonScheduleConflict     Reason = "schedule_conflict"
)

type Density struct {
	CircleKey, FireKey     string
	Participants, Capacity int
	Version                uint64
	ObservedAt             time.Time
}
type HostEligibility struct {
	HostKey                                   string
	VerificationVersion, CertificationVersion uint64
	VerifiedUntil, CertifiedUntil             time.Time
	Trained, SubanVetted                      bool
}
type Notice struct {
	TemplateKey   string
	Version       uint64
	Locale        string
	AudienceCount int
	Digest        string
}
type Audit struct {
	Sequence                             uint64
	Kind, CommandID, ActorKey, Reference string
	At                                   time.Time
}
type State struct {
	ID, ActorKey       string
	Operation          Operation
	Reason             Reason
	ReasonRef          string
	Density            Density
	Host               *HostEligibility
	Notice             Notice
	NoticeAcknowledged bool
	Revision           uint64
	Audit              []Audit
	AppliedIDs         []string
}
type Proposal struct{ state State }

func Propose(id, actor string, operation Operation, reason Reason, reasonRef string, density Density, host *HostEligibility, notice Notice, command string, at time.Time) (Proposal, error) {
	if !opaque.MatchString(id) || !opaque.MatchString(actor) || !opaque.MatchString(reasonRef) || !token.MatchString(command) ||
		at.IsZero() || !validDensity(density) || !validOperation(operation, reason, host) || !validNotice(notice, density.Participants) ||
		(host != nil && !validHost(*host, at)) {
		return Proposal{}, ErrInvalid
	}
	if host != nil {
		copyHost := *host
		host = &copyHost
	}
	s := State{ID: id, ActorKey: actor, Operation: operation, Reason: reason, ReasonRef: reasonRef, Density: density, Host: host, Notice: notice, Revision: 1, AppliedIDs: []string{command}}
	s.Audit = []Audit{{1, "proposed", command, actor, reasonRef, at.UTC()}}
	return Proposal{s}, nil
}

func (p Proposal) AcknowledgeNotice(actor, digest, command string, at time.Time) (Proposal, error) {
	if p.state.NoticeAcknowledged || actor != p.state.ActorKey || digest != p.state.Notice.Digest || !token.MatchString(command) || at.IsZero() || slices.Contains(p.state.AppliedIDs, command) {
		return Proposal{}, ErrInvalid
	}
	n := p.State()
	n.NoticeAcknowledged = true
	n.Revision++
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Audit = append(n.Audit, Audit{n.Revision, "notice_acknowledged", command, actor, digest, at.UTC()})
	return Proposal{n}, nil
}
func Rehydrate(s State) (Proposal, error) {
	s = clone(s)
	if !validState(s) {
		return Proposal{}, ErrInvalid
	}
	return Proposal{s}, nil
}
func (p Proposal) State() State              { return clone(p.state) }
func (p Proposal) ID() string                { return p.state.ID }
func (p Proposal) Revision() uint64          { return p.state.Revision }
func (p Proposal) ReadyForHumanReview() bool { return p.state.NoticeAcknowledged }

func NoticeDigest(template string, version uint64, locale string, audience int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s:%d", template, version, locale, audience)))
	return hex.EncodeToString(h[:])
}
func validDensity(d Density) bool {
	return opaque.MatchString(d.CircleKey) && opaque.MatchString(d.FireKey) && d.Participants >= 0 && d.Capacity > 0 && d.Capacity <= MaxCapacity && d.Participants <= d.Capacity && d.Version > 0 && !d.ObservedAt.IsZero()
}
func validHost(h HostEligibility, at time.Time) bool {
	return opaque.MatchString(h.HostKey) && h.VerificationVersion > 0 && h.CertificationVersion > 0 && h.Trained && h.SubanVetted && h.VerifiedUntil.After(at) && h.CertifiedUntil.After(at)
}
func validNotice(n Notice, audience int) bool {
	return token.MatchString(n.TemplateKey) && n.Version > 0 && token.MatchString(n.Locale) && n.AudienceCount == audience && opaque.MatchString(n.Digest) && n.Digest == NoticeDigest(n.TemplateKey, n.Version, n.Locale, n.AudienceCount)
}
func validOperation(o Operation, r Reason, h *HostEligibility) bool {
	validReason := map[Reason]bool{ReasonHostUnavailable: true, ReasonCertificationExpired: true, ReasonSafetyCapacity: true, ReasonScheduleConflict: true}[r]
	needsHost := o == AssignHost || o == ReplaceHost
	return validReason && (needsHost || o == CancelFire || o == RescheduleFire) && (h != nil) == needsHost
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.ActorKey) || !opaque.MatchString(s.ReasonRef) || !validDensity(s.Density) || !validOperation(s.Operation, s.Reason, s.Host) || !validNotice(s.Notice, s.Density.Participants) || s.Revision == 0 || len(s.Audit) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	if s.Host != nil && (!opaque.MatchString(s.Host.HostKey) || s.Host.VerificationVersion == 0 || s.Host.CertificationVersion == 0 || !s.Host.Trained || !s.Host.SubanVetted) {
		return false
	}
	for i, a := range s.Audit {
		if a.Sequence != uint64(i+1) || a.CommandID != s.AppliedIDs[i] || !token.MatchString(a.CommandID) || a.At.IsZero() {
			return false
		}
	}
	return true
}
func clone(s State) State {
	s.Audit = append([]Audit(nil), s.Audit...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	if s.Host != nil {
		h := *s.Host
		s.Host = &h
	}
	return s
}
