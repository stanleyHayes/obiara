// Package application exposes the least-privilege verification desk boundary.
package application

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Scope string
type Outcome string

const (
	ScopeQueueRead    Scope = "verification.queue.read"
	ScopeEvidenceRead Scope = "verification.evidence.read"
	ScopeReview       Scope = "verification.review"
	ScopeOperations   Scope = "operations.manage"
	ScopeFinance      Scope = "finance.settle"
	ScopeSafety       Scope = "safety.review"

	OutcomeApprove Outcome = "approve"
	OutcomeReject  Outcome = "reject"
)

var (
	ErrForbidden           = errors.New("verification admin action is not permitted")
	ErrMFARequired         = errors.New("evidence access requires recent MFA")
	ErrCaseNotFound        = errors.New("verification case not found")
	ErrCaseClosed          = errors.New("verification case is closed")
	ErrStaleCase           = errors.New("verification case version is stale")
	ErrIdempotencyConflict = errors.New("idempotency key was used for another decision")
	ErrReasonRequired      = errors.New("a bounded decision reason is required")
	ErrPurposeRequired     = errors.New("a bounded evidence purpose is required")
	ErrIdempotencyRequired = errors.New("idempotency key is required")
	ErrInvalidOutcome      = errors.New("invalid verification outcome")
)

var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)

type Principal struct {
	ActorID     string
	Scopes      []Scope
	MFAVerified bool
}

func (principal Principal) Has(scope Scope) bool {
	if strings.TrimSpace(principal.ActorID) == "" {
		return false
	}
	for _, assigned := range principal.Scopes {
		if assigned == scope {
			return true
		}
	}
	return false
}

type CaseSummary struct {
	ID          string
	SubjectRef  string
	ReasonCode  string
	SubmittedAt time.Time
	Version     int64
}

type CaseDetail struct {
	CaseSummary
	Status string
}

type Evidence struct {
	CaseID     string
	MaskedCard string
	AgeBand    string
	// AgeAssuranceMethod and AgeAssuredAt outlive the birth date the band is
	// derived from. Once retention has stripped that date the band reads
	// "unknown", and these are what still show the reviewer that the check
	// happened and how strong it was.
	AgeAssuranceMethod string
	AgeAssuredAt       string
	ProviderStatus     string
}

type DecisionResult struct {
	Case     CaseDetail
	Outcome  Outcome
	Replayed bool
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository, now func() time.Time) Service {
	return Service{repository: repository, now: now}
}

func (service Service) ListQueue(ctx context.Context, principal Principal, limit int) ([]CaseSummary, error) {
	if !principal.Has(ScopeQueueRead) {
		return nil, ErrForbidden
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	return service.repository.ListQueued(ctx, limit)
}

func (service Service) Detail(ctx context.Context, principal Principal, caseID string) (CaseDetail, error) {
	if !principal.Has(ScopeQueueRead) {
		return CaseDetail{}, ErrForbidden
	}
	return service.repository.Detail(ctx, strings.TrimSpace(caseID))
}

func (service Service) OpenEvidence(ctx context.Context, principal Principal, caseID, purpose, reason, correlationID string) (Evidence, error) {
	if !principal.Has(ScopeEvidenceRead) {
		return Evidence{}, ErrForbidden
	}
	if !principal.MFAVerified {
		return Evidence{}, ErrMFARequired
	}
	purpose = strings.TrimSpace(purpose)
	reason = strings.TrimSpace(reason)
	if !codePattern.MatchString(purpose) {
		return Evidence{}, ErrPurposeRequired
	}
	if len(reason) < 8 || len(reason) > 240 {
		return Evidence{}, ErrReasonRequired
	}
	return service.repository.AccessEvidence(ctx, EvidenceAccess{
		CaseID: strings.TrimSpace(caseID), ActorID: principal.ActorID,
		Purpose: purpose, Reason: reason, CorrelationID: correlationID,
		OccurredAt: service.now().UTC(),
	})
}

func (service Service) Decide(ctx context.Context, principal Principal, caseID string, outcome Outcome, reason, idempotencyKey, correlationID string, expectedVersion int64) (DecisionResult, error) {
	if !principal.Has(ScopeReview) {
		return DecisionResult{}, ErrForbidden
	}
	if outcome != OutcomeApprove && outcome != OutcomeReject {
		return DecisionResult{}, ErrInvalidOutcome
	}
	reason = strings.TrimSpace(reason)
	if len(reason) < 8 || len(reason) > 240 {
		return DecisionResult{}, ErrReasonRequired
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || len(idempotencyKey) > 128 {
		return DecisionResult{}, ErrIdempotencyRequired
	}
	if expectedVersion < 1 {
		return DecisionResult{}, ErrStaleCase
	}
	return service.repository.Decide(ctx, DecisionCommand{
		CaseID: strings.TrimSpace(caseID), ActorID: principal.ActorID,
		Outcome: outcome, Reason: reason, IdempotencyKey: idempotencyKey,
		ExpectedVersion: expectedVersion, CorrelationID: correlationID,
		OccurredAt: service.now().UTC(),
	})
}
