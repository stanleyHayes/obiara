package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
)

// ConsentMap is the inbound port for the consent switchboard (Doc 08 §8).
type ConsentMap interface {
	Switchboard(ctx context.Context, memberID string) (map[domain.Purpose]bool, error)
	Set(ctx context.Context, memberID string, purpose domain.Purpose, enable bool) (bool, error)
}

// RegisterConsentRoutes adds the consent switchboard routes.
func RegisterConsentRoutes(mux *http.ServeMux, consentMap ConsentMap) {
	mux.Handle("GET /v1/consent/{memberId}", switchboardHandler(consentMap))
	mux.Handle("PUT /v1/consent/{memberId}/{purpose}", setConsentHandler(consentMap))
}

func switchboardHandler(consentMap ConsentMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		board, err := consentMap.Switchboard(r.Context(), r.PathValue("memberId"))
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		purposes := make(map[string]bool, len(board))
		for purpose, enabled := range board {
			purposes[string(purpose)] = enabled
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Purposes map[string]bool `json:"purposes"`
		}{Purposes: purposes})
	})
}

type setConsentRequest struct {
	Enabled bool `json:"enabled"`
}

type consentStateResponse struct {
	Purpose string `json:"purpose"`
	Enabled bool   `json:"enabled"`
}

func setConsentHandler(consentMap ConsentMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}
		var body setConsentRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		purpose := domain.Purpose(r.PathValue("purpose"))
		enabled, err := consentMap.Set(r.Context(), r.PathValue("memberId"), purpose, body.Enabled)
		if err != nil {
			writeConsentError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, consentStateResponse{Purpose: string(purpose), Enabled: enabled})
	})
}

func writeConsentError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidPurpose):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "unknown_purpose",
			Message: "That consent purpose does not exist.",
		})
	case errors.Is(err, domain.ErrPurposeLocked):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "purpose_locked",
			Message: "Identity and safety processing is required for the service and cannot be turned off.",
		})
	case errors.Is(err, domain.ErrWrongDirection):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "purpose_direction",
			Message: "That purpose only allows one direction of change.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
