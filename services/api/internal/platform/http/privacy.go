package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/privacy/application"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/domain"
)

// Privacy is the inbound port for data-subject requests (E03-S10).
type Privacy interface {
	RequestExport(ctx context.Context, accountID string) (domain.PrivacyRequest, error)
	RequestDeletion(ctx context.Context, accountID string) (domain.PrivacyRequest, error)
	Status(ctx context.Context, requestID string) (domain.PrivacyRequest, error)
}

// RegisterPrivacyRoutes adds the privacy baseline routes to mux.
func RegisterPrivacyRoutes(mux *http.ServeMux, privacy Privacy) {
	mux.Handle("POST /v1/privacy/exports", openPrivacyRequestHandler(privacy, domain.KindExport))
	mux.Handle("POST /v1/privacy/deletions", openPrivacyRequestHandler(privacy, domain.KindDeletion))
	mux.Handle("GET /v1/privacy/requests/{id}", privacyStatusHandler(privacy))
}

type privacyRequestBody struct {
	AccountID string `json:"accountId"`
}

type privacyRequestResponse struct {
	RequestID   string     `json:"requestId"`
	Kind        string     `json:"kind"`
	Status      string     `json:"status"`
	DueAt       time.Time  `json:"dueAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

func openPrivacyRequestHandler(privacy Privacy, kind domain.Kind) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body privacyRequestBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.AccountID = strings.TrimSpace(body.AccountID)
		if body.AccountID == "" || len(body.AccountID) > maxIdentifierLength || !identifierPattern.MatchString(body.AccountID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "accountId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}

		var request domain.PrivacyRequest
		var err error
		if kind == domain.KindExport {
			request, err = privacy.RequestExport(r.Context(), body.AccountID)
		} else {
			request, err = privacy.RequestDeletion(r.Context(), body.AccountID)
		}
		if err != nil {
			writePrivacyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toPrivacyResponse(request))
	})
}

func privacyStatusHandler(privacy Privacy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := privacy.Status(r.Context(), r.PathValue("id"))
		if err != nil {
			writePrivacyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPrivacyResponse(request))
	})
}

func toPrivacyResponse(request domain.PrivacyRequest) privacyRequestResponse {
	return privacyRequestResponse{
		RequestID:   request.ID(),
		Kind:        string(request.Kind()),
		Status:      string(request.Status()),
		DueAt:       request.DueAt(),
		CompletedAt: request.CompletedAt(),
	}
}

func writePrivacyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrOpenRequestExists):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "privacy_request_exists",
			Message: "An open request of this kind already exists.",
		})
	case errors.Is(err, domain.ErrLegalHoldActive):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "legal_hold_active",
			Message: "This account's data is preserved and cannot be deleted right now.",
		})
	case errors.Is(err, application.ErrRequestNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "privacy_request_not_found",
			Message: "No such privacy request.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
