// Package domain models identity verification cases (E03-S03). A case is
// created per Ghana Card submission and decided by the provider adapter or
// the human fallback queue. Raw card artifacts never persist here — the
// case holds the provider reference and decision only (data-classification
// C4: minimal retained proof set).
package domain

import (
	"errors"
	"strings"
	"time"
)

// CaseStatus is the verification case lifecycle.
type CaseStatus string

const (
	StatusPending      CaseStatus = "pending"
	StatusApproved     CaseStatus = "approved"
	StatusRejected     CaseStatus = "rejected"
	StatusQueuedManual CaseStatus = "queued_manual"
)

var (
	ErrCaseIDRequired      = errors.New("verification case id is required")
	ErrAccountIDRequired   = errors.New("account id is required")
	ErrCardNumberRequired  = errors.New("ghana card number is required")
	ErrCaseNotOpen         = errors.New("verification case is not open")
	ErrDecisionReasonEmpty = errors.New("decision reason is required")
)

// VerificationCase is one identity-verification attempt for an account.
type VerificationCase struct {
	id          string
	accountID   string
	cardNumber  string
	status      CaseStatus
	providerRef string
	reason      string
	dateOfBirth time.Time
	version     int64
	createdAt   time.Time
	decidedAt   *time.Time
}

// NewCase opens a pending case for a Ghana Card submission.
func NewCase(id, accountID, cardNumber string, dateOfBirth time.Time, now time.Time) (VerificationCase, error) {
	if strings.TrimSpace(id) == "" {
		return VerificationCase{}, ErrCaseIDRequired
	}
	if strings.TrimSpace(accountID) == "" {
		return VerificationCase{}, ErrAccountIDRequired
	}
	if strings.TrimSpace(cardNumber) == "" {
		return VerificationCase{}, ErrCardNumberRequired
	}
	return VerificationCase{
		id:          id,
		accountID:   accountID,
		cardNumber:  cardNumber,
		status:      StatusPending,
		dateOfBirth: dateOfBirth.UTC(),
		version:     1,
		createdAt:   now.UTC(),
	}, nil
}

// ReconstituteCase rebuilds a stored case without policy checks.
func ReconstituteCase(id, accountID, cardNumber string, status CaseStatus, providerRef, reason string, dateOfBirth time.Time, version int64, createdAt time.Time, decidedAt *time.Time) VerificationCase {
	return VerificationCase{
		id:          id,
		accountID:   accountID,
		cardNumber:  cardNumber,
		status:      status,
		providerRef: providerRef,
		reason:      reason,
		dateOfBirth: dateOfBirth,
		version:     version,
		createdAt:   createdAt,
		decidedAt:   decidedAt,
	}
}

func (c *VerificationCase) decide(status CaseStatus, reason string, now time.Time) error {
	if c.status != StatusPending && c.status != StatusQueuedManual {
		return ErrCaseNotOpen
	}
	if strings.TrimSpace(reason) == "" {
		return ErrDecisionReasonEmpty
	}
	c.status = status
	c.reason = reason
	decided := now.UTC()
	c.decidedAt = &decided
	c.version++
	return nil
}

// Approve records a provider or human approval.
func (c *VerificationCase) Approve(providerRef, reason string, now time.Time) error {
	if err := c.decide(StatusApproved, reason, now); err != nil {
		return err
	}
	c.providerRef = providerRef
	return nil
}

// Reject records a provider or human rejection.
func (c *VerificationCase) Reject(providerRef, reason string, now time.Time) error {
	if err := c.decide(StatusRejected, reason, now); err != nil {
		return err
	}
	c.providerRef = providerRef
	return nil
}

// QueueForManualReview routes the case to the human fallback queue
// (provider unavailable or uncertain — never a silent pass, FR-103).
func (c *VerificationCase) QueueForManualReview(reason string, now time.Time) error {
	if c.status != StatusPending {
		return ErrCaseNotOpen
	}
	if strings.TrimSpace(reason) == "" {
		return ErrDecisionReasonEmpty
	}
	c.status = StatusQueuedManual
	c.reason = reason
	c.version++
	return nil
}

func (c VerificationCase) ID() string             { return c.id }
func (c VerificationCase) AccountID() string      { return c.accountID }
func (c VerificationCase) CardNumber() string     { return c.cardNumber }
func (c VerificationCase) Status() CaseStatus     { return c.status }
func (c VerificationCase) ProviderRef() string    { return c.providerRef }
func (c VerificationCase) Reason() string         { return c.reason }
func (c VerificationCase) DateOfBirth() time.Time { return c.dateOfBirth }
func (c VerificationCase) Version() int64         { return c.version }
func (c VerificationCase) CreatedAt() time.Time   { return c.createdAt }
func (c VerificationCase) DecidedAt() *time.Time  { return c.decidedAt }
