package apihttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	oware "github.com/stanleyHayes/obiara/services/api/internal/games/oware/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	sessiondomain "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
)

type OwareSessions interface {
	Create(context.Context, application.Command, string, time.Duration) (sessiondomain.Projection, error)
	Move(context.Context, application.Command, int) (sessiondomain.Projection, error)
	View(context.Context, application.Command) (sessiondomain.Projection, error)
}

type OwarePairResolver interface {
	Pair(context.Context, string, string) (string, error)
}

func RegisterOwareRoutes(
	mux *http.ServeMux,
	sessions OwareSessions,
	pairs OwarePairResolver,
	auth SessionAuthenticator,
	gate MemberGate,
) {
	mux.Handle("POST /v1/circles/{circleId}/oware", gate.guard(auth, "games.play", "game", createOwareHandler(sessions, pairs, auth)))
	mux.Handle("GET /v1/circles/{circleId}/oware/{gameId}", viewOwareHandler(sessions, auth))
	mux.Handle("POST /v1/circles/{circleId}/oware/{gameId}/moves", gate.guard(auth, "games.play", "game", moveOwareHandler(sessions, auth)))
}

type owareResponse struct {
	ID           string    `json:"id"`
	Houses       [12]int   `json:"houses"`
	Captured     [2]int    `json:"captured"`
	Turn         string    `json:"turn"`
	YourPlayer   string    `json:"yourPlayer"`
	YourTurn     bool      `json:"yourTurn"`
	Status       string    `json:"status"`
	Winner       int       `json:"winner"`
	Revision     uint64    `json:"revision"`
	MoveDeadline time.Time `json:"moveDeadline"`
	ServerTime   time.Time `json:"serverTime"`
}

func projectOware(game sessiondomain.Projection) owareResponse {
	turn := "south"
	if game.Turn == oware.North {
		turn = "north"
	}
	player := "south"
	if game.YourPlayer == oware.North {
		player = "north"
	}
	return owareResponse{
		ID: game.ID, Houses: game.Houses, Captured: game.Captured,
		Turn: turn, YourPlayer: player, YourTurn: game.Turn == game.YourPlayer,
		Status: string(game.Status), Winner: game.Winner, Revision: game.Revision,
		MoveDeadline: game.MoveDeadline, ServerTime: game.ServerTime,
	}
}

func createOwareHandler(sessions OwareSessions, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
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
		circleID := r.PathValue("circleId")
		second, err := pairs.Pair(r.Context(), circleID, memberID)
		if err != nil {
			writeOwareError(w, r, err)
			return
		}
		game, err := sessions.Create(r.Context(), application.Command{
			ID: commandID, RoomID: circleID, ActorID: memberID,
		}, second, sessiondomain.MaxMoveWindow)
		if err != nil {
			writeOwareError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectOware(game))
	})
}

func viewOwareHandler(sessions OwareSessions, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		game, err := sessions.View(r.Context(), application.Command{
			SessionID: r.PathValue("gameId"), RoomID: r.PathValue("circleId"),
			ActorID: memberID,
		})
		if err != nil {
			writeOwareError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectOware(game))
	})
}

type owareMoveRequest struct {
	Pit              int    `json:"pit"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

func moveOwareHandler(sessions OwareSessions, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body owareMoveRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		if body.Pit < 0 || body.Pit > 11 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "Pit must be between 0 and 11.",
			})
			return
		}
		game, err := sessions.Move(r.Context(), application.Command{
			ID: commandID, SessionID: r.PathValue("gameId"),
			RoomID: r.PathValue("circleId"), ActorID: memberID,
			ExpectedRevision: body.ExpectedRevision,
		}, body.Pit)
		if err != nil {
			writeOwareError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectOware(game))
	})
}

func writeOwareError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrConflict), errors.Is(err, application.ErrApplied):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "oware_conflict", Message: "The board changed. Refresh and try again.",
		})
	default:
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "oware_not_available",
			Message: "That private game is not available in this room.",
		})
	}
}
