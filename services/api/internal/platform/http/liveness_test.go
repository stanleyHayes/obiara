package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
	livenessdomain "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

type livenessStub struct {
	submit func(context.Context, application.SubmitRequest) (application.Result, error)
}

func (stub livenessStub) Submit(ctx context.Context, request application.SubmitRequest) (application.Result, error) {
	return stub.submit(ctx, request)
}

func TestLivenessUsesAuthenticatedSubjectAndReturnsManualReview(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access", now.Add(time.Hour), "refresh", "", now.Add(24*time.Hour),
		1, now, now,
	)
	attempt, err := livenessdomain.NewAttempt(
		"attempt-1", "liveness-request-1", strings.Repeat("a", 64),
		strings.Repeat("b", 64), now,
	)
	if err != nil {
		t.Fatal(err)
	}
	attempt, err = attempt.QueueManual(
		livenessdomain.ReasonProviderUncertain, strings.Repeat("a", 64),
		now.Add(time.Second), attempt.Version(),
	)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterLivenessRoutes(
		mux,
		livenessStub{submit: func(_ context.Context, request application.SubmitRequest) (application.Result, error) {
			if request.SubjectID != "member-1" || request.CommandID != "liveness-request-1" ||
				request.VoiceArtifactRef != "voice-asset-1" || request.FaceArtifactRef != "face-asset-1" {
				t.Fatalf("request = %+v", request)
			}
			return application.Result{Attempt: attempt}, application.ErrManualReviewNeeded
		}},
		nil,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}},
	)
	request := httptest.NewRequest(
		http.MethodPost, "/v1/verifications/liveness",
		strings.NewReader(`{"voiceArtifactRef":"voice-asset-1","faceArtifactRef":"face-asset-1"}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Idempotency-Key", "liveness-request-1")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
