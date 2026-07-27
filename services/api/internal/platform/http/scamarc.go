package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/application"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/domain"
)

// ScamArc is the inbound port for scam-arc signals (E11-S11).
type ScamArc interface {
	Observe(ctx context.Context, roomID, actorID string, kind domain.SignalKind) (domain.RoomState, *application.EducationCard, error)
}

// RegisterScamArcRoutes adds the scam-arc signal route. Producers are
// server-side classifiers; the route exists for pipeline wiring and tests.
func RegisterScamArcRoutes(mux *http.ServeMux, scamArc ScamArc) {
	mux.Handle("POST /v1/scam-arc/signals", observeSignalHandler(scamArc))
}

type observeSignalRequest struct {
	RoomID  string `json:"roomId"`
	ActorID string `json:"actorId"`
	Kind    string `json:"kind"`
}

type scamArcStateResponse struct {
	Ladder  string `json:"ladder"`
	Educate bool   `json:"educate"`
}

func observeSignalHandler(scamArc ScamArc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}
		var body observeSignalRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.RoomID = strings.TrimSpace(body.RoomID)
		body.ActorID = strings.TrimSpace(body.ActorID)
		if !validOpaqueID(body.RoomID) || !validOpaqueID(body.ActorID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "roomId/actorId", Reason: "must be opaque identifiers"}},
			})
			return
		}
		state, card, err := scamArc.Observe(r.Context(), body.RoomID, body.ActorID, domain.SignalKind(body.Kind))
		if err != nil {
			writeScamArcError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, scamArcStateResponse{Ladder: string(state.Ladder), Educate: card != nil})
	})
}

func writeScamArcError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrMonitoringOptedOut):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "monitoring_opted_out",
			Message: "This room has opted out of scam-arc monitoring.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
