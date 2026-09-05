package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
)

// Listening is the inbound port for playback telemetry (E06-S03).
type Listening interface {
	RecordHeartbeats(ctx context.Context, listenerID, assetID string, assetDuration float64, ranges []application.HeartbeatRange) (domain.Playback, error)
	Eligibility(ctx context.Context, listenerID, assetID string) (eligible bool, totalSeconds float64, err error)
}

// RegisterListeningRoutes adds playback telemetry and eligibility routes.
func RegisterListeningRoutes(mux *http.ServeMux, listening Listening, sessions SessionAuthenticator, gate MemberGate) {
	// Listening to a Voice of Introduction is a romantic surface (FR-101a).
	mux.Handle("POST /v1/listening/heartbeats", gate.guard(sessions, "introductions.view", "introduction", recordHeartbeatsHandler(listening, sessions)))
	mux.Handle("GET /v1/listening/eligibility/{assetId}", gate.guard(sessions, "introductions.view", "introduction", eligibilityHandler(listening, sessions)))
}

type heartbeatsRequest struct {
	VoiceAssetID  string                       `json:"voiceAssetId"`
	AssetDuration float64                      `json:"assetDuration"`
	Ranges        []application.HeartbeatRange `json:"ranges"`
}

type eligibilityResponse struct {
	Eligible        bool    `json:"eligible"`
	TotalSeconds    float64 `json:"totalSeconds"`
	RequiredSeconds float64 `json:"requiredSeconds"`
}

func recordHeartbeatsHandler(listening Listening, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listenerID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body heartbeatsRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.VoiceAssetID = strings.TrimSpace(body.VoiceAssetID)
		var details []FieldError
		if !validOpaqueID(body.VoiceAssetID) {
			details = append(details, FieldError{Field: "voiceAssetId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.AssetDuration <= 0 || body.AssetDuration > 3600 {
			details = append(details, FieldError{Field: "assetDuration", Reason: "must be between 0 and 3600 seconds"})
		}
		if len(body.Ranges) == 0 || len(body.Ranges) > 100 {
			details = append(details, FieldError{Field: "ranges", Reason: "must contain 1-100 ranges"})
		}
		for _, heartbeat := range body.Ranges {
			if heartbeat.Start < 0 || heartbeat.End <= heartbeat.Start {
				details = append(details, FieldError{Field: "ranges", Reason: "each range must satisfy 0 <= start < end"})
				break
			}
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		record, err := listening.RecordHeartbeats(r.Context(), listenerID, body.VoiceAssetID, body.AssetDuration, body.Ranges)
		if err != nil {
			writeListeningError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, eligibilityResponse{
			Eligible:        record.Eligible(),
			TotalSeconds:    record.TotalSeconds(),
			RequiredSeconds: domain.RequiredSeconds,
		})
	})
}

func eligibilityHandler(listening Listening, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listenerID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		assetID := r.PathValue("assetId")
		var details []FieldError
		if !validOpaqueID(assetID) {
			details = append(details, FieldError{Field: "assetId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		eligible, total, err := listening.Eligibility(r.Context(), listenerID, assetID)
		if err != nil {
			writeListeningError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, eligibilityResponse{
			Eligible:        eligible,
			TotalSeconds:    total,
			RequiredSeconds: domain.RequiredSeconds,
		})
	})
}

func validOpaqueID(value string) bool {
	return value != "" && len(value) <= maxIdentifierLength && identifierPattern.MatchString(value)
}

func writeListeningError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRange), errors.Is(err, domain.ErrInvalidDuration),
		errors.Is(err, domain.ErrListenerRequired), errors.Is(err, domain.ErrAssetRequired):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
