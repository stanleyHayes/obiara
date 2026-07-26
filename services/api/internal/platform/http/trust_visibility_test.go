package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/visibility"
)

type trustVisibilityStub struct {
	explain func(context.Context, visibility.Request) ([]visibility.Explanation, error)
}

func (stub trustVisibilityStub) Explain(ctx context.Context, request visibility.Request) ([]visibility.Explanation, error) {
	return stub.explain(ctx, request)
}

type sessionAuthenticatorStub struct {
	authenticate func(context.Context, string) (identitydomain.Session, error)
}

func (stub sessionAuthenticatorStub) Authenticate(ctx context.Context, token string) (identitydomain.Session, error) {
	return stub.authenticate(ctx, token)
}

func TestTrustVisibilityBindsAuthenticatedOwnerAndReturnsEnvelope(t *testing.T) {
	session := trustSession(t, "member-1")
	var received visibility.Request
	handler := trustHTTPHandler(
		trustVisibilityStub{explain: func(_ context.Context, request visibility.Request) ([]visibility.Explanation, error) {
			received = request
			return []visibility.Explanation{{
				TargetID: "member-2", Hops: 1,
				Steps: []visibility.ExplanationStep{{
					SourceID: "member-1", TargetID: "member-2",
					Reason: visibility.ReasonKnownConnection,
				}},
			}}, nil
		}},
		sessionAuthenticatorStub{authenticate: func(_ context.Context, token string) (identitydomain.Session, error) {
			if token != "access-token" {
				t.Fatalf("token = %q", token)
			}
			return session, nil
		}},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/members/member-1/trust-paths?depth=2&nodes=10", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	request.Header.Set(CorrelationIDHeader, "trust-request-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if received.RequesterID != "member-1" || received.RootID != "member-1" ||
		received.MaxDepth != 2 || received.MaxNodes != 10 {
		t.Fatalf("request = %#v", received)
	}
	body := response.Body.String()
	for _, expected := range []string{
		`"targetId":"member-2"`, `"reason":"known_connection"`,
		`"correlationId":"trust-request-1"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body %q missing %q", body, expected)
		}
	}
	for _, forbidden := range []string{"edge-", "consent-", "provenance-"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body leaked %q: %s", forbidden, body)
		}
	}
}

func TestTrustVisibilityUniformlyHidesAuthOwnerAndPolicyFailures(t *testing.T) {
	session := trustSession(t, "member-1")
	cases := []struct {
		name       string
		path       string
		header     string
		authErr    error
		serviceErr error
	}{
		{name: "missing bearer", path: "/v1/members/member-1/trust-paths"},
		{name: "invalid session", path: "/v1/members/member-1/trust-paths", header: "Bearer invalid", authErr: errors.New("invalid")},
		{name: "owner mismatch", path: "/v1/members/hidden-member/trust-paths", header: "Bearer access"},
		{name: "policy denied", path: "/v1/members/member-1/trust-paths", header: "Bearer access", serviceErr: visibility.ErrNotVisible},
	}
	var expectedBody string
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			serviceCalled := false
			handler := trustHTTPHandler(
				trustVisibilityStub{explain: func(context.Context, visibility.Request) ([]visibility.Explanation, error) {
					serviceCalled = true
					return nil, test.serviceErr
				}},
				sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
					if test.authErr != nil {
						return identitydomain.Session{}, test.authErr
					}
					return session, nil
				}},
			)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("Authorization", test.header)
			request.Header.Set(CorrelationIDHeader, "uniform-trust")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "hidden-member") ||
				strings.Contains(response.Body.String(), "invalid session") {
				t.Fatalf("failure detail leaked: %s", response.Body.String())
			}
			if expectedBody == "" {
				expectedBody = response.Body.String()
			} else if response.Body.String() != expectedBody {
				t.Fatalf("non-uniform body:\ngot  %s\nwant %s", response.Body.String(), expectedBody)
			}
			if (test.name == "missing bearer" || test.name == "invalid session" || test.name == "owner mismatch") && serviceCalled {
				t.Fatal("visibility service called before authenticated owner binding")
			}
		})
	}
}

func TestTrustVisibilityRejectsBoundsAndRegistersNoGraphBrowser(t *testing.T) {
	session := trustSession(t, "member-1")
	called := false
	handler := trustHTTPHandler(
		trustVisibilityStub{explain: func(context.Context, visibility.Request) ([]visibility.Explanation, error) {
			called = true
			return nil, nil
		}},
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}},
	)
	for _, query := range []string{"depth=0", "depth=5", "nodes=1", "nodes=101", "depth=all", "nodes=-1"} {
		request := httptest.NewRequest(http.MethodGet, "/v1/members/member-1/trust-paths?"+query, nil)
		request.Header.Set("Authorization", "Bearer access")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("query %q status = %d, body = %s", query, response.Code, response.Body.String())
		}
	}
	if called {
		t.Fatal("service called with invalid bounds")
	}
	for _, path := range []string{"/v1/trust-paths", "/v1/trust-graph", "/v1/members/member-1/trust-paths/reverse"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("unexpected graph route %q status = %d", path, response.Code)
		}
	}
}

func trustHTTPHandler(service TrustVisibility, sessions SessionAuthenticator) http.Handler {
	mux := http.NewServeMux()
	RegisterTrustVisibilityRoutes(mux, service, sessions)
	return Correlation(mux)
}

func trustSession(t *testing.T, memberID string) identitydomain.Session {
	t.Helper()
	now := time.Now()
	access, err := identitydomain.IssueAccessToken("session-1", now)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := identitydomain.IssueRefreshToken("session-1", now)
	if err != nil {
		t.Fatal(err)
	}
	session, err := identitydomain.Start("session-1", memberID, "device-1", access, refresh, now)
	if err != nil {
		t.Fatal(err)
	}
	return session
}
