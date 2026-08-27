package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type adminMemberAccountsStub struct {
	items []identitydomain.Account
}

func (stub adminMemberAccountsStub) List(context.Context, int) ([]identitydomain.Account, error) {
	return stub.items, nil
}

type adminMemberKeyerStub struct{}

func (adminMemberKeyerStub) Key(string, string) (string, error) { return "member_ref_7K2", nil }

func TestAdminMemberDirectoryIsRedacted(t *testing.T) {
	created := time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)
	contact := identitydomain.ReconstituteContact(identitydomain.ChannelSMS, "+233550000101")
	account := identitydomain.ReconstituteAccount(
		"member-secret-id", contact, identitydomain.AccountActive,
		identitydomain.TierVerified, 3, nil, created,
	)
	mux := http.NewServeMux()
	RegisterAdminMemberRoutes(
		mux,
		adminMemberAccountsStub{items: []identitydomain.Account{account}},
		adminMemberKeyerStub{},
		func(*http.Request) (admin.Principal, error) {
			return admin.Principal{ActorID: "operator-1", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
		},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/members", nil)
	request.Header.Set("Authorization", "Bearer session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, forbidden := range []string{"+233550000101", "member-secret-id"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "member_ref_7K2") {
		t.Fatalf("response missing privacy ref: %s", body)
	}
}

func TestAdminMemberDirectoryRequiresOperationsOrSafety(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAdminMemberRoutes(
		mux, adminMemberAccountsStub{}, adminMemberKeyerStub{},
		func(*http.Request) (admin.Principal, error) {
			return admin.Principal{ActorID: "finance-1", Scopes: []admin.Scope{admin.ScopeFinance}}, nil
		},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/members", nil)
	request.Header.Set("Authorization", "Bearer session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
