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
	listErr error
}

func (stub *waitlistStoreStub) Join(_ context.Context, name, email, consent string) (waitlist.Entry, bool, error) {
	stub.name, stub.email, stub.consent = name, email, consent
	return stub.entry, stub.created, nil
}

func (stub *waitlistStoreStub) List(context.Context, int) ([]waitlist.Entry, error) {
	return []waitlist.Entry{stub.entry}, stub.listErr
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

func TestJoinWaitlistThrottlesPerClientIP(t *testing.T) {
	store := &waitlistStoreStub{created: true, entry: waitlist.Entry{Email: "ama@example.com"}}
	mux := http.NewServeMux()
	RegisterWaitlistRoutes(mux, store, nil)
	for attempt := 1; attempt <= waitlistJoinLimit+1; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/v1/waitlist", strings.NewReader(`{"name":"Ama","email":"ama@example.com","consent":true}`))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if attempt <= waitlistJoinLimit && response.Code != http.StatusCreated {
			t.Fatalf("attempt %d: status = %d, want 201", attempt, response.Code)
		}
		if attempt > waitlistJoinLimit && response.Code != http.StatusTooManyRequests {
			t.Fatalf("attempt %d: status = %d, want 429, body = %s", attempt, response.Code, response.Body.String())
		}
	}
}

func TestJoinWaitlistThrottleIsPerIP(t *testing.T) {
	throttle := newWaitlistJoinThrottle()
	for attempt := 0; attempt < waitlistJoinLimit; attempt++ {
		if !throttle.allow("192.0.2.1:1234") {
			t.Fatalf("attempt %d under the limit must be allowed", attempt)
		}
	}
	if throttle.allow("192.0.2.1:1234") {
		t.Fatal("attempt beyond the limit must be rejected")
	}
	if !throttle.allow("198.51.100.7:1234") {
		t.Fatal("a different client IP must not inherit the limit")
	}
}

func TestAdminWaitlistStoreErrorReturns503(t *testing.T) {
	store := &waitlistStoreStub{listErr: context.Canceled}
	mux := http.NewServeMux()
	RegisterWaitlistRoutes(mux, store, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "ops-1", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
	})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/waitlist", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503, body = %s", response.Code, response.Body.String())
	}
}
