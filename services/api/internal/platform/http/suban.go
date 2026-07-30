package apihttp

import (
	"context"
	"net/http"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

// Suban is the inbound port for the member-visible character ledger
// (E15-S04; Doc 08 §4: members view every event behind their marks).
type Suban interface {
	Marks(ctx context.Context, subjectID string) ([]domain.Mark, error)
	Events(ctx context.Context, subjectID string) ([]domain.Event, error)
}

// RegisterSubanRoutes adds the suban marks and events routes.
func RegisterSubanRoutes(mux *http.ServeMux, suban Suban, explanation SubanExplanation, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/suban/marks/{memberId}", subanMarksHandler(suban, sessions))
	mux.Handle("GET /v1/suban/events/{memberId}", subanEventsHandler(suban, sessions))
	mux.Handle("GET /v1/suban/explanation", subanExplanationHandler(explanation, sessions))
	mux.Handle("POST /v1/suban/appeals", subanAppealHandler(explanation, sessions))
}

type marksResponse struct {
	Marks []string `json:"marks"`
}

func subanSubject(w http.ResponseWriter, r *http.Request, sessions SessionAuthenticator) (string, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok || sessions == nil {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "authentication_required", Message: "A valid member session is required."})
		return "", false
	}
	session, err := sessions.Authenticate(r.Context(), token)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "authentication_required", Message: "A valid member session is required."})
		return "", false
	}
	return session.MemberID(), true
}

func subanMarksHandler(suban Suban, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		subject, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		if subject != r.PathValue("memberId") {
			writeError(w, r, http.StatusForbidden, APIError{Code: "access_denied", Message: "Suban records belong to another member."})
			return
		}
		marks, err := suban.Marks(r.Context(), subject)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		labels := make([]string, 0, len(marks))
		for _, mark := range marks {
			labels = append(labels, string(mark))
		}
		writeSuccess(w, r, http.StatusOK, marksResponse{Marks: labels})
	})
}

type subanEventResponse struct {
	Kind       string    `json:"kind"`
	Source     string    `json:"source"`
	Ref        string    `json:"ref"`
	OccurredAt time.Time `json:"occurredAt"`
}

type eventsResponse struct {
	Events []subanEventResponse `json:"events"`
}

func subanEventsHandler(suban Suban, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		subject, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		if subject != r.PathValue("memberId") {
			writeError(w, r, http.StatusForbidden, APIError{Code: "access_denied", Message: "Suban records belong to another member."})
			return
		}
		events, err := suban.Events(r.Context(), subject)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		response := eventsResponse{Events: make([]subanEventResponse, 0, len(events))}
		for _, event := range events {
			response.Events = append(response.Events, subanEventResponse{
				Kind:   string(event.Kind),
				Source: event.Provenance.Source,
				// Never project the source-system reference back to a member.
				// The event ID is sufficient for explanation and appeal flows.
				Ref:        event.ID,
				OccurredAt: event.OccurredAt,
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}
