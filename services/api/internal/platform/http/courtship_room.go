package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	closuredomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	honestydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	pacedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	pausedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	queueapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/application"
	queuedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
	safetydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
)

// CourtshipRoom is the inbound port for the mechanics of a courtship room:
// the rhythm the pair keeps, the pause either may call, the honesty ribbon,
// closure, and the in-room safety actions.
type CourtshipRoom interface {
	Submit(ctx context.Context, roomID, deviceRef, actorID, payloadRef, commandID string, baseSequence uint64) (queueapp.Result, error)
	Timeline(ctx context.Context, roomID, memberID string, after uint64, limit int) ([]queuedomain.Event, error)
	Start(ctx context.Context, roomID string, members []string, commandID, actorID string) (pacedomain.Pace, error)
	Relight(ctx context.Context, roomID, memberID, commandID string, expectedRevision uint64) (pacedomain.Pace, error)
	ApplyPause(ctx context.Context, roomID, memberID, commandID string, action pausedomain.Action, expectedRevision uint64) (pausedomain.Stone, error)
	SetHonesty(ctx context.Context, roomID, memberID, commandID string, grant bool, expectedRevision uint64) (honestydomain.Ribbon, error)
	Close(ctx context.Context, roomID, memberID, commandID string, expectedRevision uint64) (closuredomain.Closure, error)
	Block(ctx context.Context, roomID, memberID, commandID string, expectedRevision uint64) (safetydomain.Safety, error)
	Report(ctx context.Context, roomID, memberID, commandID string, category safetydomain.Category, evidenceRef string, expectedRevision uint64) (safetydomain.Safety, error)
}

// RegisterCourtshipRoomRoutes adds the room routes.
//
// Every action carries a command id and the revision the caller believes it
// is acting on. A courtship room is two people acting on one shared state
// from two handsets, so a retried tap must not apply twice and a race must
// conflict rather than let the later write win silently.
func RegisterCourtshipRoomRoutes(mux *http.ServeMux, room CourtshipRoom, sessions SessionAuthenticator, gate MemberGate) {
	// What advances a courtship is gated at Tier 1 (FR-101a). What ends one,
	// slows it, or reports it never is: a member demoted to Tier 0 for safety
	// must still be able to pause, close, block and report, or the gate would
	// shut them in a room with the person they need to get away from.
	mux.Handle("POST /v1/courtship/rooms", gate.guard(sessions, "rooms.participate", "room", startRoomHandler(room, sessions)))
	mux.Handle("POST /v1/courtship/rooms/{id}/turns", gate.guard(sessions, "rooms.participate", "room", submitTurnHandler(room, sessions)))
	mux.Handle("GET /v1/courtship/rooms/{id}/turns", roomTimelineHandler(room, sessions))
	mux.Handle("POST /v1/courtship/rooms/{id}/pace/relight", gate.guard(sessions, "rooms.participate", "room", paceRelightHandler(room, sessions)))
	mux.Handle("POST /v1/courtship/rooms/{id}/pause", pauseHandler(room, sessions))
	mux.Handle("POST /v1/courtship/rooms/{id}/honesty", gate.guard(sessions, "rooms.participate", "room", honestyHandler(room, sessions)))
	mux.Handle("POST /v1/courtship/rooms/{id}/closure", closureHandler(room, sessions))
	mux.Handle("POST /v1/courtship/rooms/{id}/safety/block", safetyBlockHandler(room, sessions))
	mux.Handle("POST /v1/courtship/rooms/{id}/safety/report", safetyReportHandler(room, sessions))
}

type roomCommandRequest struct {
	CommandID        string `json:"commandId"`
	ExpectedRevision uint64 `json:"expectedRevision"`
	// Action is used by the pause endpoint: pause, acknowledge or resume.
	Action string `json:"action"`
	// Grant is used by the honesty endpoint.
	Grant bool `json:"grant"`
	// Category and EvidenceRef are used by the report endpoint. The evidence
	// reference is opaque; no report content crosses this boundary.
	Category    string `json:"category"`
	EvidenceRef string `json:"evidenceRef"`
}

type roomStateResponse struct {
	RoomID   string `json:"roomId"`
	Revision uint64 `json:"revision"`
	Status   string `json:"status,omitempty"`
}

