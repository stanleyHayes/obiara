package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

// Sows is the inbound port for the sow: the atomic gesture a member makes
// toward another, carrying their answer and costing a seed.
type Sows interface {
	Send(context.Context, sowapplication.Command) (sowapplication.Result, error)
}

// RegisterSowRoutes exposes sending a sow.
//
// Gated at the sowing rung like every other reach toward a person (FR-101b).
// There is no route to read somebody else's sow here on purpose: a sow is
// delivered through the pod at the house front, not fetched by whoever knows
// an id.
func RegisterSowRoutes(mux *http.ServeMux, sows Sows, sessions SessionAuthenticator, gate MemberGate) {
	mux.Handle("POST /v1/seed/sows", gate.guard(sessions, "seeds.sow", "seed", sendSowHandler(sows, sessions)))
}

type sendSowRequest struct {
	Body      string   `json:"body"`
	MediaRefs []string `json:"mediaRefs,omitempty"`
	// Confirmed is the deliberate gesture, not a checkbox the client can
	// default to true. The domain refuses without it because a sow costs a
	// seed and reaches a person, and neither should happen by brushing a
	// screen.
	Confirmed bool `json:"confirmed"`
}

type sowResponse struct {
	SowID string `json:"sowId"`
	// Status is pending_review while a person reads it. The member is told
	// plainly rather than shown a delivery that has not happened.
	Status   string `json:"status"`
	Replayed bool   `json:"replayed"`
}

func sendSowHandler(sows Sows, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		// The sower is the session. A member cannot sow on somebody else's
		// behalf, and the seed comes out of their own allowance.
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
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
		var body sendSowRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if sows == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}

		result, err := sows.Send(r.Context(), sowapplication.Command{
			ID: commandID, ActorID: actorID, Body: body.Body,
			MediaRefs: body.MediaRefs, Confirmed: body.Confirmed,
		})
		if err != nil {
			writeSowError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, sowResponse{
			SowID:    result.Sow.ID,
			Status:   string(result.Sow.Status),
			Replayed: result.Replayed,
		})
	})
}

func writeSowError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sowdomain.ErrNotConfirmed):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "confirmation_required",
			Message: "Hold to send. A sow costs a seed and reaches a person.",
		})
	case errors.Is(err, sowapplication.ErrMediaNotOwned):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "recording_not_yours",
			Message: "You can only send a recording you made yourself.",
		})
	case errors.Is(err, sowapplication.ErrInsufficientAllowance):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "no_seeds_left",
			Message: "You have used this week's seeds. More arrive at the start of next week.",
		})
	case errors.Is(err, sowdomain.ErrScreeningRejected):
		// Deliberately says nothing about what was wrong with it. A refusal
		// that explains itself teaches someone how to word the next one.
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "sow_not_sent",
			Message: "This could not be sent.",
		})
	case errors.Is(err, sowdomain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, sowapplication.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "feature_unavailable", Message: "This is not available right now.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
