// Package domain models host applications and institutional verification.
// Institutional documents remain in evidence storage; this aggregate retains
// only opaque references, keyed issuer/applicant identifiers, and metadata.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	MaxApprovalTerm = 365 * 24 * time.Hour
	RecheckWindow   = 30 * 24 * time.Hour
)

var (
	ErrInvalidApplication = errors.New("invalid host application")
	ErrInvalidTransition  = errors.New("invalid host application transition")
	ErrStaleVersion       = errors.New("stale host application version")
	ErrCommandConflict    = errors.New("host command conflicts with prior use")
	ErrProofExpired       = errors.New("institutional proof expired")
)

type Status string

const (
	StatusSubmitted    Status = "submitted"
	StatusQueuedManual Status = "queued_manual"
	StatusApproved     Status = "approved"
	StatusRejected     Status = "rejected"
	StatusExpired      Status = "expired"
)

type InstitutionKind string

const (
	InstitutionUniversity InstitutionKind = "university"
	InstitutionEmployer   InstitutionKind = "employer"
	InstitutionCommunity  InstitutionKind = "community"
)

type Reason string

const (
	ReasonProviderVerified    Reason = "provider_verified"
	ReasonProviderRejected    Reason = "provider_rejected"
	ReasonProviderUncertain   Reason = "provider_uncertain"
	ReasonProviderUnavailable Reason = "provider_unavailable"
	ReasonManualApproved      Reason = "manual_approved"
	ReasonManualRejected      Reason = "manual_rejected"
	ReasonApprovalExpired     Reason = "approval_expired"
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)

type Proof struct {
	reference string
	kind      InstitutionKind
	issuerKey string
	issuedAt  time.Time
	expiresAt time.Time
}

func NewProof(reference string, kind InstitutionKind, issuerKey string, issuedAt, expiresAt time.Time) (Proof, error) {
	reference = strings.TrimSpace(reference)
	issuerKey = strings.TrimSpace(issuerKey)
	if !opaquePattern.MatchString(reference) || !validDigest(issuerKey) ||
		issuedAt.IsZero() || expiresAt.IsZero() || !expiresAt.After(issuedAt) {
		return Proof{}, ErrInvalidApplication
	}
	switch kind {
	case InstitutionUniversity, InstitutionEmployer, InstitutionCommunity:
	default:
		return Proof{}, ErrInvalidApplication
	}
	return Proof{
		reference: reference, kind: kind, issuerKey: issuerKey,
		issuedAt: issuedAt.UTC(), expiresAt: expiresAt.UTC(),
	}, nil
}

func (proof Proof) Reference() string     { return proof.reference }
func (proof Proof) Kind() InstitutionKind { return proof.kind }
func (proof Proof) IssuerKey() string     { return proof.issuerKey }
func (proof Proof) IssuedAt() time.Time   { return proof.issuedAt }
func (proof Proof) ExpiresAt() time.Time  { return proof.expiresAt }

type AuditEvent struct {
	sequence   uint64
	commandID  string
	action     Status
	reason     Reason
	actorKey   string
	occurredAt time.Time
}

func (event AuditEvent) Sequence() uint64      { return event.sequence }
func (event AuditEvent) CommandID() string     { return event.commandID }
func (event AuditEvent) Action() Status        { return event.action }
func (event AuditEvent) Reason() Reason        { return event.reason }
func (event AuditEvent) ActorKey() string      { return event.actorKey }
func (event AuditEvent) OccurredAt() time.Time { return event.occurredAt }

type Command struct {
	ID       string
	ActorKey string
	At       time.Time
}

type Application struct {
	id            string
	submissionID  string
	applicantKey  string
	proof         Proof
	status        Status
	reason        Reason
	providerRef   string
	approvedUntil time.Time
	recheckDueAt  time.Time
	createdAt     time.Time
	updatedAt     time.Time
	version       uint64
	audit         []AuditEvent
}

