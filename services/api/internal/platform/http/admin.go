package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

// Admin is the inbound port for admin auth (E16-S01).
type Admin interface {
	Enroll(ctx context.Context, actorID, email string, roles []domain.Role) (domain.Principal, error)
	StartLogin(ctx context.Context, email string) error
	CompleteLogin(ctx context.Context, email, code string) (domain.Session, error)
	StepUpStart(ctx context.Context, sessionID string) error
	StepUpComplete(ctx context.Context, sessionID, code string) (domain.Session, error)
}

// RegisterAdminRoutes adds the admin auth baseline routes.
func RegisterAdminRoutes(mux *http.ServeMux, admin Admin) {
	mux.Handle("POST /v1/admin/principals", enrollHandler(admin))
	mux.Handle("POST /v1/admin/login/start", loginStartHandler(admin))
	mux.Handle("POST /v1/admin/login/complete", loginCompleteHandler(admin))
	mux.Handle("POST /v1/admin/sessions/{id}/step-up/start", stepUpStartHandler(admin))
	mux.Handle("POST /v1/admin/sessions/{id}/step-up/complete", stepUpCompleteHandler(admin))
}

func adminJSONGuard(w http.ResponseWriter, r *http.Request) bool {
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{
			Code:    "unsupported_media_type",
			Message: "Content-Type must be application/json.",
		})
		return false
	}
	return true
}

type enrollRequest struct {
	ActorID string   `json:"actorId"`
	Email   string   `json:"email"`
	Roles   []string `json:"roles"`
}

type principalResponse struct {
	PrincipalID string    `json:"principalId"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	CreatedAt   time.Time `json:"createdAt"`
}

func enrollHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body enrollRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		body.ActorID = strings.TrimSpace(body.ActorID)
		body.Email = strings.TrimSpace(body.Email)
		if !validOpaqueID(body.ActorID) || len(body.Roles) == 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "actorId/roles", Reason: "actorId must be an opaque id and at least one role is required"}},
			})
			return
		}
		roles := make([]domain.Role, 0, len(body.Roles))
		for _, role := range body.Roles {
			roles = append(roles, domain.Role(role))
		}
		principal, err := admin.Enroll(r.Context(), body.ActorID, body.Email, roles)
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		roleNames := make([]string, 0, len(principal.Roles()))
		for _, role := range principal.Roles() {
			roleNames = append(roleNames, string(role))
		}
		writeSuccess(w, r, http.StatusCreated, principalResponse{
			PrincipalID: principal.ID(), Email: principal.Email(), Roles: roleNames, CreatedAt: principal.CreatedAt(),
		})
	})
}

type emailRequest struct {
	Email string `json:"email"`
}

func loginStartHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body emailRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if strings.TrimSpace(body.Email) == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "One or more fields are invalid."})
			return
		}
		if err := admin.StartLogin(r.Context(), strings.TrimSpace(body.Email)); err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, map[string]string{"status": "code_sent"})
	})
}

type loginCompleteRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type adminSessionResponse struct {
	SessionID string    `json:"sessionId"`
	Roles     []string  `json:"roles"`
	SteppedUp bool      `json:"steppedUp"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func toSessionResponse(session domain.Session) adminSessionResponse {
	roles := make([]string, 0, len(session.Roles()))
	for _, role := range session.Roles() {
		roles = append(roles, string(role))
	}
	return adminSessionResponse{SessionID: session.ID(), Roles: roles, SteppedUp: session.SteppedUp(), ExpiresAt: session.ExpiresAt()}
}

func loginCompleteHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body loginCompleteRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		session, err := admin.CompleteLogin(r.Context(), strings.TrimSpace(body.Email), strings.TrimSpace(body.Code))
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSessionResponse(session))
	})
}

func stepUpStartHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		if err := admin.StepUpStart(r.Context(), r.PathValue("id")); err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, map[string]string{"status": "code_sent"})
	})
}

type stepUpCompleteRequest struct {
	Code string `json:"code"`
}

func stepUpCompleteHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body stepUpCompleteRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		session, err := admin.StepUpComplete(r.Context(), r.PathValue("id"), strings.TrimSpace(body.Code))
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSessionResponse(session))
	})
}

func writeAdminError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotAdmin):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "admin_role_required",
			Message: "Enrollment requires the admin role.",
		})
	case errors.Is(err, application.ErrPrincipalExists):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "principal_exists",
			Message: "A principal with that email already exists.",
		})
	case errors.Is(err, domain.ErrMfaMismatch), errors.Is(err, domain.ErrMfaConsumed), errors.Is(err, application.ErrChallengeNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_invalid",
			Message: "The code is not valid for this account.",
		})
	case errors.Is(err, domain.ErrMfaExpired):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_expired",
			Message: "The code has expired. Request a new one.",
		})
	case errors.Is(err, domain.ErrMfaAttempts):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_attempts_exceeded",
			Message: "Too many wrong attempts. Request a new code.",
		})
	case errors.Is(err, domain.ErrSessionClosed), errors.Is(err, application.ErrSessionNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "session_closed",
			Message: "The admin session is not active.",
		})
	case errors.Is(err, application.ErrPrincipalNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "principal_not_found",
			Message: "No admin principal for that account.",
		})
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrNoRoles):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
