package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	adminapp "github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

type adminAccountAuthenticatorStub struct {
	session   admindomain.Session
	principal admindomain.Principal
	err       error
}

func (stub adminAccountAuthenticatorStub) Authenticate(context.Context, string) (admindomain.Session, admindomain.Principal, error) {
	return stub.session, stub.principal, stub.err
}

func TestAdminAccountReturnsOnlyCurrentAuthenticatedFacts(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	principal, err := admindomain.NewPrincipal("principal-secret", "operator@obiara.com", []admindomain.Role{admindomain.RoleFinance}, now)
	if err != nil {
		t.Fatal(err)
	}
	session := admindomain.NewSession("session-secret", principal.ID(), principal.Roles(), now)
	handler := adminAccountHandler(adminAccountAuthenticatorStub{session: session, principal: principal})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/account", nil)
	request.Header.Set("Authorization", "Bearer session-secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, forbidden := range []string{"principal-secret", "session-secret", "device", "location"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "operator@obiara.com") || !strings.Contains(body, `"finance"`) {
		t.Fatalf("account facts missing: %s", body)
	}
}

func TestAdminAccountRejectsInvalidSession(t *testing.T) {
	handler := adminAccountHandler(adminAccountAuthenticatorStub{err: adminapp.ErrSessionNotFound})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/account", nil)
	request.Header.Set("Authorization", "Bearer expired")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}
