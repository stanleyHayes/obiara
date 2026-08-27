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

// Registration is the inbound port for OTP registration (E03-S01), over
// whichever channel the member verified.
type Registration interface {
	RequestOtp(context.Context, domain.Contact) (application.OtpRequest, error)
	VerifyOtp(ctx context.Context, contact domain.Contact, code, deviceID string) (application.IssuedSession, error)
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
	// Channel is "sms" or "email" and defaults to "sms" when absent, so
	// clients written before email sign-in existed keep working untouched.
	Channel string `json:"channel"`
	// Contact is the address or number for that channel. Phone is the
	// pre-channel spelling of the same field and is still accepted.
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
}

type otpRequestResponse struct {
	ChallengeID string    `json:"challengeId"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type verifyOtpBody struct {
	Channel  string `json:"channel"`
	Contact  string `json:"contact"`
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

		contact, fieldErr := resolveContact(body.Channel, body.Contact, body.Phone)
		if fieldErr != nil {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{*fieldErr},
			})
			return
		}

		request, err := registration.RequestOtp(r.Context(), contact)
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

		body.Code = strings.TrimSpace(body.Code)
		body.DeviceID = strings.TrimSpace(body.DeviceID)
		var details []FieldError
		contact, fieldErr := resolveContact(body.Channel, body.Contact, body.Phone)
		if fieldErr != nil {
			details = append(details, *fieldErr)
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

		issued, err := registration.VerifyOtp(r.Context(), contact, body.Code, body.DeviceID)
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

// e164 validates a bare phone number for handlers that take one directly
// rather than as a channelled contact.
func e164(phone string) bool {
	return e164Pattern.MatchString(phone)
}

// resolveContact builds the identity a request is about.
//
// "contact" is the current field; "phone" is what every client sent before
// channels existed and still means an SMS contact. Accepting both keeps the
// mobile app and any deployed web build working across the rollout instead
// of breaking sign-in for everyone who has not updated.
func resolveContact(channelValue, contactValue, phoneValue string) (domain.Contact, *FieldError) {
	value := strings.TrimSpace(contactValue)
	field := "contact"
	if value == "" {
		value = strings.TrimSpace(phoneValue)
		field = "phone"
	}

	channel, err := domain.ParseChannel(channelValue)
	if err != nil {
		return domain.Contact{}, &FieldError{Field: "channel", Reason: `must be "sms" or "email"`}
	}

	contact, err := domain.NewContact(channel, value)
	if err != nil {
		reason := "must be an E.164 phone number"
		if channel == domain.ChannelEmail {
			reason = "must be a valid email address"
		}
		return domain.Contact{}, &FieldError{Field: field, Reason: reason}
	}
	return contact, nil
}

func writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrChannelUnavailable):
		// Checked before the delivery case, which this error also carries:
		// the member picked a channel this deployment cannot send on, and
		// the only useful thing to tell them is to pick the other one.
		logServerError(r.Context(), r, http.StatusUnprocessableEntity, "otp_channel_unavailable", err)
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "otp_channel_unavailable",
			Message: "That sign-in method is not available. Please use the other one.",
			Details: []FieldError{{Field: "channel", Reason: "not configured for this deployment"}},
		})
	case errors.Is(err, application.ErrCodeDeliveryFailed):
		// Minted but undeliverable. Retrying cannot help until the SMS
		// provider is fixed, so this is reported as an unavailable
		// dependency and the real cause goes to the logs.
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "otp_delivery_failed", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code:    "otp_delivery_failed",
			Message: "We could not send your code right now. Please try again shortly.",
		})
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
		// Unmapped errors are genuine faults. The envelope stays opaque; the
		// cause goes to the logs so a 500 is never a dead end during triage.
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
