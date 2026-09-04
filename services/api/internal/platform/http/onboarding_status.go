package apihttp

import (
	"context"
	"errors"
	"net/http"

	consentapplication "github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	verificationapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	verificationdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
	livenessapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
	livenessdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

// OnboardingConsentState reads back whether a member has already accepted a
// versioned purpose. It is the consent context's own Effective port.
type OnboardingConsentState interface {
	Effective(ctx context.Context, subjectID, purposeID string, version uint64) (bool, error)
}

// OnboardingIdentityState reads back a member's standing identity case.
type OnboardingIdentityState interface {
	LatestCase(ctx context.Context, accountID string) (verificationdomain.VerificationCase, error)
}

// OnboardingLivenessState reads back a member's standing liveness attempt.
type OnboardingLivenessState interface {
	LatestAttempt(ctx context.Context, subjectID string) (livenessdomain.Attempt, error)
}

// RegisterOnboardingStatusRoutes exposes where a member stands in the walk.
func RegisterOnboardingStatusRoutes(
	mux *http.ServeMux,
	consents OnboardingConsentState,
	identity OnboardingIdentityState,
	liveness OnboardingLivenessState,
	sessions SessionAuthenticator,
) {
	mux.Handle("GET /v1/onboarding/status", onboardingStatusHandler(consents, identity, liveness, sessions))
}

// Step states are deliberately coarse. The console needs to know which of the
// four doors to open, not the provider's reasoning — that stays inside the
// verification context, where the audit trail lives.
const (
	stepUnstarted = "unstarted"
	stepPending   = "pending"
	stepPassed    = "passed"
	stepRejected  = "rejected"
	stepInReview  = "in_review"
)

type onboardingStatusResponse struct {
	ConsentsAccepted bool   `json:"consentsAccepted"`
	Identity         string `json:"identity"`
	Liveness         string `json:"liveness"`
}

// onboardingStatusHandler answers "where was I".
//
// Onboarding is four steps long and each one costs the member something: a
// message, a card number, a camera check, sometimes a reviewer's time. The
// console held all of that in a reducer that reset on refresh, so a closed tab
// or an expired access token sent a member back to step one and spent every
// one of those again — including queueing a second identity case for a
// reviewer already looking at the first.
//
// Every fact here is already recorded by the context that owns it. Nothing is
// invented, and no provider reasoning, card value or capture reference leaves
// those contexts: the projection is three coarse states and a boolean.
func onboardingStatusHandler(
	consents OnboardingConsentState,
	identity OnboardingIdentityState,
	liveness OnboardingLivenessState,
	sessions SessionAuthenticator,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || sessions == nil || consents == nil || identity == nil || liveness == nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(r.Context(), token)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		memberID := session.MemberID()

		accepted, err := onboardingConsentsAccepted(r.Context(), consents, memberID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}

		identityState, err := onboardingIdentityState(r.Context(), identity, memberID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}

		livenessState, err := onboardingLivenessState(r.Context(), liveness, memberID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}

		writeSuccess(w, r, http.StatusOK, onboardingStatusResponse{
			ConsentsAccepted: accepted,
			Identity:         identityState,
			Liveness:         livenessState,
		})
	})
}

// onboardingConsentsAccepted requires all three receipts at the current
// version. Two of three is not consent: the walk asks for the Promise, the
// terms and the adult affirmation together, and resuming past a missing one
// would record a member as having agreed to something they never saw.
func onboardingConsentsAccepted(ctx context.Context, consents OnboardingConsentState, memberID string) (bool, error) {
	for _, purposeID := range []string{
		consentapplication.CommunityPromiseID,
		consentapplication.ServiceTermsID,
		consentapplication.AdultAgeID,
	} {
		effective, err := consents.Effective(ctx, memberID, purposeID, consentapplication.CurrentVersion)
		if err != nil {
			return false, err
		}
		if !effective {
			return false, nil
		}
	}
	return true, nil
}

func onboardingIdentityState(ctx context.Context, identity OnboardingIdentityState, memberID string) (string, error) {
	verificationCase, err := identity.LatestCase(ctx, memberID)
	if errors.Is(err, verificationapplication.ErrCaseNotFound) {
		return stepUnstarted, nil
	}
	if err != nil {
		return "", err
	}
	switch verificationCase.Status() {
	case verificationdomain.StatusApproved:
		return stepPassed, nil
	case verificationdomain.StatusQueuedManual:
		return stepInReview, nil
	case verificationdomain.StatusRejected:
		return stepRejected, nil
	default:
		return stepPending, nil
	}
}

func onboardingLivenessState(ctx context.Context, liveness OnboardingLivenessState, memberID string) (string, error) {
	attempt, err := liveness.LatestAttempt(ctx, memberID)
	if errors.Is(err, livenessapplication.ErrAttemptNotFound) {
		return stepUnstarted, nil
	}
	if err != nil {
		return "", err
	}
	switch attempt.Status() {
	case livenessdomain.StatusPassed:
		return stepPassed, nil
	case livenessdomain.StatusQueuedManual:
		return stepInReview, nil
	case livenessdomain.StatusFailed:
		return stepRejected, nil
	default:
		return stepPending, nil
	}
}
