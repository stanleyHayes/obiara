package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/domain"
)

// Embers is the inbound port for ember issuance and redemption (E06-S10).
type Embers interface {
	Issue(ctx context.Context, fireID, fromID, toID string) (domain.Ember, error)
	Redeem(ctx context.Context, emberID, memberID string) (domain.Ember, error)
}

// RegisterEmberRoutes adds ember routes.
func RegisterEmberRoutes(mux *http.ServeMux, embers Embers, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/fires/{id}/embers", issueEmberHandler(embers, sessions))
	mux.Handle("POST /v1/embers/{id}/redeem", redeemEmberHandler(embers, sessions))
}

type issueEmberRequest struct {
	ToID string `json:"toId"`
}

type emberResponse struct {
	EmberID    string     `json:"emberId"`
	Status     string     `json:"status"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	RedeemedAt *time.Time `json:"redeemedAt,omitempty"`
}

func toEmberResponse(ember domain.Ember) emberResponse {
	return emberResponse{
		EmberID:    ember.ID(),
		Status:     string(ember.Status()),
		ExpiresAt:  ember.ExpiresAt(),
		RedeemedAt: ember.RedeemedAt(),
	}
}

func issueEmberHandler(embers Embers, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromID, ok := subanSubject(w, r, sessions)
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

		var body issueEmberRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.ToID = strings.TrimSpace(body.ToID)
		var details []FieldError
		if !validOpaqueID(body.ToID) {
			details = append(details, FieldError{Field: "toId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if fromID == body.ToID {
			details = append(details, FieldError{Field: "toId", Reason: "must differ from fromId"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		ember, err := embers.Issue(r.Context(), r.PathValue("id"), fromID, body.ToID)
		if err != nil {
			writeEmberError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toEmberResponse(ember))
	})
}

func redeemEmberHandler(embers Embers, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
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

		var body struct{}
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		ember, err := embers.Redeem(r.Context(), r.PathValue("id"), memberID)
		if err != nil {
			writeEmberError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toEmberResponse(ember))
	})
}

func writeEmberError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotCoAttendee):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "not_co_attendee",
			Message: "Embers are only for people who attended the same fire.",
		})
	case errors.Is(err, application.ErrNotRecipient):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "not_recipient",
			Message: "Only the recipient can redeem an ember.",
		})
	case errors.Is(err, application.ErrEmberAlreadyGiven):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "ember_already_given",
			Message: "You already gave an ember at this fire.",
		})
	case errors.Is(err, domain.ErrEmberExpired):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "ember_expired",
			Message: "This ember's 24-hour window has closed.",
		})
	case errors.Is(err, domain.ErrEmberNotOpen):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "ember_not_open",
			Message: "This ember is no longer open.",
		})
	case errors.Is(err, application.ErrEmberNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "ember_not_found",
			Message: "No such ember.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
