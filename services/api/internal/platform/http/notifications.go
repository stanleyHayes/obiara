package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/notifications/domain"
)

// Notifications is the inbound port for notification preferences (E13-S01).
type Notifications interface {
	Get(ctx context.Context, memberID string) (domain.Preferences, error)
	Configure(ctx context.Context, memberID string, muted map[domain.Category]bool, quietStart, quietEnd int, timezone string) (domain.Preferences, error)
}

// RegisterNotificationRoutes adds notification preference routes.
func RegisterNotificationRoutes(mux *http.ServeMux, notifications Notifications) {
	mux.Handle("GET /v1/notification-preferences/{memberId}", getPreferencesHandler(notifications))
	mux.Handle("PUT /v1/notification-preferences/{memberId}", putPreferencesHandler(notifications))
}

type preferencesResponse struct {
	Muted      map[string]bool `json:"muted"`
	QuietStart int             `json:"quietStart"`
	QuietEnd   int             `json:"quietEnd"`
	Timezone   string          `json:"timezone"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func toPreferencesResponse(preferences domain.Preferences) preferencesResponse {
	muted := make(map[string]bool, len(preferences.Muted()))
	for category, value := range preferences.Muted() {
		muted[string(category)] = value
	}
	return preferencesResponse{
		Muted:      muted,
		QuietStart: preferences.QuietStart(),
		QuietEnd:   preferences.QuietEnd(),
		Timezone:   preferences.Timezone(),
		UpdatedAt:  preferences.UpdatedAt(),
	}
}

func getPreferencesHandler(notifications Notifications) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		preferences, err := notifications.Get(r.Context(), r.PathValue("memberId"))
		if err != nil {
			writeNotificationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPreferencesResponse(preferences))
	})
}

type configurePreferencesRequest struct {
	Muted      map[string]bool `json:"muted"`
	QuietStart int             `json:"quietStart"`
	QuietEnd   int             `json:"quietEnd"`
	Timezone   string          `json:"timezone"`
}

func putPreferencesHandler(notifications Notifications) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body configurePreferencesRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		memberID := r.PathValue("memberId")
		body.Timezone = strings.TrimSpace(body.Timezone)
		var details []FieldError
		if !validOpaqueID(memberID) {
			details = append(details, FieldError{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		for category := range body.Muted {
			switch domain.Category(category) {
			case domain.CategoryRitual, domain.CategoryPods, domain.CategoryRooms, domain.CategorySafety:
			default:
				details = append(details, FieldError{Field: "muted", Reason: "unknown category " + category})
			}
		}
		if body.Timezone == "" {
			details = append(details, FieldError{Field: "timezone", Reason: "must be an IANA time zone"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		muted := make(map[domain.Category]bool, len(body.Muted))
		for category, value := range body.Muted {
			muted[domain.Category(category)] = value
		}
		preferences, err := notifications.Configure(r.Context(), memberID, muted, body.QuietStart, body.QuietEnd, body.Timezone)
		if err != nil {
			writeNotificationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPreferencesResponse(preferences))
	})
}

func writeNotificationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrSafetyCannotBeMuted):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "safety_cannot_be_muted",
			Message: "Safety notifications cannot be turned off.",
		})
	case errors.Is(err, domain.ErrInvalidTimezone):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "The time zone must be a valid IANA name.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
