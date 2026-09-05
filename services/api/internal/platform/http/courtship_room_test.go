package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	queueapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/application"
	queuedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"

	closuredomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	honestydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	pacedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	pausedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	safetydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
)

type roomStub struct {
	called    string
	actor     string
	commandID string
	revision  uint64
	action    pausedomain.Action
	category  safetydomain.Category
	grant     bool
	members   []string
	device    string
	payload   string
	base      uint64
	after     uint64
	limit     int
	err       error
}

func (stub *roomStub) Submit(_ context.Context, _, deviceRef, actorID, payloadRef, commandID string, base uint64) (queueapp.Result, error) {
	stub.called, stub.actor, stub.commandID, stub.device, stub.payload, stub.base = "submit", actorID, commandID, deviceRef, payloadRef, base
	return queueapp.Result{Event: queuedomain.Event{Sequence: base + 1}}, stub.err
}
func (stub *roomStub) Timeline(_ context.Context, _, memberID string, after uint64, limit int) ([]queuedomain.Event, error) {
	stub.called, stub.actor, stub.after, stub.limit = "timeline", memberID, after, limit
	return []queuedomain.Event{{Sequence: after + 1, AcceptedAt: time.Now()}}, stub.err
}
func (stub *roomStub) Start(_ context.Context, _ string, members []string, commandID, actorID string) (pacedomain.Pace, error) {
	stub.called, stub.actor, stub.commandID, stub.members = "start", actorID, commandID, members
	return pacedomain.Pace{}, stub.err
}
func (stub *roomStub) Relight(_ context.Context, _, memberID, commandID string, revision uint64) (pacedomain.Pace, error) {
	stub.called, stub.actor, stub.commandID, stub.revision = "relight", memberID, commandID, revision
	return pacedomain.Pace{}, stub.err
}
func (stub *roomStub) ApplyPause(_ context.Context, _, memberID, commandID string, action pausedomain.Action, revision uint64) (pausedomain.Stone, error) {
	stub.called, stub.actor, stub.commandID, stub.action, stub.revision = "pause", memberID, commandID, action, revision
	return pausedomain.Stone{}, stub.err
}
func (stub *roomStub) SetHonesty(_ context.Context, _, memberID, commandID string, grant bool, revision uint64) (honestydomain.Ribbon, error) {
	stub.called, stub.actor, stub.commandID, stub.grant, stub.revision = "honesty", memberID, commandID, grant, revision
	return honestydomain.Ribbon{}, stub.err
}
func (stub *roomStub) Close(_ context.Context, _, memberID, commandID string, revision uint64) (closuredomain.Closure, error) {
	stub.called, stub.actor, stub.commandID, stub.revision = "closure", memberID, commandID, revision
	return closuredomain.Closure{}, stub.err
}
func (stub *roomStub) Block(_ context.Context, _, memberID, commandID string, revision uint64) (safetydomain.Safety, error) {
	stub.called, stub.actor, stub.commandID, stub.revision = "block", memberID, commandID, revision
	return safetydomain.Safety{}, stub.err
}
func (stub *roomStub) Report(_ context.Context, _, memberID, commandID string, category safetydomain.Category, _ string, revision uint64) (safetydomain.Safety, error) {
	stub.called, stub.actor, stub.commandID, stub.category, stub.revision = "report", memberID, commandID, category, revision
	return safetydomain.Safety{}, stub.err
}

func roomHandler(room CourtshipRoom, memberID string) http.Handler {
	mux := http.NewServeMux()
	RegisterCourtshipRoomRoutes(mux, room, sessionStub{memberID: memberID}, verifiedGate())
	return Correlation(mux)
}

func postRoom(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

// TestRoomActionsCarryTheActorAndRevision is what makes a shared room safe
// for two handsets: the actor comes from the session, and the revision the
// caller believed it was acting on reaches the aggregate.
func TestRoomActionsCarryTheActorAndRevision(t *testing.T) {
	cases := map[string]struct {
		path, body, want string
	}{
		"pace relight": {"/v1/courtship/rooms/room_1/pace/relight", `{"commandId":"c1","expectedRevision":4}`, "relight"},
		"pause":        {"/v1/courtship/rooms/room_1/pause", `{"commandId":"c1","action":"pause","expectedRevision":4}`, "pause"},
		"honesty":      {"/v1/courtship/rooms/room_1/honesty", `{"commandId":"c1","grant":true,"expectedRevision":4}`, "honesty"},
		"closure":      {"/v1/courtship/rooms/room_1/closure", `{"commandId":"c1","expectedRevision":4}`, "closure"},
		"block":        {"/v1/courtship/rooms/room_1/safety/block", `{"commandId":"c1","expectedRevision":4}`, "block"},
		"report":       {"/v1/courtship/rooms/room_1/safety/report", `{"commandId":"c1","category":"harassment","expectedRevision":4}`, "report"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			stub := &roomStub{}
			response := postRoom(t, roomHandler(stub, "mem_actor"), testCase.path, testCase.body)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
			}
			if stub.called != testCase.want {
				t.Errorf("called %q, want %q", stub.called, testCase.want)
			}
			if stub.commandID != "c1" || stub.revision != 4 {
				t.Errorf("command = %q revision = %d", stub.commandID, stub.revision)
			}
			if stub.actor != "mem_actor" {
				t.Errorf("actor = %q, want the session's member", stub.actor)
			}
		})
	}
}

