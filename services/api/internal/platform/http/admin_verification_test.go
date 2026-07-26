package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type adminVerificationStub struct {
	evidenceAccess admin.EvidenceAccess
	decision       admin.DecisionCommand
}

func (stub *adminVerificationStub) ListQueue(_ context.Context, _ admin.Principal, _ int) ([]admin.CaseSummary, error) {
	return []admin.CaseSummary{{
		ID: "IDV-1", SubjectRef: "member_a1b2", ReasonCode: "provider_uncertain",
		SubmittedAt: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC), Version: 2,
	}}, nil
}

func (stub *adminVerificationStub) Detail(_ context.Context, _ admin.Principal, _ string) (admin.CaseDetail, error) {
	return admin.CaseDetail{}, nil
}

func (stub *adminVerificationStub) OpenEvidence(_ context.Context, principal admin.Principal, caseID, purpose, reason, correlationID string) (admin.Evidence, error) {
	stub.evidenceAccess = admin.EvidenceAccess{
		CaseID: caseID, ActorID: principal.ActorID, Purpose: purpose,
		Reason: reason, CorrelationID: correlationID,
	}
	return admin.Evidence{CaseID: caseID, MaskedCard: "•••• 1234", AgeBand: "25_34", ProviderStatus: "provider_uncertain"}, nil
}

func (stub *adminVerificationStub) Decide(_ context.Context, principal admin.Principal, caseID string, outcome admin.Outcome, reason, key, correlationID string, version int64) (admin.DecisionResult, error) {
	stub.decision = admin.DecisionCommand{
		CaseID: caseID, ActorID: principal.ActorID, Outcome: outcome, Reason: reason,
		IdempotencyKey: key, ExpectedVersion: version, CorrelationID: correlationID,
	}
	return admin.DecisionResult{Case: admin.CaseDetail{CaseSummary: admin.CaseSummary{ID: caseID, Version: version + 1}, Status: "approved"}, Outcome: outcome}, nil
}

func verifierPrincipal(*http.Request) (admin.Principal, error) {
	return admin.Principal{
		ActorID: "verifier-1", MFAVerified: true,
		Scopes: []admin.Scope{admin.ScopeQueueRead, admin.ScopeEvidenceRead, admin.ScopeReview},
	}, nil
}

func TestAdminQueueContainsOnlyRedactedContractFields(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAdminVerificationRoutes(mux, &adminVerificationStub{}, verifierPrincipal)
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/verifications?limit=10", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, forbidden := range []string{"cardNumber", "dateOfBirth", "providerRef", "accountId"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("queue leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"subjectRef":"member_a1b2"`) {
		t.Fatalf("redacted subject missing: %s", body)
	}
}

func TestEvidenceAccessAndDecisionForwardDedicatedAuditInputs(t *testing.T) {
	stub := &adminVerificationStub{}
	mux := http.NewServeMux()
	RegisterAdminVerificationRoutes(mux, stub, verifierPrincipal)

	evidenceRequest := httptest.NewRequest(http.MethodPost, "/v1/admin/verifications/IDV-1/evidence-access", strings.NewReader(`{"purpose":"identity_review","reason":"Provider result needs human comparison"}`))
	evidenceRequest.Header.Set("Content-Type", "application/json")
	evidenceRequest.Header.Set("X-Correlation-ID", "corr-1")
	evidenceResponse := httptest.NewRecorder()
	mux.ServeHTTP(evidenceResponse, evidenceRequest)
	if evidenceResponse.Code != http.StatusOK || stub.evidenceAccess.Purpose != "identity_review" || stub.evidenceAccess.ActorID != "verifier-1" {
		t.Fatalf("evidence status=%d access=%+v", evidenceResponse.Code, stub.evidenceAccess)
	}

	decisionRequest := httptest.NewRequest(http.MethodPost, "/v1/admin/verifications/IDV-1/decisions", strings.NewReader(`{"outcome":"approve","reason":"Document and provider result are consistent","expectedVersion":2}`))
	decisionRequest.Header.Set("Content-Type", "application/json")
	decisionRequest.Header.Set("Idempotency-Key", "command-1")
	decisionResponse := httptest.NewRecorder()
	mux.ServeHTTP(decisionResponse, decisionRequest)
	if decisionResponse.Code != http.StatusOK || stub.decision.IdempotencyKey != "command-1" || stub.decision.ExpectedVersion != 2 {
		t.Fatalf("decision status=%d command=%+v", decisionResponse.Code, stub.decision)
	}
}

func TestAdminResolverDenialFailsClosed(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAdminVerificationRoutes(mux, &adminVerificationStub{}, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{}, errors.New("no session")
	})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/admin/verifications", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Error APIError `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Error.Code != "forbidden" {
		t.Fatalf("error=%+v decode=%v", envelope.Error, err)
	}
}
