package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// Registration is the inbound port for phone OTP registration (E03-S01).
type Registration interface {
	RequestOtp(context.Context, string) (application.OtpRequest, error)
	VerifyOtp(ctx context.Context, phone, code, deviceID string) (application.IssuedSession, error)
}

// Sessions is the inbound port for session rotation. Access tokens live
// fifteen minutes; without this port a member would have to complete the SMS
// OTP flow again every fifteen minutes, at the cost of one SMS each time.
type Sessions interface {
	Refresh(ctx context.Context, refreshToken string) (application.IssuedSession, error)
}

// RegisterAuthRoutes adds the authentication baseline routes to mux.
func RegisterAuthRoutes(mux *http.ServeMux, registration Registration, sessions Sessions) {
	mux.Handle("POST /v1/auth/otp", requestOtpHandler(registration))
	mux.Handle("POST /v1/auth/otp/verify", verifyOtpHandler(registration))
	mux.Handle("POST /v1/auth/refresh", refreshHandler(sessions))
}

type refreshBody struct {
	RefreshToken string `json:"refreshToken"`
}

// refreshHandler rotates a refresh token into a fresh token pair.
//
// Rotation is single-use: presenting a token that has already been rotated
// out is treated as theft by the session service, which revokes the whole
// session. That is why every failure here answers with the same
// unauthorized envelope — distinguishing "expired" from "stolen" would tell
// an attacker which of the two they hold.
func refreshHandler(sessions Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body refreshBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.RefreshToken = strings.TrimSpace(body.RefreshToken)
		if body.RefreshToken == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "refreshToken", Reason: "is required"}},
			})
			return
		}

		issued, err := sessions.Refresh(r.Context(), body.RefreshToken)
		if err != nil {
			writeRefreshError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, sessionResponse{
			SessionID:        issued.Session.ID(),
			MemberID:         issued.Session.MemberID(),
			AccessToken:      issued.AccessToken,
			AccessExpiresAt:  issued.Session.AccessExpiresAt(),
			RefreshToken:     issued.RefreshToken,
			RefreshExpiresAt: issued.Session.RefreshExpiresAt(),
		})
	})
}

// writeRefreshError answers every rejected rotation identically. Only a
// genuine fault is separated out, so it can be logged and investigated.
func writeRefreshError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrRefreshReuse),
		errors.Is(err, domain.ErrRefreshTokenMismatch),
		errors.Is(err, domain.ErrSessionExpired),
		errors.Is(err, domain.ErrSessionNotActive),
		errors.Is(err, domain.ErrTokenMalformed),
		errors.Is(err, application.ErrSessionNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "refresh_invalid",
			Message: "Your sign-in has expired. Please sign in again.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}

type requestOtpBody struct {
	Phone string `json:"phone"`
}

type otpRequestResponse struct {
	ChallengeID string    `json:"challengeId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type verifyOtpBody struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	DeviceID string `json:"deviceId"`
}

type sessionResponse struct {
	SessionID        string    `json:"sessionId"`
	MemberID         string    `json:"memberId"`
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refreshToken"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

func requestOtpHandler(registration Registration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body requestOtpBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.Phone = strings.TrimSpace(body.Phone)
		if !e164(body.Phone) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "phone", Reason: "must be an E.164 phone number"}},
			})
			return
		}

		request, err := registration.RequestOtp(r.Context(), body.Phone)
		if err != nil {
			writeAuthError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, otpRequestResponse{
			ChallengeID: request.ChallengeID,
			ExpiresAt:   request.ExpiresAt,
		})
	})
}

func verifyOtpHandler(registration Registration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body verifyOtpBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.Phone = strings.TrimSpace(body.Phone)
		body.Code = strings.TrimSpace(body.Code)
		body.DeviceID = strings.TrimSpace(body.DeviceID)
		var details []FieldError
		if !e164(body.Phone) {
			details = append(details, FieldError{Field: "phone", Reason: "must be an E.164 phone number"})
		}
		if len(body.Code) != 6 {
			details = append(details, FieldError{Field: "code", Reason: "must be the 6-digit code"})
		}
		if body.DeviceID == "" || len(body.DeviceID) > maxIdentifierLength || !identifierPattern.MatchString(body.DeviceID) {
			details = append(details, FieldError{Field: "deviceId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		issued, err := registration.VerifyOtp(r.Context(), body.Phone, body.Code, body.DeviceID)
		if err != nil {
			writeAuthError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, sessionResponse{
			SessionID:        issued.Session.ID(),
			MemberID:         issued.Session.MemberID(),
			AccessToken:      issued.AccessToken,
			AccessExpiresAt:  issued.Session.AccessExpiresAt(),
			RefreshToken:     issued.RefreshToken,
			RefreshExpiresAt: issued.Session.RefreshExpiresAt(),
		})
	})
}

var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

func e164(phone string) bool {
	return e164Pattern.MatchString(phone)
}

func writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrOtpRateLimited):
		writeError(w, r, http.StatusTooManyRequests, APIError{
			Code:    "otp_rate_limited",
			Message: "Too many code requests. Please wait and try again.",
		})
	case errors.Is(err, domain.ErrOtpExpired):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "otp_expired",
			Message: "The code has expired. Request a new one.",
		})
	case errors.Is(err, domain.ErrOtpAttemptsExceeded):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "otp_attempts_exceeded",
			Message: "Too many wrong attempts. Request a new code.",
		})
	case errors.Is(err, domain.ErrOtpMismatch),
		errors.Is(err, domain.ErrOtpConsumed),
		errors.Is(err, domain.ErrTokenMalformed),
		errors.Is(err, application.ErrChallengeNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "otp_invalid",
			Message: "The code is not valid for this number.",
		})
	case errors.Is(err, domain.ErrAccountNotUsable):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "account_not_active",
			Message: "This account is not active.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
