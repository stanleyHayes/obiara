package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestLeastPrivilegeScopesAreIndependent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	service := NewService(repository, time.Now)
	queueOnly := Principal{ActorID: "verifier-1", Scopes: []Scope{ScopeQueueRead}}
	repository.EXPECT().ListQueued(gomock.Any(), 25).Return([]CaseSummary{{ID: "case-1"}}, nil)
	if _, err := service.ListQueue(context.Background(), queueOnly, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.OpenEvidence(context.Background(), queueOnly, "case-1", "identity_review", "Reviewing provider mismatch", "corr-1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("evidence without scope = %v", err)
	}
	if _, err := service.Decide(context.Background(), queueOnly, "case-1", OutcomeApprove, "Evidence is consistent", "command-1", "corr-1", 2); !errors.Is(err, ErrForbidden) {
		t.Fatalf("decision without scope = %v", err)
	}
}

func TestEvidenceRequiresMFAAndCreatesDedicatedAccessCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, func() time.Time { return now })
	principal := Principal{ActorID: "verifier-1", Scopes: []Scope{ScopeEvidenceRead}}
	if _, err := service.OpenEvidence(context.Background(), principal, "case-1", "identity_review", "Reviewing provider mismatch", "corr-1"); !errors.Is(err, ErrMFARequired) {
		t.Fatalf("evidence without MFA = %v", err)
	}
	principal.MFAVerified = true
	repository.EXPECT().AccessEvidence(gomock.Any(), EvidenceAccess{
		CaseID: "case-1", ActorID: "verifier-1", Purpose: "identity_review",
		Reason: "Reviewing provider mismatch", CorrelationID: "corr-1", OccurredAt: now,
	}).Return(Evidence{CaseID: "case-1", MaskedCard: "•••• 1234"}, nil)
	if _, err := service.OpenEvidence(context.Background(), principal, "case-1", "identity_review", "Reviewing provider mismatch", "corr-1"); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionRequiresReasonVersionAndIdempotency(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, func() time.Time { return now })
	principal := Principal{ActorID: "verifier-1", Scopes: []Scope{ScopeReview}}
	if _, err := service.Decide(context.Background(), principal, "case-1", OutcomeApprove, "short", "command-1", "corr-1", 1); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("short reason = %v", err)
	}
	repository.EXPECT().Decide(gomock.Any(), DecisionCommand{
		CaseID: "case-1", ActorID: "verifier-1", Outcome: OutcomeReject,
		Reason: "Document details do not match", IdempotencyKey: "command-1",
		ExpectedVersion: 2, CorrelationID: "corr-1", OccurredAt: now,
	}).Return(DecisionResult{Outcome: OutcomeReject}, nil)
	if _, err := service.Decide(context.Background(), principal, "case-1", OutcomeReject, "Document details do not match", "command-1", "corr-1", 2); err != nil {
		t.Fatal(err)
	}
}
