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
func RegisterSubanRoutes(mux *http.ServeMux, suban Suban) {
	mux.Handle("GET /v1/suban/marks/{memberId}", subanMarksHandler(suban))
	mux.Handle("GET /v1/suban/events/{memberId}", subanEventsHandler(suban))
}

type marksResponse struct {
	Marks []string `json:"marks"`
}

func subanMarksHandler(suban Suban) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		marks, err := suban.Marks(r.Context(), r.PathValue("memberId"))
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

func subanEventsHandler(suban Suban) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validOpaqueID(r.PathValue("memberId")) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		events, err := suban.Events(r.Context(), r.PathValue("memberId"))
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		response := eventsResponse{Events: make([]subanEventResponse, 0, len(events))}
		for _, event := range events {
			response.Events = append(response.Events, subanEventResponse{
				Kind:       string(event.Kind),
				Source:     event.Provenance.Source,
				Ref:        event.Provenance.Ref,
				OccurredAt: event.OccurredAt,
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}