func TestRoomActionsValidateTheirVocabulary(t *testing.T) {
	cases := map[string]string{
		"/v1/courtship/rooms/room_1/pause":         `{"commandId":"c1","action":"stop"}`,
		"/v1/courtship/rooms/room_1/safety/report": `{"commandId":"c1","category":"gossip"}`,
	}
	for path, body := range cases {
		stub := &roomStub{}
		response := postRoom(t, roomHandler(stub, "mem_actor"), path, body)
		if response.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s status = %d, want 422", path, response.Code)
		}
		if stub.called != "" {
			t.Errorf("%s reached the service with an invalid value", path)
		}
	}
}

// TestOutsidersCannotLearnARoomExists keeps "denied" answering 404 rather
// than 403: a 403 would confirm the room is real.
func TestOutsidersCannotLearnARoomExists(t *testing.T) {
	for _, cause := range []error{
		pacedomain.ErrDenied, pausedomain.ErrDenied, closuredomain.ErrDenied,
		honestydomain.ErrDenied, safetydomain.ErrDenied,
	} {
		stub := &roomStub{err: cause}
		response := postRoom(t, roomHandler(stub, "mem_outsider"),
			"/v1/courtship/rooms/room_1/closure", `{"commandId":"c1"}`)
		if response.Code != http.StatusNotFound {
			t.Errorf("%v mapped to %d, want 404", cause, response.Code)
		}
	}
}

func TestRoomConflictsAreMapped(t *testing.T) {
	cases := map[error]int{
		pausedomain.ErrSuspended:         http.StatusConflict,
		safetydomain.ErrContactBlocked:   http.StatusConflict,
		closuredomain.ErrNotDue:          http.StatusConflict,
		pacedomain.ErrStaleRevision:      http.StatusConflict,
		honestydomain.ErrCommandMismatch: http.StatusConflict,
		safetydomain.ErrInvalid:          http.StatusUnprocessableEntity,
	}
	for cause, want := range cases {
		stub := &roomStub{err: cause}
		response := postRoom(t, roomHandler(stub, "mem_actor"),
			"/v1/courtship/rooms/room_1/closure", `{"commandId":"c1"}`)
		if response.Code != want {
			t.Errorf("%v mapped to %d, want %d", cause, response.Code, want)
		}
	}
}

func TestRoomRoutesRequireASession(t *testing.T) {
	response := postRoom(t, roomHandler(&roomStub{}, ""),
		"/v1/courtship/rooms/room_1/closure", `{"commandId":"c1"}`)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", response.Code)
	}
}

func TestRoomRoutesRequireACommandID(t *testing.T) {
	stub := &roomStub{}
	response := postRoom(t, roomHandler(stub, "mem_actor"),
		"/v1/courtship/rooms/room_1/closure", `{}`)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	if stub.called != "" {
		t.Error("a command with no id reached the service")
	}
}

