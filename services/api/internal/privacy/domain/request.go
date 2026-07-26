// Package domain models data-subject privacy requests (E03-S10, FR-106):
// export within 72 hours, deletion within 30 days with cryptographic
// erasure of voice/biometric blobs, and legal holds that block destruction
// (Doc 09 retention table; agent_plan.md §15).
package domain

import (
	"errors"
	"strings"
	"time"
)

// Kind is the privacy request type.
type Kind string

const (
	KindExport   Kind = "export"
	KindDeletion Kind = "deletion"
)

// Status is the request lifecycle.
type Status string

const (
	StatusRequested  Status = "requested"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusBlocked    Status = "blocked_legal_hold"
)

// Statutory clocks from FR-106.
const (
	ExportDueWithin   = 72 * time.Hour
	DeletionDueWithin = 30 * 24 * time.Hour
)

var (
	ErrRequestIDRequired  = errors.New("privacy request id is required")
	ErrAccountIDRequired  = errors.New("account id is required")
	ErrLegalHoldActive    = errors.New("a legal hold blocks this request")
	ErrRequestNotOpen     = errors.New("privacy request is not open")
	ErrHoldReasonRequired = errors.New("legal hold reason is required")
	ErrHoldActorRequired  = errors.New("legal hold actor is required")
)

// PrivacyRequest is one export or deletion request for an account.
type PrivacyRequest struct {
	id          string
	accountID   string
	kind        Kind
	status      Status
	dueAt       time.Time
	version     int64
	createdAt   time.Time
	completedAt *time.Time
}

// NewRequest opens a request with its statutory due time.
func NewRequest(id, accountID string, kind Kind, now time.Time) (PrivacyRequest, error) {
	if strings.TrimSpace(id) == "" {
		return PrivacyRequest{}, ErrRequestIDRequired
	}
	if strings.TrimSpace(accountID) == "" {
		return PrivacyRequest{}, ErrAccountIDRequired
	}
	now = now.UTC()
	var due time.Time
	switch kind {
	case KindExport:
		due = now.Add(ExportDueWithin)
	case KindDeletion:
		due = now.Add(DeletionDueWithin)
	default:
		return PrivacyRequest{}, errors.New("unknown privacy request kind")
	}
	return PrivacyRequest{id: id, accountID: accountID, kind: kind, status: StatusRequested, dueAt: due, version: 1, createdAt: now}, nil
}

// ReconstituteRequest rebuilds a stored request without policy checks.
func ReconstituteRequest(id, accountID string, kind Kind, status Status, dueAt time.Time, version int64, createdAt time.Time, completedAt *time.Time) PrivacyRequest {
	return PrivacyRequest{id: id, accountID: accountID, kind: kind, status: status, dueAt: dueAt, version: version, createdAt: createdAt, completedAt: completedAt}
}

// Block places the request behind a legal hold.
func (request *PrivacyRequest) Block(now time.Time) error {
	if request.status == StatusCompleted {
		return ErrRequestNotOpen
	}
	request.status = StatusBlocked
	request.version++
	return nil
}

// Unblock returns a held request to the open state.
func (request *PrivacyRequest) Unblock() error {
	if request.status != StatusBlocked {
		return ErrRequestNotOpen
	}
	request.status = StatusRequested
	request.version++
	return nil
}

// StartProcessing marks execution begun (worker picks up the request).
func (request *PrivacyRequest) StartProcessing() error {
	if request.status != StatusRequested {
		return ErrRequestNotOpen
	}
	request.status = StatusProcessing
	request.version++
	return nil
}

// Complete closes the request.
func (request *PrivacyRequest) Complete(now time.Time) error {
	if request.status != StatusProcessing {
		return ErrRequestNotOpen
	}
	request.status = StatusCompleted
	completed := now.UTC()
	request.completedAt = &completed
	request.version++
	return nil
}

// Overdue reports whether the statutory clock has run out.
func (request PrivacyRequest) Overdue(now time.Time) bool {
	return request.status != StatusCompleted && now.UTC().After(request.dueAt)
}

func (request PrivacyRequest) ID() string              { return request.id }
func (request PrivacyRequest) AccountID() string       { return request.accountID }
func (request PrivacyRequest) Kind() Kind              { return request.kind }
func (request PrivacyRequest) Status() Status          { return request.status }
func (request PrivacyRequest) DueAt() time.Time        { return request.dueAt }
func (request PrivacyRequest) Version() int64          { return request.version }
func (request PrivacyRequest) CreatedAt() time.Time    { return request.createdAt }
func (request PrivacyRequest) CompletedAt() *time.Time { return request.completedAt }

// LegalHold preserves an account's data against deletion (Doc 09: T&S
// evidence and legal-hold paths override retention automation).
type LegalHold struct {
	AccountID string
	Reason    string
	PlacedBy  string
	PlacedAt  time.Time
	LiftedAt  *time.Time
}

// NewLegalHold validates a hold placement.
func NewLegalHold(accountID, reason, placedBy string, now time.Time) (LegalHold, error) {
	if strings.TrimSpace(accountID) == "" {
		return LegalHold{}, ErrAccountIDRequired
	}
	if strings.TrimSpace(reason) == "" {
		return LegalHold{}, ErrHoldReasonRequired
	}
	if strings.TrimSpace(placedBy) == "" {
		return LegalHold{}, ErrHoldActorRequired
	}
	return LegalHold{AccountID: accountID, Reason: reason, PlacedBy: placedBy, PlacedAt: now.UTC()}, nil
}

func (hold LegalHold) Active() bool { return hold.LiftedAt == nil }
