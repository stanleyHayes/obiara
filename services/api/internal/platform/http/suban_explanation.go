package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	explanationapp "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/application"
	explanationdomain "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
)

type SubanExplanation interface {
	Explain(context.Context, string, string) (explanationdomain.Explanation, error)
	File(context.Context, string, string, string, explanationdomain.Reason, string) (explanationdomain.Appeal, error)
}

type subanVisibleEventResponse struct {
	ID             string    `json:"id"`
	Kind           string    `json:"kind"`
	Effect         string    `json:"effect"`
	SourceCategory string    `json:"sourceCategory"`
	OccurredAt     time.Time `json:"occurredAt"`
}

type subanExplanationResponse struct {
	Marks       []string                    `json:"marks"`
	Events      []subanVisibleEventResponse `json:"events"`
	GeneratedAt time.Time                   `json:"generatedAt"`
}

func subanExplanationHandler(service SubanExplanation, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		explanation, err := service.Explain(r.Context(), subject, subject)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "suban_unavailable", Message: "Your Suban explanation is temporarily unavailable."})
			return
		}
		response := subanExplanationResponse{
			Marks:       make([]string, 0, len(explanation.Marks)),
			Events:      make([]subanVisibleEventResponse, 0, len(explanation.Events)),
			GeneratedAt: explanation.GeneratedAt,
		}
		for _, mark := range explanation.Marks {
			response.Marks = append(response.Marks, string(mark))
		}
		for _, event := range explanation.Events {
			response.Events = append(response.Events, subanVisibleEventResponse{
				ID: event.ID, Kind: string(event.Kind), Effect: string(event.Effect),
				SourceCategory: event.SourceCategory, OccurredAt: event.OccurredAt,
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

type subanAppealRequest struct {
	EventID string `json:"eventId"`
	Reason  string `json:"reason"`
}

type subanAppealResponse struct {
	AppealID string    `json:"appealId"`
	EventID  string    `json:"eventId"`
	Status   string    `json:"status"`
	FiledAt  time.Time `json:"filedAt"`
}

func subanAppealHandler(service SubanExplanation, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{Code: "unsupported_media_type", Message: "Content-Type must be application/json."})
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		var body subanAppealRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if !validIdentifier(commandID) || strings.TrimSpace(body.EventID) == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "An event, reason and valid Idempotency-Key are required."})
			return
		}
		appeal, err := service.File(
			r.Context(), subject, subject, strings.TrimSpace(body.EventID),
			explanationdomain.Reason(body.Reason), commandID,
		)
		if errors.Is(err, explanationapp.ErrInvalid) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "appeal_not_allowed", Message: "Only an adverse event in your record can be appealed."})
			return
		}
		if errors.Is(err, explanationapp.ErrApplied) || errors.Is(err, explanationapp.ErrConflict) {
			writeError(w, r, http.StatusConflict, APIError{Code: "appeal_exists", Message: "This event already has an appeal."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "suban_unavailable", Message: "The appeal could not be filed."})
			return
		}
		state := appeal.State()
		writeSuccess(w, r, http.StatusCreated, subanAppealResponse{
			AppealID: state.ID, EventID: state.EventID, Status: string(state.Status), FiledAt: state.FiledAt,
		})
	})
}
