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

type adminCareStub struct {
	next    func(context.Context, int) ([]safetydomain.CareCase, error)
	engage  func(context.Context, string) (safetydomain.CareCase, error)
	resolve func(context.Context, string, []safetydomain.ScriptKey) (safetydomain.CareCase, error)
}

func (stub adminCareStub) NextOpen(ctx context.Context, limit int) ([]safetydomain.CareCase, error) {
	return stub.next(ctx, limit)
}
func (stub adminCareStub) Engage(ctx context.Context, caseID string) (safetydomain.CareCase, error) {
	return stub.engage(ctx, caseID)
}
func (stub adminCareStub) Resolve(ctx context.Context, caseID string, scripts []safetydomain.ScriptKey) (safetydomain.CareCase, error) {
	return stub.resolve(ctx, caseID, scripts)
}

func careTestCase(status safetydomain.CareStatus, scripts []safetydomain.ScriptKey) safetydomain.CareCase {
	return safetydomain.ReconstituteCareCase(
		"care-1", "member-1", safetydomain.SignalVictimReport, status, scripts,
		2, time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), nil,
	)
}

func careTestService() adminCareStub {
	return adminCareStub{
		next: func(context.Context, int) ([]safetydomain.CareCase, error) {
			return []safetydomain.CareCase{careTestCase(safetydomain.CareOpen, nil)}, nil
		},
		engage: func(context.Context, string) (safetydomain.CareCase, error) {
			return careTestCase(safetydomain.CareEngaged, nil), nil
		},
		resolve: func(_ context.Context, _ string, scripts []safetydomain.ScriptKey) (safetydomain.CareCase, error) {
			return careTestCase(safetydomain.CareResolved, scripts), nil
		},
	}
}

func registerCareTestMux(principal admin.Principal) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterAdminCareRoutes(mux, careTestService(), adminSafetyKeyerStub{}, func(*http.Request) (admin.Principal, error) {
		return principal, nil
	})
	return mux
}

func TestAdminCareQueueRequiresSafetyScopeAndKeysSubject(t *testing.T) {
	mux := registerCareTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeOperations}})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/care/cases", nil)
	request.Header.Set("Authorization", "Bearer admin")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("without scope status=%d body=%s", response.Code, response.Body.String())
	}

	mux = registerCareTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}})
	request = httptest.NewRequest(http.MethodGet, "/v1/admin/care/cases", nil)
	request.Header.Set("Authorization", "Bearer admin")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"subjectRef":"safe:member-1"`) || strings.Contains(body, `"subjectId"`) {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
}

func TestAdminCareResolutionRequiresFreshMFA(t *testing.T) {
	mux := registerCareTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}})
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/care/cases/care-1/resolution", strings.NewReader(`{"scripts":["support_content"]}`))
	request.Header.Set("Authorization", "Bearer admin")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "admin_step_up_required") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}

	mux = registerCareTestMux(admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}, MFAVerified: true})
	request = httptest.NewRequest(http.MethodPost, "/v1/admin/care/cases/care-1/resolution", strings.NewReader(`{"scripts":["support_content"]}`))
	request.Header.Set("Authorization", "Bearer admin")
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"scripts":["support_content"]`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
