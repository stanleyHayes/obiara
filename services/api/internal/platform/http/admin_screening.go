package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/reviewdesk"
	screeningmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/mongodb"
	screeningdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/screening/domain"
)

// AdminScreeningQueue is the list of sows waiting on a person.
type AdminScreeningQueue interface {
	// Pending takes the reading agent because the read is audited. There is
	// no way to see a member's words without leaving a record of who looked.
	Pending(ctx context.Context, actorID string, limit int) ([]screeningmongo.Pending, error)
}

// AdminScreeningDesk settles one review: the sow first, then the record.
type AdminScreeningDesk interface {
	Decide(ctx context.Context, reference string, approve bool, actorID, commandID string) error
}

// RegisterAdminScreeningRoutes exposes the queue every sow passes through.
//
// Step-up is required for both, because the queue holds members' own words —
// the whole point of the review is that somebody reads them, and that is not
// something an operator should reach without re-asserting who they are.
func RegisterAdminScreeningRoutes(mux *http.ServeMux, queue AdminScreeningQueue, desk AdminScreeningDesk, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/screening/reviews", adminScreeningQueueHandler(queue, resolve))
	mux.Handle("POST /v1/admin/screening/reviews/{id}/decision", adminScreeningDecisionHandler(desk, resolve))
}

type screeningReviewResponse struct {
	Reference string `json:"reference"`
	Reason    string `json:"reason"`
	Body      string `json:"body"`
	LocaleTag string `json:"localeTag,omitempty"`
	// Recordings describes the media without carrying it: a reviewer needs
	// to know what is attached, and this response is not the place to hand
	// the audio around.
	Recordings []screeningMediaResponse `json:"recordings"`
	// Advisory is what the automated pass thought, shown as an opinion. It
	// is never a decision — a decision here is only ever a person's.
	Advisory []string `json:"advisory,omitempty"`
	RoutedAt string   `json:"routedAt"`
}

type screeningMediaResponse struct {
	MediaType  string `json:"mediaType"`
	Bytes      int64  `json:"bytes"`
	DurationMs int64  `json:"durationMs"`
}

type screeningQueueResponse struct {
	Reviews []screeningReviewResponse `json:"reviews"`
}

func adminScreeningQueueHandler(queue AdminScreeningQueue, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, true)
		if !ok {
			return
		}
		if queue == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		pending, err := queue.Pending(r.Context(), principal.ActorID, limit)
		if err != nil {
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}
		response := screeningQueueResponse{Reviews: make([]screeningReviewResponse, 0, len(pending))}
		for _, item := range pending {
			entry := screeningReviewResponse{
				Reference:  item.Review.ID(),
				Reason:     item.Review.Reason(),
				Body:       item.Text,
				LocaleTag:  item.LocaleTag,
				Advisory:   item.AdvisoryReasons,
				RoutedAt:   item.Review.RoutedAt().UTC().Format(time.RFC3339),
				Recordings: make([]screeningMediaResponse, 0, len(item.Media)),
			}
			for _, media := range item.Media {
				entry.Recordings = append(entry.Recordings, screeningMediaResponse{
					MediaType: media.MIME, Bytes: media.Bytes, DurationMs: media.DurationMs,
				})
			}
			response.Reviews = append(response.Reviews, entry)
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

type screeningDecisionInput struct {
	// Approve is required rather than defaulted: a missing field must not
	// mean "let it through".
	Approve *bool `json:"approve"`
}

func adminScreeningDecisionHandler(desk AdminScreeningDesk, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if commandID == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
			})
			return
		}
		var body screeningDecisionInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if body.Approve == nil {
			// An absent decision is not an approval. Defaulting here would
			// deliver a sow nobody actually cleared.
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "approve", Reason: "is required"}},
			})
			return
		}
		if desk == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		err := desk.Decide(r.Context(), r.PathValue("id"), *body.Approve, principal.ActorID, commandID)
		switch {
		case err == nil:
			writeSuccess(w, r, http.StatusOK, map[string]bool{"approved": *body.Approve})
		case errors.Is(err, reviewdesk.ErrNotFound):
			writeError(w, r, http.StatusNotFound, APIError{
				Code: "review_not_found", Message: "That review is not available.",
			})
		case errors.Is(err, screeningdomain.ErrNotPending):
			writeError(w, r, http.StatusConflict, APIError{
				Code: "review_already_decided", Message: "That review has already been decided.",
			})
		default:
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
		}
	})
}
