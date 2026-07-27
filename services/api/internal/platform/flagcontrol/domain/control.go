package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

const MaxLifetime = 2 * time.Hour

var (
	ErrInvalid   = errors.New("invalid flag control")
	ErrState     = errors.New("invalid flag control state")
	ErrSameActor = errors.New("distinct flag controller required")
	ErrExpired   = errors.New("flag control expired")
	key          = regexp.MustCompile(`^[a-f0-9]{64}$`)
	id           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
)

type Capability string

const (
	CapabilitySow      Capability = "sow"
	CapabilityFires    Capability = "fires"
	CapabilityAI       Capability = "ai"
	CapabilityPayments Capability = "payments"
	CapabilityGate     Capability = "gate"
)

type Environment string

const (
	EnvironmentStaging    Environment = "staging"
	EnvironmentProduction Environment = "production"
)

type Market string

const MarketGH Market = "GH"

type Action string

const (
	ActionEnable  Action = "enable"
	ActionDisable Action = "disable"
	ActionKill    Action = "kill"
	ActionUnkill  Action = "unkill"
)

type Reason string

const (
	ReasonStagedRollout Reason = "staged_rollout"
	ReasonIncident      Reason = "incident"
	ReasonMaintenance   Reason = "maintenance"
)

type Status string

const (
	StatusProposed Status = "proposed"
	StatusApproved Status = "approved"
	StatusApplied  Status = "applied"
	StatusExpired  Status = "expired"
)

type Proposal struct {
	id, commandID, fingerprint, proposerKey, approverKey string
	capability                                           Capability
	environment                                          Environment
	market                                               Market
	action                                               Action
	reason                                               Reason
	status                                               Status
	version                                              uint64
	createdAt, expiresAt                                 time.Time
	approvedAt, appliedAt                                *time.Time
}
type State struct {
	ID, CommandID, Fingerprint, ProposerKey, ApproverKey string
	Capability                                           Capability
	Environment                                          Environment
	Market                                               Market
	Action                                               Action
	Reason                                               Reason
	Status                                               Status
	Version                                              uint64
	CreatedAt, ExpiresAt                                 time.Time
	ApprovedAt, AppliedAt                                *time.Time
}

func NewProposal(idValue, command, proposer string, capability Capability, environment Environment, market Market, action Action, reason Reason, created, expires time.Time) (Proposal, error) {
	p := Proposal{id: idValue, commandID: command, proposerKey: proposer, capability: capability, environment: environment, market: market, action: action, reason: reason, status: StatusProposed, version: 1, createdAt: created.UTC(), expiresAt: expires.UTC()}
	if !valid(p) || !expires.After(created) || expires.Sub(created) > MaxLifetime {
		return Proposal{}, ErrInvalid
	}
	p.fingerprint = termsFingerprint(p)
	return p, nil
}
func Rehydrate(s State) (Proposal, error) {
	p := Proposal{id: s.ID, commandID: s.CommandID, fingerprint: s.Fingerprint, proposerKey: s.ProposerKey, approverKey: s.ApproverKey, capability: s.Capability, environment: s.Environment, market: s.Market, action: s.Action, reason: s.Reason, status: s.Status, version: s.Version, createdAt: s.CreatedAt.UTC(), expiresAt: s.ExpiresAt.UTC(), approvedAt: utcPtr(s.ApprovedAt), appliedAt: utcPtr(s.AppliedAt)}
	if !valid(p) || p.version == 0 || p.fingerprint != termsFingerprint(p) || !validLifecycle(p) {
		return Proposal{}, ErrInvalid
	}
	return p, nil
}
func (p Proposal) Approve(actor string, at time.Time) (Proposal, error) {
	at = at.UTC()
	if p.status != StatusProposed {
		return Proposal{}, ErrState
	}
	if !key.MatchString(actor) {
		return Proposal{}, ErrInvalid
	}
	if actor == p.proposerKey {
		return Proposal{}, ErrSameActor
	}
	if !at.Before(p.expiresAt) {
		return Proposal{}, ErrExpired
	}
	p.approverKey = actor
	p.status = StatusApproved
	p.version++
	p.approvedAt = &at
	return p, nil
}
func (p Proposal) Apply(at time.Time) (Proposal, RuntimeChange, error) {
	at = at.UTC()
	if p.status != StatusApproved {
		return Proposal{}, RuntimeChange{}, ErrState
	}
	if !at.Before(p.expiresAt) {
		return Proposal{}, RuntimeChange{}, ErrExpired
	}
	p.status = StatusApplied
	p.version++
	p.appliedAt = &at
	return p, requestedChange(p), nil
}
func (p Proposal) Expire(at time.Time) (Proposal, RuntimeChange, error) {
	at = at.UTC()
	if p.status == StatusExpired {
		return Proposal{}, RuntimeChange{}, ErrState
	}
	if at.Before(p.expiresAt) {
		return Proposal{}, RuntimeChange{}, ErrState
	}
	p.status = StatusExpired
	p.version++
	return p, FailClosedChange(p), nil
}

