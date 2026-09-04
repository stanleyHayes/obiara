package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	consentapplication "github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	verificationapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	verificationdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
	livenessapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
	livenessdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

type onboardingConsentStateStub struct {
	granted map[string]bool
}

func (stub onboardingConsentStateStub) Effective(_ context.Context, _, purposeID string, _ uint64) (bool, error) {
	return stub.granted[purposeID], nil
}

type onboardingIdentityStateStub struct {
	verificationCase verificationdomain.VerificationCase
	err              error
}

func (stub onboardingIdentityStateStub) LatestCase(context.Context, string) (verificationdomain.VerificationCase, error) {
	return stub.verificationCase, stub.err
}

type onboardingLivenessStateStub struct {
	attempt livenessdomain.Attempt
	err     error
}

func (stub onboardingLivenessStateStub) LatestAttempt(context.Context, string) (livenessdomain.Attempt, error) {
	return stub.attempt, stub.err
}

func onboardingStatusFixture(
	t *testing.T,
	consents onboardingConsentStateStub,
	identity onboardingIdentityStateStub,
	liveness onboardingLivenessStateStub,
) onboardingStatusResponse {
	t.Helper()
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access", now.Add(time.Hour), "refresh", "", now.Add(24*time.Hour),
		1, now, now,
	)
	mux := http.NewServeMux()
	RegisterOnboardingStatusRoutes(mux, consents, identity, liveness,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/onboarding/status", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data onboardingStatusResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v body=%s", err, response.Body.String())
	}
	return envelope.Data
}

func allOnboardingConsents() onboardingConsentStateStub {
	return onboardingConsentStateStub{granted: map[string]bool{
		consentapplication.CommunityPromiseID: true,
		consentapplication.ServiceTermsID:     true,
		consentapplication.AdultAgeID:         true,
	}}
}

func TestOnboardingStatusReportsAnUntouchedWalk(t *testing.T) {
	status := onboardingStatusFixture(t,
		onboardingConsentStateStub{granted: map[string]bool{}},
		onboardingIdentityStateStub{err: verificationapplication.ErrCaseNotFound},
		onboardingLivenessStateStub{err: livenessapplication.ErrAttemptNotFound},
	)
	if status.ConsentsAccepted || status.Identity != stepUnstarted || status.Liveness != stepUnstarted {
		t.Fatalf("status = %+v", status)
	}
}

func TestOnboardingStatusRefusesPartialConsent(t *testing.T) {
	// Two of three is not consent. Resuming past a missing acknowledgement
	// would record a member as having agreed to something they never saw.
	status := onboardingStatusFixture(t,
		onboardingConsentStateStub{granted: map[string]bool{
			consentapplication.CommunityPromiseID: true,
			consentapplication.ServiceTermsID:     true,
		}},
		onboardingIdentityStateStub{err: verificationapplication.ErrCaseNotFound},
		onboardingLivenessStateStub{err: livenessapplication.ErrAttemptNotFound},
	)
	if status.ConsentsAccepted {
		t.Fatalf("partial consent reported as accepted: %+v", status)
	}
}

func TestOnboardingStatusProjectsAQueuedIdentityCaseAsReview(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	queued := verificationdomain.ReconstituteCase(
		"vc_QWx0Z2ViZXJ0X21lbWJlcg", "member-1", "card-key", "GHA-***",
		verificationdomain.StatusQueuedManual, "", "", now, 2, now, nil,
	)
	status := onboardingStatusFixture(t,
		allOnboardingConsents(),
		onboardingIdentityStateStub{verificationCase: queued},
		onboardingLivenessStateStub{err: livenessapplication.ErrAttemptNotFound},
	)
	if !status.ConsentsAccepted || status.Identity != stepInReview {
		t.Fatalf("status = %+v", status)
	}
	// The reviewer's queue is not a card number: no keyed or masked card
	// value may reach the console through this projection.
	if body, _ := json.Marshal(status); string(body) == "" ||
		containsAny(string(body), "card-key", "GHA-") {
		t.Fatalf("projection leaked card material: %s", body)
	}
}

func TestOnboardingStatusProjectsAPassedLivenessAttempt(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	pending, err := livenessdomain.NewAttempt(
		"la_1", "cmd_1", strings.Repeat("a", 64), strings.Repeat("b", 64), now,
	)
	if err != nil {
		t.Fatalf("new attempt: %v", err)
	}
	passed, err := pending.ProviderDecision(
		true, "provider:proof", strings.Repeat("c", 64), now.Add(time.Minute), pending.Version(),
	)
	if err != nil {
		t.Fatalf("provider decision: %v", err)
	}
	status := onboardingStatusFixture(t,
		allOnboardingConsents(),
		onboardingIdentityStateStub{err: verificationapplication.ErrCaseNotFound},
		onboardingLivenessStateStub{attempt: passed},
	)
	if status.Liveness != stepPassed {
		t.Fatalf("status = %+v", status)
	}
}

func containsAny(haystack string, needles ...string) bool {
	for _, needle := range needles {
		for index := 0; index+len(needle) <= len(haystack); index++ {
			if haystack[index:index+len(needle)] == needle {
				return true
			}
		}
	}
	return false
}