// TestStartRoomAlwaysIncludesTheCaller stops a member from opening a room
// between two other people.
func TestStartRoomAlwaysIncludesTheCaller(t *testing.T) {
	stub := &roomStub{}
	response := postRoom(t, roomHandler(stub, "mem_caller"), "/v1/courtship/rooms",
		`{"commandId":"c1","roomId":"room_9","counterpartId":"mem_other"}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.actor != "mem_caller" {
		t.Errorf("actor = %q, want the session's member", stub.actor)
	}
	if len(stub.members) != 2 || stub.members[0] != "mem_caller" || stub.members[1] != "mem_other" {
		t.Errorf("members = %v, want the caller and the counterpart", stub.members)
	}
}

func TestStartRoomRejectsSelfPairing(t *testing.T) {
	stub := &roomStub{}
	response := postRoom(t, roomHandler(stub, "mem_caller"), "/v1/courtship/rooms",
		`{"commandId":"c1","roomId":"room_9","counterpartId":"mem_caller"}`)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	if stub.called != "" {
		t.Error("a self-paired room reached the service")
	}
}

// TestSubmitTurnCarriesTheDeviceCursor is what keeps two handsets in one room
// consistent: a device that has fallen behind must be told to catch up rather
// than writing over turns it has not read.
func TestSubmitTurnCarriesTheDeviceCursor(t *testing.T) {
	stub := &roomStub{}
	response := postRoom(t, roomHandler(stub, "mem_actor"), "/v1/courtship/rooms/room_1/turns",
		`{"commandId":"c1","deviceRef":"dev_1","payloadRef":"pay_1","baseSequence":7}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.actor != "mem_actor" {
		t.Errorf("actor = %q, want the session's member", stub.actor)
	}
	if stub.device != "dev_1" || stub.payload != "pay_1" || stub.base != 7 {
		t.Errorf("submit = device %q payload %q base %d", stub.device, stub.payload, stub.base)
	}
}

func TestStaleDeviceIsToldToCatchUp(t *testing.T) {
	stub := &roomStub{err: queuedomain.ErrStaleDevice}
	response := postRoom(t, roomHandler(stub, "mem_actor"), "/v1/courtship/rooms/room_1/turns",
		`{"commandId":"c1","deviceRef":"dev_1","payloadRef":"pay_1","baseSequence":1}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", response.Code)
	}
}

func TestSubmitTurnValidatesItsReferences(t *testing.T) {
	for _, body := range []string{
		`{"commandId":"","deviceRef":"d","payloadRef":"p"}`,
		`{"commandId":"c","deviceRef":"","payloadRef":"p"}`,
		`{"commandId":"c","deviceRef":"d","payloadRef":""}`,
	} {
		stub := &roomStub{}
		response := postRoom(t, roomHandler(stub, "mem_actor"), "/v1/courtship/rooms/room_1/turns", body)
		if response.Code != http.StatusUnprocessableEntity {
			t.Errorf("body %s status = %d, want 422", body, response.Code)
		}
		if stub.called != "" {
			t.Errorf("body %s reached the service", body)
		}
	}
}

// TestTimelineReturnsOnlySequenceAndTime keeps room content out of the log a
// device reconciles against.
func TestTimelineReturnsOnlySequenceAndTime(t *testing.T) {
	stub := &roomStub{}
	request := httptest.NewRequest(http.MethodGet, "/v1/courtship/rooms/room_1/turns?after=4&limit=10", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	roomHandler(stub, "mem_actor").ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if stub.after != 4 || stub.limit != 10 {
		t.Errorf("after = %d limit = %d", stub.after, stub.limit)
	}
	body := response.Body.String()
	for _, leaked := range []string{"payload", "actorKey", "deviceKey", "fingerprint"} {
		if strings.Contains(body, leaked) {
			t.Errorf("timeline leaked %q: %s", leaked, body)
		}
	}
}

func TestTimelineRejectsBadPaging(t *testing.T) {
	for _, query := range []string{"?after=soon", "?limit=0", "?limit=101"} {
		request := httptest.NewRequest(http.MethodGet, "/v1/courtship/rooms/room_1/turns"+query, nil)
		request.Header.Set("Authorization", "Bearer token")
		response := httptest.NewRecorder()
		roomHandler(&roomStub{}, "mem_actor").ServeHTTP(response, request)
		if response.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s status = %d, want 422", query, response.Code)
		}
	}
}

// TestTimelineIsGatedOnMembership pins the fix for an IDOR found by review:
// the turn queue carries no membership of its own, so reading its log has to
// be gated at the composition seam. Without it any member who learned a room
// id could read another couple's activity.
func TestTimelineIsGatedOnMembership(t *testing.T) {
	stub := &roomStub{}
	request := httptest.NewRequest(http.MethodGet, "/v1/courtship/rooms/room_1/turns", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	roomHandler(stub, "mem_reader").ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	// The member must reach the gate; a handler that drops it cannot enforce
	// membership no matter what the seam does.
	if stub.actor != "mem_reader" {
		t.Errorf("timeline called for %q, want the session's member", stub.actor)
	}
}

// TestOutsiderIsRefusedTheTimeline covers the seam's own answer.
func TestOutsiderIsRefusedTheTimeline(t *testing.T) {
	stub := &roomStub{err: pacedomain.ErrDenied}
	request := httptest.NewRequest(http.MethodGet, "/v1/courtship/rooms/room_1/turns", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	roomHandler(stub, "mem_outsider").ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 so an outsider cannot tell the room is real", response.Code)
	}
}
