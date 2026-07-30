package apihttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

type AmpeRounds interface {
	Create(context.Context, application.Command) (application.Projection, error)
	Apply(context.Context, application.Command, domain.Action, domain.Choice) (application.Projection, error)
	View(context.Context, application.Command) (application.Projection, error)
}
type AmpePresence interface {
	Observe(context.Context, application.Command) (application.Projection, error)
}

func RegisterAmpeRoutes(
	mux *http.ServeMux,
	rounds AmpeRounds,
	presence AmpePresence,
	pairs OwarePairResolver,
	auth SessionAuthenticator,
) {
	mux.Handle("POST /v1/circles/{circleId}/ampe", createAmpeHandler(rounds, pairs, auth))
	mux.Handle("GET /v1/circles/{circleId}/ampe/{roundId}", viewAmpeHandler(presence, pairs, auth))
	mux.Handle("POST /v1/circles/{circleId}/ampe/{roundId}/commands", commandAmpeHandler(rounds, pairs, auth))
}

type ampePlayerResponse struct {
	Ready     bool `json:"ready"`
	Connected bool `json:"connected"`
	Locked    bool `json:"locked"`
}

type ampeResponse struct {
	ID          string             `json:"id"`
	Sequence    uint64             `json:"sequence"`
	You         ampePlayerResponse `json:"you"`
	Other       ampePlayerResponse `json:"other"`
	Paused      bool               `json:"paused"`
	OwnChoice   *domain.Choice     `json:"ownChoice,omitempty"`
	YourReveal  *domain.Choice     `json:"yourReveal,omitempty"`
	OtherReveal *domain.Choice     `json:"otherReveal,omitempty"`
	Complete    bool               `json:"complete"`
}

func projectAmpe(round application.Projection) ampeResponse {
	return ampeResponse{
		ID: round.ID, Sequence: round.Sequence,
		You: ampePlayerResponse(round.You), Other: ampePlayerResponse(round.Other),
		Paused: round.Paused, OwnChoice: round.OwnChoice,
		YourReveal: round.YourReveal, OtherReveal: round.OtherReveal,
		Complete: round.Complete,
	}
}

func ampeCommand(
	r *http.Request,
	memberID, otherID, commandID string,
	expected uint64,
) application.Command {
	return application.Command{
		ID: commandID, RoundID: r.PathValue("roundId"),
		RoomID: r.PathValue("circleId"), ActorID: memberID,
		FirstPlayerID: memberID, SecondPlayerID: otherID,
		ExpectedSequence: expected,
	}
}

func createAmpeHandler(rounds AmpeRounds, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct{}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		otherID, err := pairs.Pair(r.Context(), r.PathValue("circleId"), memberID)
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		round, err := rounds.Create(r.Context(), ampeCommand(r, memberID, otherID, commandID, 0))
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectAmpe(round))
	})
}

func viewAmpeHandler(presence AmpePresence, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		otherID, err := pairs.Pair(r.Context(), r.PathValue("circleId"), memberID)
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		round, err := presence.Observe(r.Context(), ampeCommand(r, memberID, otherID, "", 0))
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectAmpe(round))
	})
}

type ampeCommandRequest struct {
	Action           domain.Action `json:"action"`
	Choice           domain.Choice `json:"choice,omitempty"`
	ExpectedSequence uint64        `json:"expectedSequence"`
}

func commandAmpeHandler(rounds AmpeRounds, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body ampeCommandRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		if body.Action != domain.ActionReady && body.Action != domain.ActionLock {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "Action must be ready or lock.",
			})
			return
		}
		otherID, err := pairs.Pair(r.Context(), r.PathValue("circleId"), memberID)
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		round, err := rounds.Apply(
			r.Context(),
			ampeCommand(r, memberID, otherID, commandID, body.ExpectedSequence),
			body.Action, body.Choice,
		)
		if err != nil {
			writeAmpeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectAmpe(round))
	})
}

func writeAmpeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, application.ErrConflict) {
		writeError(w, r, http.StatusConflict, APIError{
			Code: "ampe_conflict", Message: "The round changed. Refresh and try again.",
		})
		return
	}
	writeError(w, r, http.StatusNotFound, APIError{
		Code: "ampe_not_available", Message: "That private Ampe round is not available.",
	})
}