func NewApplication(id, submissionID, applicantKey string, proof Proof, command Command) (Application, error) {
	id = strings.TrimSpace(id)
	submissionID = strings.TrimSpace(submissionID)
	applicantKey = strings.TrimSpace(applicantKey)
	if !opaquePattern.MatchString(id) || !opaquePattern.MatchString(submissionID) ||
		!validDigest(applicantKey) || proof.reference == "" || command.At.IsZero() ||
		command.At.UTC().Before(proof.issuedAt) || !command.At.UTC().Before(proof.expiresAt) {
		return Application{}, ErrInvalidApplication
	}
	event, err := newAudit(1, command, StatusSubmitted, "")
	if err != nil {
		return Application{}, err
	}
	return Application{
		id: id, submissionID: submissionID, applicantKey: applicantKey, proof: proof,
		status: StatusSubmitted, createdAt: command.At.UTC(), updatedAt: command.At.UTC(),
		version: 1, audit: []AuditEvent{event},
	}, nil
}

func Rehydrate(id, submissionID, applicantKey string, proof Proof, status Status, reason Reason, providerRef string, approvedUntil, recheckDueAt, createdAt, updatedAt time.Time, version uint64, audit []AuditEvent) (Application, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(id)) ||
		!opaquePattern.MatchString(strings.TrimSpace(submissionID)) ||
		!validDigest(strings.TrimSpace(applicantKey)) || proof.reference == "" ||
		version == 0 || uint64(len(audit)) != version ||
		createdAt.IsZero() || updatedAt.Before(createdAt) {
		return Application{}, ErrInvalidApplication
	}
	for index, event := range audit {
		if event.sequence != uint64(index+1) || !opaquePattern.MatchString(event.commandID) ||
			!validActor(event.actorKey) || event.occurredAt.IsZero() ||
			(index > 0 && event.occurredAt.Before(audit[index-1].occurredAt)) {
			return Application{}, ErrInvalidApplication
		}
	}
	if audit[len(audit)-1].action != status {
		return Application{}, ErrInvalidApplication
	}
	switch status {
	case StatusSubmitted, StatusQueuedManual, StatusRejected, StatusExpired:
	case StatusApproved:
		if approvedUntil.IsZero() || recheckDueAt.IsZero() || approvedUntil.After(proof.expiresAt) {
			return Application{}, ErrInvalidApplication
		}
	default:
		return Application{}, ErrInvalidApplication
	}
	return Application{
		id: id, submissionID: submissionID, applicantKey: applicantKey, proof: proof,
		status: status, reason: reason, providerRef: providerRef,
		approvedUntil: approvedUntil.UTC(), recheckDueAt: recheckDueAt.UTC(),
		createdAt: createdAt.UTC(), updatedAt: updatedAt.UTC(), version: version,
		audit: append([]AuditEvent(nil), audit...),
	}, nil
}

func NewAuditEvent(sequence uint64, commandID string, action Status, reason Reason, actorKey string, occurredAt time.Time) (AuditEvent, error) {
	return newAudit(sequence, Command{ID: commandID, ActorKey: actorKey, At: occurredAt}, action, reason)
}

func (application Application) QueueManual(reason Reason, providerRef string, command Command, expected uint64) (Application, error) {
	if reason != ReasonProviderUncertain && reason != ReasonProviderUnavailable {
		return Application{}, ErrInvalidTransition
	}
	if application.status != StatusSubmitted && application.status != StatusApproved {
		return Application{}, ErrInvalidTransition
	}
	return application.transition(StatusQueuedManual, reason, providerRef, command, expected)
}

func (application Application) ProviderDecision(approved bool, providerRef string, command Command, expected uint64) (Application, error) {
	if application.status != StatusSubmitted && application.status != StatusApproved {
		return Application{}, ErrInvalidTransition
	}
	if !opaquePattern.MatchString(strings.TrimSpace(providerRef)) {
		return Application{}, ErrInvalidApplication
	}
	if approved {
		return application.approve(ReasonProviderVerified, providerRef, command, expected)
	}
	return application.transition(StatusRejected, ReasonProviderRejected, providerRef, command, expected)
}

func (application Application) ManualDecision(approved bool, command Command, expected uint64) (Application, error) {
	if application.status != StatusQueuedManual {
		return Application{}, ErrInvalidTransition
	}
	if approved {
		return application.approve(ReasonManualApproved, "", command, expected)
	}
	return application.transition(StatusRejected, ReasonManualRejected, "", command, expected)
}