type RuntimeChange struct {
	Capability      Capability
	Enabled, Killed bool
}

func requestedChange(p Proposal) RuntimeChange {
	change := RuntimeChange{Capability: p.capability}
	switch p.action {
	case ActionEnable:
		change.Enabled = true
	case ActionDisable:
	case ActionKill:
		change.Killed = true
	case ActionUnkill:
	}
	return change
}
func FailClosedChange(p Proposal) RuntimeChange {
	return RuntimeChange{Capability: p.capability, Enabled: false, Killed: true}
}
func valid(p Proposal) bool {
	return id.MatchString(p.id) && id.MatchString(p.commandID) && key.MatchString(p.proposerKey) && validCapability(p.capability) &&
		(p.environment == EnvironmentStaging || p.environment == EnvironmentProduction) && p.market == MarketGH &&
		(p.action == ActionEnable || p.action == ActionDisable || p.action == ActionKill || p.action == ActionUnkill) &&
		(p.reason == ReasonStagedRollout || p.reason == ReasonIncident || p.reason == ReasonMaintenance) &&
		(p.status == StatusProposed || p.status == StatusApproved || p.status == StatusApplied || p.status == StatusExpired) &&
		!p.createdAt.IsZero() && p.expiresAt.After(p.createdAt) && p.expiresAt.Sub(p.createdAt) <= MaxLifetime
}
func validCapability(c Capability) bool {
	return c == CapabilitySow || c == CapabilityFires || c == CapabilityAI || c == CapabilityPayments || c == CapabilityGate
}
func validLifecycle(p Proposal) bool {
	switch p.status {
	case StatusProposed:
		return p.version == 1 && p.approverKey == "" && p.approvedAt == nil && p.appliedAt == nil
	case StatusApproved:
		return p.version >= 2 && key.MatchString(p.approverKey) && p.approvedAt != nil && p.appliedAt == nil
	case StatusApplied:
		return p.version >= 3 && key.MatchString(p.approverKey) && p.approvedAt != nil && p.appliedAt != nil
	case StatusExpired:
		return p.version >= 2
	}
	return false
}
func termsFingerprint(p Proposal) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{p.commandID, p.proposerKey, string(p.capability), string(p.environment), string(p.market), string(p.action), string(p.reason), p.createdAt.Format(time.RFC3339Nano), p.expiresAt.Format(time.RFC3339Nano)}, "\x00")))
	return hex.EncodeToString(sum[:])
}
func utcPtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	x := v.UTC()
	return &x
}
func (p Proposal) State() State {
	return State{p.id, p.commandID, p.fingerprint, p.proposerKey, p.approverKey, p.capability, p.environment, p.market, p.action, p.reason, p.status, p.version, p.createdAt, p.expiresAt, utcPtr(p.approvedAt), utcPtr(p.appliedAt)}
}
func (p Proposal) ID() string             { return p.id }
func (p Proposal) CommandID() string      { return p.commandID }
func (p Proposal) Fingerprint() string    { return p.fingerprint }
func (p Proposal) Status() Status         { return p.status }
func (p Proposal) Version() uint64        { return p.version }
func (p Proposal) Capability() Capability { return p.capability }
func (p Proposal) ExpiresAt() time.Time   { return p.expiresAt }

type AuditKind string

const (
	AuditProposed AuditKind = "proposed"
	AuditApproved AuditKind = "approved"
	AuditApplied  AuditKind = "applied"
	AuditExpired  AuditKind = "expired"
)

type Audit struct {
	ID, ProposalID, ActorKey string
	Kind                     AuditKind
	Version                  uint64
	At                       time.Time
}

func NewAudit(idValue string, p Proposal, actor string, kind AuditKind, at time.Time) (Audit, error) {
	a := Audit{idValue, p.id, actor, kind, p.version, at.UTC()}
	if !id.MatchString(a.ID) || !id.MatchString(a.ProposalID) || !key.MatchString(a.ActorKey) || (kind != AuditProposed && kind != AuditApproved && kind != AuditApplied && kind != AuditExpired) || a.Version == 0 || a.At.IsZero() {
		return Audit{}, ErrInvalid
	}
	return a, nil
}
