package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	safetydomain "github.com/stanleyHayes/obiara/internal/safety/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type adminSafetyCasesStub struct {
	next   func(context.Context, safetydomain.Queue, int) ([]safetydomain.Case, error)
	find   func(context.Context, string) (safetydomain.Case, error)
	assign func(context.Context, string, string) (safetydomain.Case, error)
}

func (stub adminSafetyCasesStub) NextQueued(ctx context.Context, queue safetydomain.Queue, limit int) ([]safetydomain.Case, error) {
	return stub.next(ctx, queue, limit)
}
func (stub adminSafetyCasesStub) Find(ctx context.Context, caseID string) (safetydomain.Case, error) {
	return stub.find(ctx, caseID)
}
func (stub adminSafetyCasesStub) Assign(ctx context.Context, caseID, actorID string) (safetydomain.Case, error) {
	return stub.assign(ctx, caseID, actorID)
}

type adminSafetyEvidenceStub struct {
	view func(context.Context, string, string, safetydomain.Purpose) (safetydomain.Bundle, error)
}

func (stub adminSafetyEvidenceStub) View(ctx context.Context, caseID, actorID string, purpose safetydomain.Purpose) (safetydomain.Bundle, error) {
	return stub.view(ctx, caseID, actorID, purpose)
}

type adminSafetyKeyerStub struct{}

func (adminSafetyKeyerStub) MemberKey(memberID string) (string, error) {
	return "safe:" + memberID, nil
}

func safetyTestCase(assignedTo string) safetydomain.Case {
	return safetydomain.ReconstituteCase(
		"case-1", "report-1", "member-subject", safetydomain.TierB,
		safetydomain.QueueTriage, time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC),
		map[bool]safetydomain.CaseStatus{true: safetydomain.CaseInReview, false: safetydomain.CaseQueued}[assignedTo != ""],
		assignedTo, 2, time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), nil,
	)
}

func safetyTestCases(value safetydomain.Case) adminSafetyCasesStub {
	return adminSafetyCasesStub{
		next: func(context.Context, safetydomain.Queue, int) ([]safetydomain.Case, error) {
			return []safetydomain.Case{value}, nil
		},
		find: func(context.Context, string) (safetydomain.Case, error) { return value, nil },
		assign: func(context.Context, string, string) (safetydomain.Case, error) {
			return safetyTestCase("agent-1"), nil
		},
	}
}

func safetyTestEvidence() adminSafetyEvidenceStub {
	return adminSafetyEvidenceStub{view: func(_ context.Context, caseID, actorID string, purpose safetydomain.Purpose) (safetydomain.Bundle, error) {
		return safetydomain.Bundle{
			CaseID: caseID, Tier: safetydomain.TierB, Category: safetydomain.CategoryHarassment,
			Surface: safetydomain.SurfaceRoom, ContextRef: "room-1", SubjectID: "member-subject",
			Description: "Contact [redacted]",
		}, nil
	}}
}

func registerSafetyTestMux(principal admin.Principal, value safetydomain.Case) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterAdminSafetyRoutes(mux, safetyTestCases(value), safetyTestEvidence(), adminSafetyKeyerStub{}, func(*http.Request) (admin.Principal, error) {
		return principal, nil
	})
	return mux
}

func TestAdminSafetyQueueRequiresSafetyScopeAndKeysSubjects(t *testing.T) {
	mux := registerSafetyTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeOperations}}, safetyTestCase(""))
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/safety/cases", nil)
	request.Header.Set("Authorization", "Bearer admin")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("without scope status=%d body=%s", response.Code, response.Body.String())
	}

	mux = registerSafetyTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}}, safetyTestCase(""))
	request = httptest.NewRequest(http.MethodGet, "/v1/admin/safety/cases", nil)
	request.Header.Set("Authorization", "Bearer admin")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"subjectRef":"safe:member-subject"`) {
		t.Fatalf("with scope status=%d body=%s", response.Code, body)
	}
	if strings.Contains(body, `"subjectId"`) {
		t.Fatalf("raw subject field leaked: %s", body)
	}
	if strings.Contains(body, "reporter") {
		t.Fatalf("reporter field leaked: %s", body)
	}
}

func TestAdminSafetyEvidenceRequiresFreshMFAAndAssignment(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		principal  admin.Principal
		assignedTo string
		wantStatus int
		wantCode   string
	}{
		{"fresh MFA", admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}}, "agent-1", http.StatusForbidden, "admin_step_up_required"},
		{"assignment", admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}, MFAVerified: true}, "agent-2", http.StatusForbidden, "safety_assignment_required"},
		{"authorized", admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}, MFAVerified: true}, "agent-1", http.StatusOK, `"description":"Contact [redacted]"`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			mux := registerSafetyTestMux(testCase.principal, safetyTestCase(testCase.assignedTo))
			request := httptest.NewRequest(http.MethodPost, "/v1/admin/safety/cases/case-1/evidence-access", strings.NewReader(`{"purpose":"triage"}`))
			request.Header.Set("Authorization", "Bearer admin")
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != testCase.wantStatus || !strings.Contains(response.Body.String(), testCase.wantCode) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
