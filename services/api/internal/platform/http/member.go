package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/member/application"
	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

const (
	maxRequestBodyBytes = 1 << 20
	maxIdentifierLength = 128
	maxEmailLength      = 254
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)

// RegisterMember is the inbound port needed by the member HTTP adapter.
type RegisterMember func(context.Context, application.RegisterMemberCommand) (domain.Member, error)

// RegisterMemberRoutes adds the member-facing baseline route to mux.
func RegisterMemberRoutes(mux *http.ServeMux, register RegisterMember) {
	mux.Handle("POST /v1/members", registerMemberHandler(register))
}

type registerMemberRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type memberResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func registerMemberHandler(register RegisterMember) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if register == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code:    "service_unavailable",
				Message: "The service is temporarily unavailable.",
			})
			return
		}

		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var request registerMemberRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		request.ID = strings.TrimSpace(request.ID)
		request.Email = strings.ToLower(strings.TrimSpace(request.Email))
		if details := validateRegisterMember(request, r.Header.Get("Idempotency-Key")); len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		member, err := register(r.Context(), application.RegisterMemberCommand{
			ID:             request.ID,
			Email:          request.Email,
			IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		})
		if err != nil {
			writeApplicationError(w, r, err)
			return
		}

		w.Header().Set("Location", "/v1/members/"+member.ID())
		writeSuccess(w, r, http.StatusCreated, memberResponse{
			ID:        member.ID(),
			Email:     member.Email(),
			CreatedAt: member.CreatedAt(),
		})
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func validateRegisterMember(request registerMemberRequest, idempotencyKey string) []FieldError {
	var details []FieldError
	if request.ID == "" || len(request.ID) > maxIdentifierLength || !identifierPattern.MatchString(request.ID) {
		details = append(details, FieldError{Field: "id", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
	}
	if !validEmail(request.Email) {
		details = append(details, FieldError{Field: "email", Reason: "must be a valid email address"})
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || len(idempotencyKey) > maxIdentifierLength || !identifierPattern.MatchString(idempotencyKey) {
		details = append(details, FieldError{Field: "Idempotency-Key", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
	}
	return details
}

func validEmail(value string) bool {
	if value == "" || len(value) > maxEmailLength || strings.ContainsAny(value, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(value, "@")
}

func writeApplicationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrDuplicateMember):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "member_conflict",
			Message: "A member with this identifier already exists.",
		})
	case errors.Is(err, application.ErrIdempotencyKeyRequired),
		errors.Is(err, domain.ErrInvalidID),
		errors.Is(err, domain.ErrInvalidEmail):
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
