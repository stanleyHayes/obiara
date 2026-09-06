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
	held, err := sowdomain.Accept("sow_1", "actor-key", "target-key", "hello",
		[]sowdomain.Media{{Key: "media-key", ScreeningKey: "screen-key"}},
		"cmd-1", "fingerprint", 1, sowdomain.StatusPendingReview, "review-1",
		sowdomain.Delivery{SowerID: "member_1", TargetID: "member_2", MediaRefs: []string{"mine"}},
		nowForSow())
	if err != nil {
		t.Fatal(err)
	}
	stub := &sowStub{result: sowapplication.Result{Sow: held}}

	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hello","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.command.ActorID != "member_1" {
		t.Fatalf("actor = %q, want the authenticated member", stub.command.ActorID)
	}
	if stub.command.ID != "cmd-1" {
		t.Fatalf("command id = %q, want the idempotency key", stub.command.ID)
	}
	// And the sow knows who it is toward. Without this the service has
	// nobody to check the listen gate, the blocks or the declines against.
	if stub.command.TargetID != "member_2" {
		t.Fatalf("target = %q, want the member the sow names", stub.command.TargetID)
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
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hello","mediaRefs":["mine"]}`, "cmd-1")
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
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hi","mediaRefs":["theirs"],"confirmed":true}`, "cmd-1")
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
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
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
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hello","mediaRefs":["mine"],"confirmed":true}`, "")
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
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sows", strings.NewReader(`{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`))
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

func TestAClosedReachIsNotToldWhichRuleClosedIt(t *testing.T) {
	// A block and a decline answer identically on purpose. Telling them
	// apart would tell a member they were declined, which is the rejection
	// signal FR-205 exists to withhold.
	stub := &sowStub{err: sowapplication.ErrReachNotAvailable}
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "reach_unavailable") {
		t.Fatalf("body = %s", response.Body.String())
	}
	for _, leak := range []string{"block", "declin"} {
		if strings.Contains(strings.ToLower(response.Body.String()), leak) {
			t.Fatalf("the refusal named which rule closed it (%q): %s", leak, response.Body.String())
		}
	}
}

func TestASowBeforeListeningSaysWhatToDoAboutIt(t *testing.T) {
	// Not heard yet is a rule, not a fault, and one the member can act on.
	// Without its own case it fell to the default and read as a server
	// error, which is untrue and no help.
	stub := &sowStub{err: sowapplication.ErrNotHeard}
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "not_heard_yet") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestASowThatDidNotArriveTellsTheMemberNotToResend(t *testing.T) {
	// The seed is already spent and the sow is already recorded. A member
	// told "try again" would send a second one and be charged twice for the
	// same gesture.
	stub := &sowStub{err: sowapplication.ErrNotDelivered}
	response := sowRequest(t, stub,
		`{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "sow_not_delivered") {
		t.Fatalf("body = %s", response.Body.String())
	}
	if !strings.Contains(strings.ToLower(response.Body.String()), "do not send it again") {
		t.Fatalf("the member was not told to leave it alone: %s", response.Body.String())
	}
}

func TestASowMustCarryARecording(t *testing.T) {
	// A pod is a recording resting at a house front, so a sow with none has
	// nothing to place. The transport passes it through and the service
	// refuses; what this proves is that the empty list reaches the service
	// as an empty list rather than being filled in on the way.
	stub := &sowStub{err: sowdomain.ErrInvalid}
	response := sowRequest(t, stub, `{"targetId":"member_2","body":"hi","confirmed":true}`, "cmd-1")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if len(stub.command.MediaRefs) != 0 {
		t.Fatalf("the transport invented recordings: %v", stub.command.MediaRefs)
	}
}
