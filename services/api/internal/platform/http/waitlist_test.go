package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/waitlist"
)

type waitlistStoreStub struct {
	entry   waitlist.Entry
	created bool
	name    string
	email   string
	consent string
}

func (stub *waitlistStoreStub) Join(_ context.Context, name, email, consent string) (waitlist.Entry, bool, error) {
	stub.name, stub.email, stub.consent = name, email, consent
	return stub.entry, stub.created, nil
}

func (stub *waitlistStoreStub) List(context.Context, int) ([]waitlist.Entry, error) {
	return []waitlist.Entry{stub.entry}, nil
}

func TestJoinWaitlistNormalizesAndRecordsPurposeConsent(t *testing.T) {
	store := &waitlistStoreStub{created: true, entry: waitlist.Entry{Email: "ama@example.com"}}
	mux := http.NewServeMux()
	RegisterWaitlistRoutes(mux, store, nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/waitlist", strings.NewReader(`{"name":"  Ama Mensah  ","email":"AMA@EXAMPLE.COM","consent":true}`))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if store.name != "Ama Mensah" || store.email != "ama@example.com" || store.consent != waitlistConsentVersion {
		t.Fatalf("stored values = %q %q %q", store.name, store.email, store.consent)
	}
}

func TestJoinWaitlistRequiresExplicitConsent(t *testing.T) {
	mux := http.NewServeMux()
	RegisterWaitlistRoutes(mux, &waitlistStoreStub{}, nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/waitlist", strings.NewReader(`{"name":"Ama","email":"ama@example.com","consent":false}`))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
}

func TestAdminWaitlistRequiresOperationsScope(t *testing.T) {
	store := &waitlistStoreStub{entry: waitlist.Entry{Name: "Ama", Email: "ama@example.com", SignedUpAt: time.Now(), NotificationState: "pending"}}
	mux := http.NewServeMux()
	RegisterWaitlistRoutes(mux, store, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "finance-1", Scopes: []admin.Scope{admin.ScopeFinance}}, nil
	})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/waitlist", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
