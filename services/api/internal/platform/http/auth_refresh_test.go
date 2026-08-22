package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

func issuedSessionFixture(t *testing.T) application.IssuedSession {
	t.Helper()
	now := time.Date(2026, time.August, 22, 6, 0, 0, 0, time.UTC)
	access, err := domain.IssueAccessToken("ses_1", now)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	refresh, err := domain.IssueRefreshToken("ses_1", now)
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}
	session, err := domain.Start("ses_1", "mem_1", "device_1", access, refresh, now)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return application.IssuedSession{
		Session: session, AccessToken: access.Plain, RefreshToken: refresh.Plain,
	}
}

// TestRefreshRotatesTheTokenPair is the endpoint's reason for existing:
// access tokens live fifteen minutes, so without rotation every member
// redoes the SMS OTP flow four times an hour at one SMS each.
func TestRefreshRotatesTheTokenPair(t *testing.T) {
	issued := issuedSessionFixture(t)
	var presented string
	handler := authHandlerWith(registrationStub{}, sessionsStub{
		refresh: func(_ context.Context, token string) (application.IssuedSession, error) {
			presented = token
			return issued, nil
		},
	})

	response := postJSON(t, handler, "/v1/auth/refresh", `{"refreshToken":"rt_presented"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if presented != "rt_presented" {
		t.Errorf("service received %q", presented)
	}

	var envelope struct {
		Data struct {
			SessionID    string `json:"sessionId"`
			MemberID     string `json:"memberId"`
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envelope.Data.AccessToken == "" || envelope.Data.RefreshToken == "" {
		t.Error("response omitted a token; the client cannot continue the session")
	}
	if envelope.Data.MemberID != "mem_1" {
		t.Errorf("memberId = %q", envelope.Data.MemberID)
	}
}

// TestRefreshFailuresAreIndistinguishable keeps the endpoint from telling an
// attacker whether the token they hold is expired, wrong, or one they stole
// and had already been rotated out.
func TestRefreshFailuresAreIndistinguishable(t *testing.T) {
	cases := map[string]error{
		"reused after rotation": domain.ErrRefreshReuse,
		"wrong token":           domain.ErrRefreshTokenMismatch,
		"expired session":       domain.ErrSessionExpired,
		"revoked session":       domain.ErrSessionNotActive,
		"malformed token":       domain.ErrTokenMalformed,
		"unknown session":       application.ErrSessionNotFound,
	}

	var bodies []string
	for name, cause := range cases {
		t.Run(name, func(t *testing.T) {
			handler := authHandlerWith(registrationStub{}, sessionsStub{
				refresh: func(context.Context, string) (application.IssuedSession, error) {
					return application.IssuedSession{}, cause
				},
			})
			response := postJSON(t, handler, "/v1/auth/refresh", `{"refreshToken":"rt"}`)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", response.Code)
			}
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if envelope.Error.Code != "refresh_invalid" {
				t.Errorf("code = %q, want refresh_invalid for every rejection", envelope.Error.Code)
			}
			bodies = append(bodies, envelope.Error.Code)
		})
	}
	for _, code := range bodies {
		if code != bodies[0] {
			t.Fatal("rejection responses differ, letting a caller distinguish the causes")
		}
	}
}

func TestRefreshRejectsMalformedRequests(t *testing.T) {
	handler := authHandlerWith(registrationStub{}, sessionsStub{})

	t.Run("missing token", func(t *testing.T) {
		if got := postJSON(t, handler, "/v1/auth/refresh", `{}`).Code; got != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", got)
		}
	})
	t.Run("blank token", func(t *testing.T) {
		if got := postJSON(t, handler, "/v1/auth/refresh", `{"refreshToken":"   "}`).Code; got != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", got)
		}
	})
	t.Run("not json", func(t *testing.T) {
		if got := postJSON(t, handler, "/v1/auth/refresh", `{`).Code; got != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", got)
		}
	})
}
