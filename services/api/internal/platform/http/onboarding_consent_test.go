package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	consentapplication "github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type onboardingConsentStub struct {
	accept func(context.Context, consentapplication.OnboardingCommand) (consentapplication.OnboardingResult, error)
}

func (stub onboardingConsentStub) Accept(ctx context.Context, command consentapplication.OnboardingCommand) (consentapplication.OnboardingResult, error) {
	return stub.accept(ctx, command)
}

func TestOnboardingConsentUsesAuthenticatedMemberAndIdempotencyKey(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access", now.Add(time.Hour), "refresh", "", now.Add(24*time.Hour),
		1, now, now,
	)
	mux := http.NewServeMux()
	RegisterOnboardingConsentRoutes(
		mux,
		onboardingConsentStub{accept: func(_ context.Context, command consentapplication.OnboardingCommand) (consentapplication.OnboardingResult, error) {
			if command.SubjectID != "member-1" || command.CommandID != "consent-request-1" {
				t.Fatalf("command = %+v", command)
			}
			return consentapplication.OnboardingResult{PromiseRevision: 1, TermsRevision: 1, AgeRevision: 1}, nil
		}},
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}},
	)
	request := httptest.NewRequest(http.MethodPost, "/v1/onboarding/consents", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	request.Header.Set("Idempotency-Key", "consent-request-1")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