func (application Application) Expire(command Command, expected uint64) (Application, error) {
	if application.status != StatusApproved || command.At.UTC().Before(application.approvedUntil) {
		return Application{}, ErrInvalidTransition
	}
	return application.transition(StatusExpired, ReasonApprovalExpired, "", command, expected)
}

func (application Application) approve(reason Reason, providerRef string, command Command, expected uint64) (Application, error) {
	if !command.At.UTC().Before(application.proof.expiresAt) {
		return Application{}, ErrProofExpired
	}
	next, err := application.transition(StatusApproved, reason, providerRef, command, expected)
	if err != nil {
		return Application{}, err
	}
	next.approvedUntil = command.At.UTC().Add(MaxApprovalTerm)
	if application.proof.expiresAt.Before(next.approvedUntil) {
		next.approvedUntil = application.proof.expiresAt
	}
	next.recheckDueAt = next.approvedUntil.Add(-RecheckWindow)
	if next.recheckDueAt.Before(command.At.UTC()) {
		next.recheckDueAt = command.At.UTC()
	}
	return next, nil
}

func (application Application) transition(status Status, reason Reason, providerRef string, command Command, expected uint64) (Application, error) {
	if event, exists := application.command(command.ID); exists {
		if event.action == status && event.reason == reason && event.actorKey == command.ActorKey {
			return application, nil
		}
		return Application{}, ErrCommandConflict
	}
	if expected != application.version {
		return Application{}, ErrStaleVersion
	}
	event, err := newAudit(application.version+1, command, status, reason)
	if err != nil {
		return Application{}, err
	}
	next := application
	next.status, next.reason, next.providerRef = status, reason, strings.TrimSpace(providerRef)
	next.updatedAt, next.version = command.At.UTC(), application.version+1
	next.audit = append(application.Audit(), event)
	return next, nil
}

func newAudit(sequence uint64, command Command, action Status, reason Reason) (AuditEvent, error) {
	command.ID, command.ActorKey = strings.TrimSpace(command.ID), strings.TrimSpace(command.ActorKey)
	if sequence == 0 || !opaquePattern.MatchString(command.ID) || !validActor(command.ActorKey) || command.At.IsZero() {
		return AuditEvent{}, ErrInvalidApplication
	}
	return AuditEvent{
		sequence: sequence, commandID: command.ID, action: action, reason: reason,
		actorKey: command.ActorKey, occurredAt: command.At.UTC(),
	}, nil
}

func (application Application) command(id string) (AuditEvent, bool) {
	for _, event := range application.audit {
		if event.commandID == strings.TrimSpace(id) {
			return event, true
		}
	}
	return AuditEvent{}, false
}

func (application Application) HasCommand(id string) bool {
	_, exists := application.command(id)
	return exists
}

func (application Application) MatchesCommand(id string, status Status, reason Reason, actorKey string) bool {
	event, exists := application.command(id)
	return exists && event.action == status && event.reason == reason &&
		event.actorKey == strings.TrimSpace(actorKey)
}

func (application Application) SameCommand(other Application, id string) bool {
	left, leftExists := application.command(id)
	right, rightExists := other.command(id)
	return leftExists && rightExists && left.action == right.action &&
		left.reason == right.reason && left.actorKey == right.actorKey
}

func (application Application) ID() string               { return application.id }
func (application Application) SubmissionID() string     { return application.submissionID }
func (application Application) ApplicantKey() string     { return application.applicantKey }
func (application Application) Proof() Proof             { return application.proof }
func (application Application) Status() Status           { return application.status }
func (application Application) Reason() Reason           { return application.reason }
func (application Application) ProviderRef() string      { return application.providerRef }
func (application Application) ApprovedUntil() time.Time { return application.approvedUntil }
func (application Application) RecheckDueAt() time.Time  { return application.recheckDueAt }
func (application Application) CreatedAt() time.Time     { return application.createdAt }
func (application Application) UpdatedAt() time.Time     { return application.updatedAt }
func (application Application) Version() uint64          { return application.version }
func (application Application) Audit() []AuditEvent {
	return append([]AuditEvent(nil), application.audit...)
}
func (application Application) Eligible(at time.Time) bool {
	return application.status == StatusApproved && at.UTC().Before(application.approvedUntil)
}

func validActor(value string) bool { return value == "system" || validDigest(value) }
func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
