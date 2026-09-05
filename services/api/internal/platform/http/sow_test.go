package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"

	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

type sowStub struct {
	command sowapplication.Command
	result  sowapplication.Result
	err     error
	calls   int
}

func (stub *sowStub) Send(_ context.Context, command sowapplication.Command) (sowapplication.Result, error) {
	stub.calls++
	stub.command = command
	return stub.result, stub.err
}

func nowForSow() time.Time {
	return time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)
}

func identityTierUnverified() identitydomain.Tier { return identitydomain.TierUnverified }

func sowRequest(t *testing.T, stub *sowStub, body, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterSowRoutes(mux, stub, sessionStub{memberID: "member_1"}, sowingGate())
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sows", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

func TestASowIsSentAsTheSessionAndReportsThatItIsWaiting(t *testing.T) {
	// The sower comes from the session: a member cannot sow on somebody
	// else's behalf, and the seed comes out of their own allowance.
	held, err := sowdomain.Accept("sow_1", "actor-key", "hello", nil,
		"cmd-1", "fingerprint", 1, sowdomain.StatusPendingReview, "review-1", nowForSow())
	if err != nil {
		t.Fatal(err)
	}
	stub := &sowStub{result: sowapplication.Result{Sow: held}}

	response := sowRequest(t, stub, `{"body":"hello","confirmed":true}`, "cmd-1")
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.command.ActorID != "member_1" {
		t.Fatalf("actor = %q, want the authenticated member", stub.command.ActorID)
	}
	if stub.command.ID != "cmd-1" {
		t.Fatalf("command id = %q, want the idempotency key", stub.command.ID)
	}
	// The member is told it is waiting rather than shown a delivery that has
	// not happened.
	if !strings.Contains(response.Body.String(), "pending_review") {
		t.Fatalf("body did not say the sow is waiting: %s", response.Body.String())
	}
}

func TestAnUnconfirmedSowIsRefusedRatherThanDefaulted(t *testing.T) {
	// A sow costs a seed and reaches a person. Neither should happen by
	// brushing a screen, so the gesture is required and never assumed.
	stub := &sowStub{err: sowdomain.ErrNotConfirmed}
	response := sowRequest(t, stub, `{"body":"hello"}`, "cmd-1")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "confirmation_required") {
		t.Fatalf("body = %s", response.Body.String())
	}
	// The command still carried Confirmed=false rather than the transport
	// helpfully filling it in.
	if stub.command.Confirmed {
		t.Fatal("the transport confirmed the gesture on the member's behalf")
	}
}

func TestASowWithSomebodyElsesRecordingIsRefused(t *testing.T) {
	stub := &sowStub{err: sowapplication.ErrMediaNotOwned}
	response := sowRequest(t, stub, `{"body":"hi","mediaRefs":["theirs"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "recording_not_yours") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestARefusedSowIsNotToldWhy(t *testing.T) {
	// A refusal that explains itself teaches somebody how to word the next
	// one. The member is told it could not be sent and nothing more.
	stub := &sowStub{err: sowdomain.ErrScreeningRejected}
	response := sowRequest(t, stub, `{"body":"hi","confirmed":true}`, "cmd-1")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	for _, leak := range []string{"contact", "payment", "sexual", "threat", "screening"} {
		if strings.Contains(strings.ToLower(response.Body.String()), leak) {
			t.Fatalf("the refusal named a reason (%q): %s", leak, response.Body.String())
		}
	}
}

func TestASowWithoutARequestIdNeverReachesTheSeedEconomy(t *testing.T) {
	// Without one the sow is not idempotent, and a double submission would
	// spend two seeds for one gesture.
	stub := &sowStub{}
	response := sowRequest(t, stub, `{"body":"hello","confirmed":true}`, "")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if stub.calls != 0 {
		t.Fatal("a sow with no request id reached the service")
	}
}

func TestAnUnverifiedMemberCannotSow(t *testing.T) {
	stub := &sowStub{}
	mux := http.NewServeMux()
	RegisterSowRoutes(mux, stub, sessionStub{memberID: "member_1"}, gateAt(identityTierUnverified()))
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sows", strings.NewReader(`{"body":"hi","confirmed":true}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "cmd-1")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if stub.calls != 0 {
		t.Fatal("an unverified member's sow reached the service")
	}
}
