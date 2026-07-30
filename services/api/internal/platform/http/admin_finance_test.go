package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	reconciliationapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/application"
	reconciliationdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type financeOverviewStub struct{ overview reconciliationapp.Overview }

func (stub financeOverviewStub) Overview(context.Context, int) (reconciliationapp.Overview, error) {
	return stub.overview, nil
}

func TestAdminFinanceOverviewRequiresFinanceScopeAndBoundsKeys(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	rawKey := strings.Repeat("a", 64)
	for name, testCase := range map[string]struct {
		scopes []admin.Scope
		want   int
	}{
		"forbidden": {[]admin.Scope{admin.ScopeOperations}, http.StatusForbidden},
		"finance":   {[]admin.Scope{admin.ScopeFinance}, http.StatusOK},
	} {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			RegisterAdminFinanceRoutes(mux, financeOverviewStub{overview: reconciliationapp.Overview{
				Exceptions: []reconciliationapp.ExceptionView{{
					FactID: "fact-secret", ProviderKey: rawKey, ReferenceKey: rawKey,
					Currency: reconciliationdomain.CurrencyGHS, Minor: 12000,
					Outcome:   reconciliationdomain.OutcomeException,
					Exception: reconciliationdomain.ExceptionAmount, OccurredAt: now, RecordedAt: now,
				}},
			}}, func(*http.Request) (admin.Principal, error) {
				return admin.Principal{ActorID: "finance-1", Scopes: testCase.scopes}, nil
			})
			request := httptest.NewRequest(http.MethodGet, "/v1/admin/finance/reconciliation", nil)
			request.Header.Set("Authorization", "Bearer session")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != testCase.want {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if response.Code == http.StatusOK && strings.Contains(response.Body.String(), rawKey) {
				t.Fatalf("full privacy key leaked: %s", response.Body.String())
			}
		})
	}
}
