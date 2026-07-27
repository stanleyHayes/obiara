package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/calls/application"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/domain"
)

// Calls is the inbound port for in-app calls (E09-S09).
type Calls interface {
	Initiate(ctx context.Context, roomID, initiatorID, otherID string) (application.IssuedCall, error)
	End(ctx context.Context, callID, actorID string) error
}

// RegisterCallRoutes adds the call routes.
func RegisterCallRoutes(mux *http.ServeMux, calls Calls) {
	mux.Handle("POST /v1/rooms/{roomId}/calls", initiateCallHandler(calls))
	mux.Handle("POST /v1/calls/{id}/end", endCallHandler(calls))
}

type initiateCallRequest struct {
	InitiatorID string `json:"initiatorId"`
	OtherID     string `json:"otherId"`
}

type tokenResponse struct {
	Signed    string    `json:"signed"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type initiateCallResponse struct {
	CallID string                   `json:"callId"`
	Tokens map[string]tokenResponse `json:"tokens"`
}

func initiateCallHandler(calls Calls) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}
		var body initiateCallRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.InitiatorID = strings.TrimSpace(body.InitiatorID)
		body.OtherID = strings.TrimSpace(body.OtherID)
		var details []FieldError
		if !validOpaqueID(body.InitiatorID) {
			details = append(details, FieldError{Field: "initiatorId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if !validOpaqueID(body.OtherID) {
			details = append(details, FieldError{Field: "otherId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.InitiatorID != "" && body.InitiatorID == body.OtherID {
			details = append(details, FieldError{Field: "otherId", Reason: "must differ from initiatorId"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		issued, err := calls.Initiate(r.Context(), r.PathValue("roomId"), body.InitiatorID, body.OtherID)
		if err != nil {
			writeCallError(w, r, err)
			return
		}
		response := initiateCallResponse{CallID: issued.Call.ID(), Tokens: map[string]tokenResponse{}}
		for participant, token := range issued.Tokens {
			response.Tokens[participant] = tokenResponse{Signed: token.Signed, ExpiresAt: token.ExpiresAt}
		}
		writeSuccess(w, r, http.StatusCreated, response)
	})
}

type endCallRequest struct {
	ActorID string `json:"actorId"`
}

func endCallHandler(calls Calls) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}
		var body endCallRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.ActorID = strings.TrimSpace(body.ActorID)
		if !validOpaqueID(body.ActorID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "actorId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		if err := calls.End(r.Context(), r.PathValue("id"), body.ActorID); err != nil {
			writeCallError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]string{"status": "ended"})
	})
}

func writeCallError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotRoomMember):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "not_room_member",
			Message: "Calls are only between the two room members.",
		})
	case errors.Is(err, application.ErrNotParticipant):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "not_participant",
			Message: "Only a call participant can end the call.",
		})
	case errors.Is(err, application.ErrCallNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "call_not_found",
			Message: "No such call.",
		})
	case errors.Is(err, domain.ErrCallNotOpen):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "call_not_open",
			Message: "This call is already over.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
