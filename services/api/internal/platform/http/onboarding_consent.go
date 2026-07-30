package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	consentapplication "github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	consentdomain "github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

type OnboardingConsent interface {
	Accept(context.Context, consentapplication.OnboardingCommand) (consentapplication.OnboardingResult, error)
}

func RegisterOnboardingConsentRoutes(mux *http.ServeMux, service OnboardingConsent, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/onboarding/consents", onboardingConsentHandler(service, sessions))
}

type onboardingConsentResponse struct {
	PromiseRevision uint64 `json:"promiseRevision"`
	TermsRevision   uint64 `json:"termsRevision"`
	AgeRevision     uint64 `json:"ageRevision"`
}

func onboardingConsentHandler(service OnboardingConsent, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok || sessions == nil || service == nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(request.Context(), token)
		if err != nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		commandID := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(writer, request, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		result, err := service.Accept(request.Context(), consentapplication.OnboardingCommand{
			CommandID: commandID, SubjectID: session.MemberID(), Source: consentdomain.SourceWeb,
		})
		if err != nil {
			status, code := http.StatusInternalServerError, "internal_error"
			if errors.Is(err, consentdomain.ErrStaleRevision) ||
				errors.Is(err, consentdomain.ErrCommandMismatch) {
				status, code = http.StatusConflict, "consent_conflict"
			}
			writeError(writer, request, status, APIError{
				Code: code, Message: "The acknowledgements could not be recorded.",
			})
			return
		}
		writeSuccess(writer, request, http.StatusCreated, onboardingConsentResponse{
			PromiseRevision: result.PromiseRevision,
			TermsRevision:   result.TermsRevision,
			AgeRevision:     result.AgeRevision,
		})
	})
}
