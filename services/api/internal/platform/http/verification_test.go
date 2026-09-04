package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	verificationapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	verificationdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

type verificationStub struct {
	submit func(context.Context, string, string, time.Time) (verificationdomain.VerificationCase, error)
}

// The document path is exercised in the application package; this stub only
// has to satisfy the port so the card-number handler's tests still compile.
func (verificationStub) SubmitDocuments(
	context.Context,
	verificationapplication.SubmitDocumentsRequest,
) (verificationapplication.SubmitDocumentsResult, error) {
	panic("not used by the card-number handler tests")
}

func (stub verificationStub) SubmitGhanaCard(ctx context.Context, accountID, cardNumber string, dateOfBirth time.Time) (verificationdomain.VerificationCase, error) {
	return stub.submit(ctx, accountID, cardNumber, dateOfBirth)
}

func verificationHandler(verification Verification, sessions SessionAuthenticator) http.Handler {
	mux := http.NewServeMux()
	RegisterVerificationRoutes(mux, verification, sessions)
	return Correlation(mux)
}

func TestGhanaCardRequiresAuthenticatedSessionAndUsesItsMember(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access-hash", now.Add(time.Hour), "refresh-hash", "",
		now.Add(24*time.Hour), 1, now, now,
	)
	called := false
	handler := verificationHandler(
		verificationStub{submit: func(_ context.Context, accountID, cardNumber string, dateOfBirth time.Time) (verificationdomain.VerificationCase, error) {
			called = true
			if accountID != "member-1" || cardNumber != "GHA-123456789-0" ||
				dateOfBirth.Format("2006-01-02") != "1994-05-06" {
				t.Fatalf("unexpected verification input: %q %q %s", accountID, cardNumber, dateOfBirth)
			}
			value, err := verificationdomain.NewCase("case-1", accountID, "key_1", "6789", dateOfBirth, now)
			if err != nil {
				t.Fatal(err)
			}
			if err := value.Approve("provider-1", "issuer_match", now); err != nil {
				t.Fatal(err)
			}
			return value, nil
		}},
		sessionAuthenticatorStub{authenticate: func(_ context.Context, token string) (identitydomain.Session, error) {
			if token != "access-token" {
				t.Fatalf("token = %q", token)
			}
			return session, nil
		}},
	)

	request := httptest.NewRequest(http.MethodPost, "/v1/verifications/ghana-card", strings.NewReader(
		`{"cardNumber":"GHA-123456789-0","dateOfBirth":"1994-05-06"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || !called {
		t.Fatalf("status=%d called=%v body=%s", response.Code, called, response.Body.String())
	}
}

func TestGhanaCardRejectsMissingSessionBeforeApplication(t *testing.T) {
	handler := verificationHandler(
		verificationStub{submit: func(context.Context, string, string, time.Time) (verificationdomain.VerificationCase, error) {
			t.Fatal("verification application must not run without authentication")
			return verificationdomain.VerificationCase{}, nil
		}},
		nil,
	)
	request := httptest.NewRequest(http.MethodPost, "/v1/verifications/ghana-card", strings.NewReader(
		`{"cardNumber":"GHA-123456789-0","dateOfBirth":"1994-05-06"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestGhanaCardReturns409IdentityAlreadyVerified(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access-hash", now.Add(time.Hour), "refresh-hash", "",
		now.Add(24*time.Hour), 1, now, now,
	)
	handler := verificationHandler(
		verificationStub{submit: func(context.Context, string, string, time.Time) (verificationdomain.VerificationCase, error) {
			// Return the actual error sentinel from the application layer
			return verificationdomain.VerificationCase{}, verificationapplication.ErrIdentityAlreadyVerified
		}},
		sessionAuthenticatorStub{authenticate: func(_ context.Context, token string) (identitydomain.Session, error) {
			return session, nil
		}},
	)

	request := httptest.NewRequest(http.MethodPost, "/v1/verifications/ghana-card", strings.NewReader(
		`{"cardNumber":"GHA-123456789-0","dateOfBirth":"1994-05-06"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409, body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "identity_already_verified") {
		t.Fatalf("response body missing 'identity_already_verified', got: %s", response.Body.String())
	}
}
