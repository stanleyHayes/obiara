package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	declineapp "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/application"
	declinedomain "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
	sproutapp "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/application"
	sproutdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
)

// SeedStage is the inbound port for the seed stage: reaching toward someone,
// the bounded exchange that follows, and declining.
type SeedStage interface {
	Sprout(context.Context, sproutapp.SproutCommand) (sproutapp.SproutResult, error)
	Exchange(context.Context, sproutapp.ExchangeCommand) (sproutapp.ExchangeResult, error)
	Decline(context.Context, declineapp.Command) (declineapp.Result, error)
}

// RegisterSeedStageRoutes adds the seed-stage routes.
func RegisterSeedStageRoutes(mux *http.ServeMux, stage SeedStage, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/seed/sprouts", sproutHandler(stage, sessions))
	mux.Handle("POST /v1/seed/doorways/{id}/exchanges", exchangeHandler(stage, sessions))
	mux.Handle("POST /v1/seed/declines", declineHandler(stage, sessions))
}

type sproutRequest struct {
	CommandID string `json:"commandId"`
	TargetID  string `json:"targetId"`
	SeedRef   string `json:"seedRef"`
}

type exchangeRequest struct {
	CommandID  string `json:"commandId"`
	MessageRef string `json:"messageRef"`
}

type declineRequest struct {
	CommandID string `json:"commandId"`
	SeedID    string `json:"seedId"`
	OwnerID   string `json:"ownerId"`
}

type doorwayResponse struct {
	DoorwayID string `json:"doorwayId,omitempty"`
	Revision  uint64 `json:"revision"`
	// Opened reports that the other member had already reached toward this
	// one, so the doorway is now open between them.
	Opened   bool `json:"opened"`
	Replayed bool `json:"replayed"`
}

func sproutHandler(stage SeedStage, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		// The actor is the session: a member cannot reach toward someone on
		// another member's behalf.
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		var body sproutRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.TargetID = strings.TrimSpace(body.TargetID)
		body.SeedRef = strings.TrimSpace(body.SeedRef)

		var details []FieldError
		for field, value := range map[string]string{
			"commandId": body.CommandID, "targetId": body.TargetID, "seedRef": body.SeedRef,
		} {
			if !validOpaqueID(value) {
				details = append(details, FieldError{Field: field, Reason: "must be an opaque identifier"})
			}
		}
		if body.TargetID == actorID {
			details = append(details, FieldError{Field: "targetId", Reason: "cannot be yourself"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		result, err := stage.Sprout(r.Context(), sproutapp.SproutCommand{
			ID: body.CommandID, ActorID: actorID, TargetID: body.TargetID, SeedRef: body.SeedRef,
		})
		if err != nil {
			writeSeedStageError(w, r, err)
			return
		}

		// A doorway comes back only once both members have reached toward
		// each other. Until then the response says nothing about the other
		// member, not even that they exist.
		response := doorwayResponse{Replayed: result.Replayed}
		if result.Doorway != nil {
			response.DoorwayID = result.Doorway.ID()
			response.Revision = result.Doorway.Revision()
			response.Opened = true
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, response)
	})
}

func exchangeHandler(stage SeedStage, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		doorwayID := r.PathValue("id")
		if !validOpaqueID(doorwayID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body exchangeRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.MessageRef = strings.TrimSpace(body.MessageRef)
		if !validOpaqueID(body.CommandID) || !validOpaqueID(body.MessageRef) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}

		result, err := stage.Exchange(r.Context(), sproutapp.ExchangeCommand{
			ID: body.CommandID, DoorwayID: doorwayID, ActorID: actorID, MessageRef: body.MessageRef,
		})
		if err != nil {
			writeSeedStageError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, doorwayResponse{
			DoorwayID: result.Doorway.ID(), Revision: result.Doorway.Revision(),
			Opened: true, Replayed: result.Replayed,
		})
	})
}

func declineHandler(stage SeedStage, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		declinerID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		var body declineRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.SeedID = strings.TrimSpace(body.SeedID)
		body.OwnerID = strings.TrimSpace(body.OwnerID)
		if !validOpaqueID(body.CommandID) || !validOpaqueID(body.SeedID) || !validOpaqueID(body.OwnerID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}

		result, err := stage.Decline(r.Context(), declineapp.Command{
			ID: body.CommandID, DeclinerID: declinerID, SeedID: body.SeedID, OwnerID: body.OwnerID,
		})
		if err != nil {
			writeSeedStageError(w, r, err)
			return
		}
		// A decline is deliberately uninformative in both directions: the
		// response says only that it was recorded.
		writeSuccess(w, r, http.StatusOK, map[string]bool{"recorded": true, "replayed": result.Replayed})
	})
}

// writeSeedStageError maps the stage's failures.
//
// A doorway the caller is not part of answers exactly as a missing one does.
// Telling them apart would let anyone probe who has reached toward whom,
// which is the single most sensitive fact this stage holds.
func writeSeedStageError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sproutapp.ErrNotFound), errors.Is(err, sproutdomain.ErrNotParticipant):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "doorway_not_found", Message: "That doorway is not available.",
		})
	case errors.Is(err, sproutdomain.ErrDoorwaySealed):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "doorway_sealed", Message: "That doorway is closed.",
		})
	case errors.Is(err, sproutdomain.ErrOutOfTurn):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "out_of_turn", Message: "It is the other person's turn.",
		})
	case errors.Is(err, sproutapp.ErrConcurrentChange):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "doorway_conflict", Message: "That doorway changed. Refresh and try again.",
		})
	case errors.Is(err, sproutdomain.ErrCommandMismatch), errors.Is(err, declinedomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, sproutdomain.ErrInvalid), errors.Is(err, declinedomain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, sproutapp.ErrUnavailable), errors.Is(err, declineapp.ErrUnavailable):
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
