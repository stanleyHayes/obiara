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

// RegisterAuthRoutes adds the authentication baseline routes to mux.
func RegisterAuthRoutes(mux *http.ServeMux, registration Registration) {
	mux.Handle("POST /v1/auth/otp", requestOtpHandler(registration))
	mux.Handle("POST /v1/auth/otp/verify", verifyOtpHandler(registration))
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
