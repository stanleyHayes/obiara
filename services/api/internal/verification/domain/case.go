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
	ErrCardKeyRequired     = errors.New("ghana card key is required")
	ErrCaseNotOpen         = errors.New("verification case is not open")
	ErrDecisionReasonEmpty = errors.New("decision reason is required")
	// ErrAgeAssuranceRequired refuses a case that does not record how its
	// minimum-age determination was made. A case exists only because the age
	// gate let it, and that fact has to be written down rather than implied
	// by the case's own existence.
	ErrAgeAssuranceRequired = errors.New("age assurance is required")
)

// AgeAssuranceMethod names how the birth date behind a determination was
// obtained. It is deliberately not a boolean: "we checked a date" and "the
// issuer confirmed the date we checked" are different claims, and an audit
// trail that cannot tell them apart overstates the weaker one.
type AgeAssuranceMethod string

const (
	// AgeSelfDeclared is a date typed by the person being checked. Every
	// submission starts here, before any provider has looked at it.
	AgeSelfDeclared AgeAssuranceMethod = "self_declared_dob"
	// AgeIssuerCorroborated is a self-declared date the issuer then matched
	// against its own records.
	AgeIssuerCorroborated AgeAssuranceMethod = "issuer_corroborated_dob"
)

// AgeAssurance is the durable proof that a minimum-age determination was made.
//
// It carries no birth date on purpose. Retention strips the date thirty days
// after a decision, and the whole point of this record is to outlive it: after
// the strip, this is the only thing that still says an age check happened, by
// what method, and against what threshold.
type AgeAssurance struct {
	AssuredAt  time.Time
	Method     AgeAssuranceMethod
	MinimumAge int
}

// Recorded reports whether the assurance is complete enough to be proof.
func (assurance AgeAssurance) Recorded() bool {
	return !assurance.AssuredAt.IsZero() && assurance.Method != "" && assurance.MinimumAge > 0
}

// VerificationCase is one identity-verification attempt for an account.
type VerificationCase struct {
	id        string
	accountID string
	// cardKey is the HMAC of the card number, never the number itself. It
	// exists so two submissions of one card can be recognised as one
	// identity without the platform retaining a national ID.
	cardKey string
	// cardMask is the last four digits, which is all the review desk ever
	// displayed.
	cardMask    string
	status      CaseStatus
	providerRef string
	reason      string
	dateOfBirth time.Time
	// ageAssurance survives dateOfBirth being stripped by retention.
	ageAssurance AgeAssurance
	version      int64
	createdAt    time.Time
	decidedAt    *time.Time
}

// NewCase opens a pending case for a Ghana Card submission.
//
// It takes the key and mask rather than the card number: the plaintext is
// needed only to ask the provider, within the request that carries it, and
// this package's contract is that raw card artifacts never persist.
func NewCase(id, accountID, cardKey, cardMask string, dateOfBirth time.Time, assurance AgeAssurance, now time.Time) (VerificationCase, error) {
	if strings.TrimSpace(id) == "" {
		return VerificationCase{}, ErrCaseIDRequired
	}
	if strings.TrimSpace(accountID) == "" {
		return VerificationCase{}, ErrAccountIDRequired
	}
	if strings.TrimSpace(cardKey) == "" {
		return VerificationCase{}, ErrCardKeyRequired
	}
	if !assurance.Recorded() {
		return VerificationCase{}, ErrAgeAssuranceRequired
	}
	return VerificationCase{
		id:           id,
		accountID:    accountID,
		cardKey:      cardKey,
		cardMask:     cardMask,
		status:       StatusPending,
		dateOfBirth:  dateOfBirth.UTC(),
		ageAssurance: assurance,
		version:      1,
		createdAt:    now.UTC(),
	}, nil
}

// AgeAssurance reports how this case's minimum-age determination was made.
func (c VerificationCase) AgeAssurance() AgeAssurance { return c.ageAssurance }

// WithAgeAssurance restores a stored assurance during reconstitution. It is
// separate from ReconstituteCase so that adding this record did not have to
// change every caller of a reconstitutor shared with other contexts.
func (c VerificationCase) WithAgeAssurance(assurance AgeAssurance) VerificationCase {
	c.ageAssurance = assurance
	return c
}

// CorroborateAge upgrades a self-declared date to one the issuer matched.
//
// The assured-at time is not moved: the determination was made when the case
// opened, and this records only that a second source later agreed with it.
func (c *VerificationCase) CorroborateAge() {
	if c.ageAssurance.Method == AgeSelfDeclared {
		c.ageAssurance.Method = AgeIssuerCorroborated
	}
}

// ReconstituteCase rebuilds a stored case without policy checks.
func ReconstituteCase(id, accountID, cardKey, cardMask string, status CaseStatus, providerRef, reason string, dateOfBirth time.Time, version int64, createdAt time.Time, decidedAt *time.Time) VerificationCase {
	return VerificationCase{
		id:          id,
		accountID:   accountID,
		cardKey:     cardKey,
		cardMask:    cardMask,
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
func (c VerificationCase) CardKey() string        { return c.cardKey }
func (c VerificationCase) CardMask() string       { return c.cardMask }
func (c VerificationCase) Status() CaseStatus     { return c.status }
func (c VerificationCase) ProviderRef() string    { return c.providerRef }
func (c VerificationCase) Reason() string         { return c.reason }
func (c VerificationCase) DateOfBirth() time.Time { return c.dateOfBirth }
func (c VerificationCase) Version() int64         { return c.version }
func (c VerificationCase) CreatedAt() time.Time   { return c.createdAt }
func (c VerificationCase) DecidedAt() *time.Time  { return c.decidedAt }
