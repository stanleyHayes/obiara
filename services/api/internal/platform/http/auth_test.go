package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type registrationStub struct {
	requestOtp func(context.Context, string) (application.OtpRequest, error)
	verifyOtp  func(context.Context, string, string, string) (application.IssuedSession, error)
}

func (stub registrationStub) RequestOtp(ctx context.Context, phone string) (application.OtpRequest, error) {
	return stub.requestOtp(ctx, phone)
}

func (stub registrationStub) VerifyOtp(ctx context.Context, phone, code, deviceID string) (application.IssuedSession, error) {
	return stub.verifyOtp(ctx, phone, code, deviceID)
}

// sessionsStub drives the refresh route. Existing OTP tests pass a zero
// value, which is never called on those paths.
type sessionsStub struct {
	refresh func(context.Context, string) (application.IssuedSession, error)
}

func (stub sessionsStub) Refresh(ctx context.Context, token string) (application.IssuedSession, error) {
	if stub.refresh == nil {
		return application.IssuedSession{}, errors.New("refresh not stubbed")
	}
	return stub.refresh(ctx, token)
}

func authHandler(registration Registration) http.Handler {
	return authHandlerWith(registration, sessionsStub{})
}

func authHandlerWith(registration Registration, sessions Sessions) http.Handler {
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, registration, sessions)
	return Correlation(mux)
}

func postJSON(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestRequestOtpAccepted(t *testing.T) {
	expires := time.Date(2026, time.July, 26, 12, 10, 0, 0, time.UTC)
	handler := authHandler(registrationStub{
		requestOtp: func(_ context.Context, phone string) (application.OtpRequest, error) {
			if phone != "+233550000101" {
				t.Fatalf("phone = %q", phone)
			}
			return application.OtpRequest{ChallengeID: "ch-1", ExpiresAt: expires}, nil
		},
	})

	response := postJSON(t, handler, "/v1/auth/otp", `{"phone":"+233550000101"}`)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data otpRequestResponse `json:"data"`
	}
	decodeResponse(t, response, &envelope)
	if envelope.Data.ChallengeID != "ch-1" || !envelope.Data.ExpiresAt.Equal(expires) {
		t.Fatalf("data = %#v", envelope.Data)
	}
	if strings.Contains(response.Body.String(), "123456") {
		t.Fatal("response must never contain an OTP code")
	}
}

func TestRequestOtpValidation(t *testing.T) {
	handler := authHandler(registrationStub{
		requestOtp: func(context.Context, string) (application.OtpRequest, error) {
			t.Fatal("application must not be called for invalid input")
			return application.OtpRequest{}, nil
		},
	})

	for name, body := range map[string]string{
		"not e164":       `{"phone":"0550000101"}`,
		"empty phone":    `{"phone":""}`,
		"unknown field":  `{"phone":"+233550000101","extra":1}`,
		"malformed json": `{"phone":`,
	} {
		t.Run(name, func(t *testing.T) {
			response := postJSON(t, handler, "/v1/auth/otp", body)
			if response.Code == http.StatusAccepted {
				t.Fatalf("invalid input accepted: %s", body)
			}
		})
	}
}

func TestRequestOtpRateLimitedMaps429(t *testing.T) {
	handler := authHandler(registrationStub{
		requestOtp: func(context.Context, string) (application.OtpRequest, error) {
			return application.OtpRequest{}, domain.ErrOtpRateLimited
		},
	})
	response := postJSON(t, handler, "/v1/auth/otp", `{"phone":"+233550000101"}`)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d", response.Code)
	}
	var envelope errorEnvelope
	decodeResponse(t, response, &envelope)
	if envelope.Error.Code != "otp_rate_limited" {
		t.Fatalf("code = %q", envelope.Error.Code)
	}
}

func TestVerifyOtpIssuesSession(t *testing.T) {
	access, err := domain.IssueAccessToken("sess_1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := domain.IssueRefreshToken("sess_1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	session, err := domain.Start("sess_1", "member-1", "device-1", access, refresh, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	handler := authHandler(registrationStub{
		verifyOtp: func(_ context.Context, phone, code, deviceID string) (application.IssuedSession, error) {
			if phone != "+233550000101" || code != "123456" || deviceID != "device-1" {
				t.Fatalf("verify args = %q %q %q", phone, code, deviceID)
			}
			return application.IssuedSession{Session: session, AccessToken: access.Plain, RefreshToken: refresh.Plain}, nil
		},
	})

	response := postJSON(t, handler, "/v1/auth/otp/verify", `{"phone":"+233550000101","code":"123456","deviceId":"device-1"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data sessionResponse `json:"data"`
	}
	decodeResponse(t, response, &envelope)
	if envelope.Data.SessionID != "sess_1" || envelope.Data.MemberID != "member-1" || envelope.Data.AccessToken != access.Plain {
		t.Fatalf("data = %#v", envelope.Data)
	}
}

func TestVerifyOtpErrorMapping(t *testing.T) {
	cases := map[string]struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		"mismatch":         {domain.ErrOtpMismatch, http.StatusUnauthorized, "otp_invalid"},
		"expired":          {domain.ErrOtpExpired, http.StatusUnauthorized, "otp_expired"},
		"attempts":         {domain.ErrOtpAttemptsExceeded, http.StatusUnauthorized, "otp_attempts_exceeded"},
		"blocked account":  {domain.ErrAccountNotUsable, http.StatusForbidden, "account_not_active"},
		"internal no leak": {errors.New("mongo secret host"), http.StatusInternalServerError, "internal_error"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			handler := authHandler(registrationStub{
				verifyOtp: func(context.Context, string, string, string) (application.IssuedSession, error) {
					return application.IssuedSession{}, tc.err
				},
			})
			response := postJSON(t, handler, "/v1/auth/otp/verify", `{"phone":"+233550000101","code":"123456","deviceId":"device-1"}`)
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tc.wantStatus)
			}
			var envelope errorEnvelope
			decodeResponse(t, response, &envelope)
			if envelope.Error.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", envelope.Error.Code, tc.wantCode)
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatal("internal error leaked")
			}
		})
	}
}
