package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/reviewdesk"
	screeningmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/mongodb"
	screeningapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
	screeningdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type screeningQueueStub struct {
	pending []screeningmongo.Pending
	err     error
	// actor records who the handler said was reading, so the test can assert
	// the audit trail is given a real name rather than an empty one.
	actor *string
}

func (stub screeningQueueStub) Pending(_ context.Context, actorID string, _ int) ([]screeningmongo.Pending, error) {
	if stub.actor != nil {
		*stub.actor = actorID
	}
	return stub.pending, stub.err
}

type screeningDeskStub struct {
	approve   bool
	actor     string
	commandID string
	calls     int
	err       error
}

func (stub *screeningDeskStub) Decide(_ context.Context, _ string, approve bool, actorID, commandID string) error {
	stub.calls++
	stub.approve, stub.actor, stub.commandID = approve, actorID, commandID
	return stub.err
}

func screeningDesk() admin.Principal {
	return admin.Principal{ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}, MFAVerified: true}
}

func screeningMux(principal admin.Principal, queue AdminScreeningQueue, desk AdminScreeningDesk) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterAdminScreeningRoutes(mux, queue, desk, func(*http.Request) (admin.Principal, error) {
		return principal, nil
	})
	return mux
}

func pendingReview(t *testing.T) screeningmongo.Pending {
	t.Helper()
	review, err := screeningdomain.Route(strings.Repeat("a", 64), "uncertain",
		time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return screeningmongo.Pending{
		Review: review,
		Text:   "I liked what you said about your grandmother.",
		Media: []screeningapplication.MediaMetadata{
			{MIME: "audio/ogg", Bytes: 2048, DurationMs: 45_000},
		},
		AdvisoryReasons: []string{"uncertain"},
	}
}

func TestTheQueueShowsTheWordsAndDescribesTheRecording(t *testing.T) {
	// A reviewer must read the words — that is the job. The recording is
	// described rather than carried: they need to know what is attached, and
	// this response is not where audio is handed around.
	var actor string
	mux := screeningMux(screeningDesk(),
		screeningQueueStub{pending: []screeningmongo.Pending{pendingReview(t)}, actor: &actor},
		&screeningDeskStub{})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/screening/reviews", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if actor != "agent-1" {
		t.Fatalf("the read was attributed to %q, want the authenticated agent", actor)
	}
	body := response.Body.String()
	if !strings.Contains(body, "grandmother") {
		t.Fatalf("the reviewer was not shown the words: %s", body)
	}
	if !strings.Contains(body, "audio/ogg") || !strings.Contains(body, "45000") {
		t.Fatalf("the recording was not described: %s", body)
	}
}

func TestOnlyASteppedUpSafetyDeskSeesMembersWords(t *testing.T) {
	for name, principal := range map[string]admin.Principal{
		"no step-up":  {ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeSafety}},
		"wrong scope": {ActorID: "agent-1", Scopes: []admin.Scope{admin.ScopeOperations}, MFAVerified: true},
	} {
		mux := screeningMux(principal, screeningQueueStub{pending: []screeningmongo.Pending{pendingReview(t)}}, &screeningDeskStub{})
		request := httptest.NewRequest(http.MethodGet, "/v1/admin/screening/reviews", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code == http.StatusOK {
			t.Fatalf("%s: the queue was opened anyway", name)
		}
		if strings.Contains(response.Body.String(), "grandmother") {
			t.Fatalf("%s: a member's words leaked in the refusal", name)
		}
	}
}

func decisionRequest(t *testing.T, desk *screeningDeskStub, body, key string) *httptest.ResponseRecorder {
	t.Helper()
	mux := screeningMux(screeningDesk(), screeningQueueStub{}, desk)
	request := httptest.NewRequest(http.MethodPost,
		"/v1/admin/screening/reviews/"+strings.Repeat("a", 64)+"/decision", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func TestAnAbsentDecisionIsNeverAnApproval(t *testing.T) {
	// Defaulting here would deliver a sow nobody cleared.
	desk := &screeningDeskStub{}
	response := decisionRequest(t, desk, `{}`, "cmd-1")
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if desk.calls != 0 {
		t.Fatal("an absent decision reached the desk")
	}
}

func TestADecisionCarriesTheReviewerAndTheRequestId(t *testing.T) {
	desk := &screeningDeskStub{}
	response := decisionRequest(t, desk, `{"approve":true}`, "cmd-1")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if !desk.approve || desk.actor != "agent-1" || desk.commandID != "cmd-1" {
		t.Fatalf("desk = %#v", desk)
	}
	// Without a request id a repeated decision would refund a seed twice.
	bare := &screeningDeskStub{}
	if code := decisionRequest(t, bare, `{"approve":false}`, "").Code; code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", code)
	}
	if bare.calls != 0 {
		t.Fatal("a decision with no request id reached the desk")
	}
}

func TestAnAlreadyDecidedReviewSaysSo(t *testing.T) {
	desk := &screeningDeskStub{err: screeningdomain.ErrNotPending}
	if code := decisionRequest(t, desk, `{"approve":true}`, "cmd-1").Code; code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", code)
	}
	missing := &screeningDeskStub{err: reviewdesk.ErrNotFound}
	if code := decisionRequest(t, missing, `{"approve":true}`, "cmd-1").Code; code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
}
