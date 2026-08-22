package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	pushdomain "github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

// PushRegistry is the inbound port for device push registration.
type PushRegistry interface {
	Register(ctx context.Context, memberID, token string, platform pushdomain.Platform) error
	Forget(ctx context.Context, memberID string) error
}

// RegisterPushRoutes adds device push registration routes. A device registers
// after sign-in and deregisters at sign-out, so a shared handset stops
// receiving the previous member's notifications.
func RegisterPushRoutes(mux *http.ServeMux, push PushRegistry, sessions SessionAuthenticator) {
	mux.Handle("PUT /v1/push-devices", registerPushDeviceHandler(push, sessions))
	mux.Handle("DELETE /v1/push-devices", forgetPushDevicesHandler(push, sessions))
}

type pushDeviceRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func authenticatedMember(w http.ResponseWriter, r *http.Request, sessions SessionAuthenticator) (string, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok || sessions == nil {
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code: "authentication_required", Message: "A valid member session is required.",
		})
		return "", false
	}
	session, err := sessions.Authenticate(r.Context(), token)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code: "authentication_required", Message: "A valid member session is required.",
		})
		return "", false
	}
	return session.MemberID(), true
}

func registerPushDeviceHandler(push PushRegistry, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		// The member is taken from the session, never the body: a device must
		// not be able to register itself against another member's id.
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}

		var body pushDeviceRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}

		err := push.Register(r.Context(), memberID,
			strings.TrimSpace(body.Token), pushdomain.Platform(strings.ToLower(strings.TrimSpace(body.Platform))))
		switch {
		case err == nil:
			writeSuccess(w, r, http.StatusOK, map[string]string{"status": "registered"})
		case errors.Is(err, pushdomain.ErrTokenInvalid):
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "token", Reason: "must be an Expo push token"}},
			})
		case errors.Is(err, pushdomain.ErrPlatformInvalid):
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "platform", Reason: "must be ios, android or web"}},
			})
		default:
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
		}
	})
}

func forgetPushDevicesHandler(push PushRegistry, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if err := push.Forget(r.Context(), memberID); err != nil {
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]string{"status": "forgotten"})
	})
}