// roomCommand parses and authenticates the parts every room action shares.
func roomCommand(w http.ResponseWriter, r *http.Request, sessions SessionAuthenticator) (memberID, roomID string, body roomCommandRequest, ok bool) {
	if !proposalJSONGuard(w, r) {
		return "", "", body, false
	}
	memberID, ok = authenticatedMember(w, r, sessions)
	if !ok {
		return "", "", body, false
	}
	roomID = r.PathValue("id")
	if !validOpaqueID(roomID) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
		return "", "", body, false
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, http.StatusBadRequest, APIError{
			Code: "invalid_json", Message: "The request body must be one valid JSON object.",
		})
		return "", "", body, false
	}
	body.CommandID = strings.TrimSpace(body.CommandID)
	if !validOpaqueID(body.CommandID) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
			Details: []FieldError{{Field: "commandId", Reason: "must be an opaque identifier"}},
		})
		return "", "", body, false
	}
	return memberID, roomID, body, true
}

type startRoomRequest struct {
	CommandID     string `json:"commandId"`
	RoomID        string `json:"roomId"`
	CounterpartID string `json:"counterpartId"`
}

// startRoomHandler opens a room between the caller and one counterpart. The
// caller is always one of the two members, taken from the session, so a
// member cannot open a room between two other people.
func startRoomHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		var body startRoomRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.RoomID = strings.TrimSpace(body.RoomID)
		body.CounterpartID = strings.TrimSpace(body.CounterpartID)

		var details []FieldError
		if !validOpaqueID(body.CommandID) {
			details = append(details, FieldError{Field: "commandId", Reason: "must be an opaque identifier"})
		}
		if !validOpaqueID(body.RoomID) {
			details = append(details, FieldError{Field: "roomId", Reason: "must be an opaque identifier"})
		}
		if !validOpaqueID(body.CounterpartID) {
			details = append(details, FieldError{Field: "counterpartId", Reason: "must be an opaque identifier"})
		}
		if body.CounterpartID == memberID {
			details = append(details, FieldError{Field: "counterpartId", Reason: "cannot be yourself"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		pace, err := room.Start(r.Context(), body.RoomID,
			[]string{memberID, body.CounterpartID}, body.CommandID, memberID)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, roomStateResponse{
			RoomID: body.RoomID, Revision: pace.Revision(), Status: string(pace.Status()),
		})
	})
}

type submitTurnRequest struct {
	CommandID string `json:"commandId"`
	DeviceRef string `json:"deviceRef"`
	// PayloadRef is opaque. The turn itself is content the room holds; only
	// its reference crosses this boundary.
	PayloadRef string `json:"payloadRef"`
	// BaseSequence is the last event this device has seen. A device that has
	// fallen behind is told to catch up rather than writing over what it has
	// not read.
	BaseSequence uint64 `json:"baseSequence"`
}

type turnResponse struct {
	Sequence uint64 `json:"sequence"`
	Replayed bool   `json:"replayed"`
}

type timelineEntry struct {
	Sequence   uint64 `json:"sequence"`
	AcceptedAt string `json:"acceptedAt"`
}

type timelineResponse struct {
	Events []timelineEntry `json:"events"`
}

func submitTurnHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		roomID := r.PathValue("id")
		if !validOpaqueID(roomID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body submitTurnRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.DeviceRef = strings.TrimSpace(body.DeviceRef)
		body.PayloadRef = strings.TrimSpace(body.PayloadRef)

		var details []FieldError
		for field, value := range map[string]string{
			"commandId": body.CommandID, "deviceRef": body.DeviceRef, "payloadRef": body.PayloadRef,
		} {
			if !validOpaqueID(value) {
				details = append(details, FieldError{Field: field, Reason: "must be an opaque identifier"})
			}
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		result, err := room.Submit(r.Context(), roomID, body.DeviceRef, memberID,
			body.PayloadRef, body.CommandID, body.BaseSequence)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, turnResponse{
			Sequence: result.Event.Sequence, Replayed: result.Replayed,
		})
	})
}

// roomTimelineHandler returns the ordered log both devices reconcile against.
// Only sequence and time are returned: the turn content stays in the room.
func roomTimelineHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		roomID := r.PathValue("id")
		if !validOpaqueID(roomID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		after, limit := uint64(0), 50
		if raw := strings.TrimSpace(r.URL.Query().Get("after")); raw != "" {
			parsed, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "after", Reason: "must be a sequence number"}},
				})
				return
			}
			after = parsed
		}
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 100 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "limit", Reason: "must be between 1 and 100"}},
				})
				return
			}
			limit = parsed
		}

		events, err := room.Timeline(r.Context(), roomID, memberID, after, limit)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		response := timelineResponse{Events: make([]timelineEntry, 0, len(events))}
		for _, event := range events {
			response.Events = append(response.Events, timelineEntry{
				Sequence: event.Sequence, AcceptedAt: event.AcceptedAt.UTC().Format(time.RFC3339),
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func paceRelightHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		pace, err := room.Relight(r.Context(), roomID, memberID, body.CommandID, body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{
			RoomID: roomID, Revision: pace.Revision(), Status: string(pace.Status()),
		})
	})
}

func pauseHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		action := pausedomain.Action(strings.ToLower(strings.TrimSpace(body.Action)))
		switch action {
		case pausedomain.ActionPause, pausedomain.ActionAck, pausedomain.ActionResume:
		default:
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "action", Reason: "must be pause, acknowledge or resume"}},
			})
			return
		}
		stone, err := room.ApplyPause(r.Context(), roomID, memberID, body.CommandID, action, body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{
			RoomID: roomID, Revision: stone.Revision(), Status: string(stone.Status()),
		})
	})
}

func honestyHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		ribbon, err := room.SetHonesty(r.Context(), roomID, memberID, body.CommandID, body.Grant, body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		// The ribbon becomes visible only when both members have granted it,
		// which is the whole point of the mechanic.
		status := "hidden"
		if ribbon.Visible() {
			status = "visible"
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{
			RoomID: roomID, Revision: ribbon.Revision(), Status: status,
		})
	})
}

func closureHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		closure, err := room.Close(r.Context(), roomID, memberID, body.CommandID, body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{
			RoomID: roomID, Revision: closure.Revision(), Status: string(closure.Status()),
		})
	})
}

func safetyBlockHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		safety, err := room.Block(r.Context(), roomID, memberID, body.CommandID, body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{RoomID: roomID, Revision: safety.Revision()})
	})
}

func safetyReportHandler(room CourtshipRoom, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, roomID, body, ok := roomCommand(w, r, sessions)
		if !ok {
			return
		}
		category := safetydomain.Category(strings.ToLower(strings.TrimSpace(body.Category)))
		switch category {
		case safetydomain.CategoryHarassment, safetydomain.CategoryIdentity,
			safetydomain.CategoryThreat, safetydomain.CategoryOther:
		default:
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "category", Reason: "must be harassment, identity, threat or other"}},
			})
			return
		}
		safety, err := room.Report(r.Context(), roomID, memberID, body.CommandID,
			category, strings.TrimSpace(body.EvidenceRef), body.ExpectedRevision)
		if err != nil {
			writeCourtshipRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, roomStateResponse{RoomID: roomID, Revision: safety.Revision()})
	})
}

// writeCourtshipRoomError maps the family's failures.
//
// Every slice uses the same vocabulary, and "denied" deliberately answers 404
// rather than 403: a member who is not in a room must not be able to learn
// that it exists.
func writeCourtshipRoomError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, pacedomain.ErrDenied), errors.Is(err, pausedomain.ErrDenied),
		errors.Is(err, closuredomain.ErrDenied), errors.Is(err, honestydomain.ErrDenied),
		errors.Is(err, safetydomain.ErrDenied):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "room_not_found", Message: "That room is not available.",
		})
	case errors.Is(err, pausedomain.ErrSuspended):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "room_paused", Message: "This room is paused.",
		})
	case errors.Is(err, safetydomain.ErrContactBlocked):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "contact_blocked", Message: "Contact in this room is blocked.",
		})
	case errors.Is(err, closuredomain.ErrNotDue):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "closure_not_due", Message: "This room is not yet inactive enough to close.",
		})
	case errors.Is(err, pacedomain.ErrStaleRevision), errors.Is(err, pausedomain.ErrStaleRevision),
		errors.Is(err, closuredomain.ErrStaleRevision), errors.Is(err, honestydomain.ErrStaleRevision),
		errors.Is(err, safetydomain.ErrStaleRevision):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "room_conflict", Message: "This room changed. Refresh and try again.",
		})
	case errors.Is(err, pacedomain.ErrCommandMismatch), errors.Is(err, pausedomain.ErrCommandMismatch),
		errors.Is(err, closuredomain.ErrCommandMismatch), errors.Is(err, honestydomain.ErrCommandMismatch),
		errors.Is(err, safetydomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, queuedomain.ErrNotYourTurn):
		// A product rule, not a fault. Without this case it would fall to the
		// default and be reported as a server error, which is both wrong and
		// the kind of thing a member retries until something breaks.
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "not_your_turn",
			Message: "You have already spoken. Wait for them to answer.",
		})
	case errors.Is(err, queuedomain.ErrStaleDevice):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "device_behind", Message: "This device is behind. Catch up and try again.",
		})
	case errors.Is(err, queuedomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, queuedomain.ErrInvalid),
		errors.Is(err, pacedomain.ErrInvalid), errors.Is(err, pausedomain.ErrInvalid),
		errors.Is(err, closuredomain.ErrInvalid), errors.Is(err, honestydomain.ErrInvalid),
		errors.Is(err, safetydomain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
