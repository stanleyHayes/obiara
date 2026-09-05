package apihttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	competition "github.com/stanleyHayes/obiara/services/api/internal/games/competition"
	competitionapp "github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	cohortapp "github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/application"
	competitiondomain "github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
)

type CompetitionCohorts interface {
	Create(context.Context, cohortapp.Command, int) (cohortapp.Projection, error)
	Join(context.Context, cohortapp.Command) (cohortapp.Projection, error)
	Leave(context.Context, cohortapp.Command) (cohortapp.Projection, error)
	View(context.Context, cohortapp.Command) (cohortapp.Projection, error)
	ViewForManager(context.Context, cohortapp.Command) (cohortapp.Projection, error)
}

type CompetitionManager interface {
	Start(context.Context, competition.StartCommand) (competition.StartResult, error)
}

type PrivateCompetitions interface {
	ViewPrivate(context.Context, competitionapp.Command) (competitionapp.PrivateProjection, error)
	OpenReviewPrivate(context.Context, competitionapp.Command, string, string) (competitionapp.PrivateProjection, error)
	AppealPrivate(context.Context, competitionapp.Command, string) (competitionapp.PrivateProjection, error)
}

type CompetitionReviewDesk interface {
	ViewForReviewer(context.Context, competitionapp.Command) (competitionapp.PrivateProjection, error)
	ResolveReviewPrivate(context.Context, competitionapp.Command, string, competitiondomain.Decision) (competitionapp.PrivateProjection, error)
	ResolveAppealPrivate(context.Context, competitionapp.Command, string, competitiondomain.Decision) (competitionapp.PrivateProjection, error)
}

type TournamentOwareGames interface {
	Launch(context.Context, competition.TournamentOwareCommand, time.Duration) (competition.TournamentOwareProjection, error)
	View(context.Context, competition.TournamentOwareCommand) (competition.TournamentOwareProjection, error)
	Move(context.Context, competition.TournamentOwareCommand, int) (competition.TournamentOwareProjection, error)
	Finalize(context.Context, competition.TournamentOwareCommand) (competitionapp.PrivateProjection, error)
}

func RegisterCompetitionRoutes(
	mux *http.ServeMux,
	cohorts CompetitionCohorts,
	manager CompetitionManager,
	competitions PrivateCompetitions,
	reviews CompetitionReviewDesk,
	oware TournamentOwareGames,
	memberAuth SessionAuthenticator,
	adminAuth AdminPrincipalResolver,
	gate MemberGate,
) {
	mux.Handle("POST /v1/admin/game-cohorts", createCompetitionCohortHandler(cohorts, adminAuth))
	mux.Handle("GET /v1/admin/game-cohorts/{cohortId}", adminCompetitionCohortHandler(cohorts, adminAuth))
	mux.Handle("POST /v1/admin/game-cohorts/{cohortId}/start", startCompetitionHandler(manager, adminAuth))
	mux.Handle("GET /v1/game-cohorts/{cohortId}", viewCompetitionCohortHandler(cohorts, memberAuth))
	// Joining a cohort is taking part. Leaving one is not.
	mux.Handle("POST /v1/game-cohorts/{cohortId}/join", gate.guard(memberAuth, "games.play", "game", joinCompetitionCohortHandler(cohorts, memberAuth)))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/leave", leaveCompetitionCohortHandler(cohorts, memberAuth))
	mux.Handle("GET /v1/game-cohorts/{cohortId}/competitions/{competitionId}", viewCompetitionHandler(competitions, memberAuth))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware", launchTournamentOwareHandler(oware, memberAuth))
	mux.Handle("GET /v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}", viewTournamentOwareHandler(oware, memberAuth))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/moves", moveTournamentOwareHandler(oware, memberAuth))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/oware/{gameId}/finalize", finalizeTournamentOwareHandler(oware, memberAuth))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/competitions/{competitionId}/matches/{matchId}/reviews", openCompetitionReviewHandler(competitions, memberAuth))
	mux.Handle("POST /v1/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/appeal", appealCompetitionReviewHandler(competitions, memberAuth))
	mux.Handle("GET /v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}", viewCompetitionReviewDeskHandler(reviews, adminAuth))
	mux.Handle("POST /v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve", resolveCompetitionReviewHandler(reviews, adminAuth, false))
	mux.Handle("POST /v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve-appeal", resolveCompetitionReviewHandler(reviews, adminAuth, true))
}

type competitionCohortResponse struct {
	ID            string `json:"id"`
	Capacity      int    `json:"capacity"`
	Enrolled      int    `json:"enrolled"`
	Joined        bool   `json:"joined"`
	Status        string `json:"status"`
	CompetitionID string `json:"competitionId,omitempty"`
	Revision      uint64 `json:"revision"`
}

func projectCompetitionCohort(cohort cohortapp.Projection) competitionCohortResponse {
	return competitionCohortResponse{
		ID: cohort.ID, Capacity: cohort.Capacity, Enrolled: cohort.Enrolled,
		Joined: cohort.Joined, Status: string(cohort.Status),
		CompetitionID: cohort.CompetitionID, Revision: cohort.Revision,
	}
}

func createCompetitionCohortHandler(cohorts CompetitionCohorts, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			Capacity int `json:"capacity"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		cohort, err := cohorts.Create(r.Context(), cohortapp.Command{ID: commandID, ActorID: principal.ActorID}, body.Capacity)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectCompetitionCohort(cohort))
	})
}

func adminCompetitionCohortHandler(cohorts CompetitionCohorts, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		cohort, err := cohorts.ViewForManager(r.Context(), cohortapp.Command{CohortID: r.PathValue("cohortId"), ActorID: principal.ActorID})
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCompetitionCohort(cohort))
	})
}

func startCompetitionHandler(manager CompetitionManager, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			ExpectedRevision uint64 `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		result, err := manager.Start(r.Context(), competition.StartCommand{
			ID: commandID, CohortID: r.PathValue("cohortId"),
			ActorID: principal.ActorID, ExpectedRevision: body.ExpectedRevision,
		})
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, struct {
			Cohort      competitionCohortResponse  `json:"cohort"`
			Competition privateCompetitionResponse `json:"competition"`
		}{projectCompetitionCohort(result.Cohort), projectPrivateCompetition(result.Competition)})
	})
}

func memberCohortCommand(w http.ResponseWriter, r *http.Request, auth SessionAuthenticator, requireCommand bool) (cohortapp.Command, bool) {
	memberID, ok := subanSubject(w, r, auth)
	if !ok {
		return cohortapp.Command{}, false
	}
	command := cohortapp.Command{CohortID: r.PathValue("cohortId"), ActorID: memberID}
	if requireCommand {
		command.ID, ok = circleCommandID(w, r)
		if !ok {
			return cohortapp.Command{}, false
		}
		var body struct {
			ExpectedRevision uint64 `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return cohortapp.Command{}, false
		}
		command.ExpectedRevision = body.ExpectedRevision
	}
	return command, true
}

func viewCompetitionCohortHandler(cohorts CompetitionCohorts, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command, ok := memberCohortCommand(w, r, auth, false)
		if !ok {
			return
		}
		cohort, err := cohorts.View(r.Context(), command)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCompetitionCohort(cohort))
	})
}

func joinCompetitionCohortHandler(cohorts CompetitionCohorts, auth SessionAuthenticator) http.Handler {
	return competitionCohortMutationHandler(cohorts.Join, auth)
}
func leaveCompetitionCohortHandler(cohorts CompetitionCohorts, auth SessionAuthenticator) http.Handler {
	return competitionCohortMutationHandler(cohorts.Leave, auth)
}
func competitionCohortMutationHandler(mutate func(context.Context, cohortapp.Command) (cohortapp.Projection, error), auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command, ok := memberCohortCommand(w, r, auth, true)
		if !ok {
			return
		}
		cohort, err := mutate(r.Context(), command)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCompetitionCohort(cohort))
	})
}

type privateCompetitionResponse struct {
	ID       string                              `json:"id"`
	Status   string                              `json:"status"`
	Revision uint64                              `json:"revision"`
	Matches  []competitionapp.PrivateMatch       `json:"matches"`
	Ladder   []competitionapp.PrivateLadderEntry `json:"ladder"`
	Reviews  []competitionapp.PrivateReview      `json:"reviews"`
}

func projectPrivateCompetition(value competitionapp.PrivateProjection) privateCompetitionResponse {
	return privateCompetitionResponse{
		ID: value.ID, Status: string(value.Status), Revision: value.Revision,
		Matches: value.Matches, Ladder: value.Ladder, Reviews: value.Reviews,
	}
}

func viewCompetitionHandler(competitions PrivateCompetitions, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		value, err := competitions.ViewPrivate(r.Context(), competitionapp.Command{
			CompetitionID: r.PathValue("competitionId"),
			CohortID:      r.PathValue("cohortId"), ActorID: memberID,
		})
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPrivateCompetition(value))
	})
}

type tournamentOwareResponse struct {
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

func projectTournamentOware(game competition.TournamentOwareProjection) tournamentOwareResponse {
	turn, player := "south", "south"
	if game.Turn == 1 {
		turn = "north"
	}
	if game.YourPlayer == 1 {
		player = "north"
	}
	return tournamentOwareResponse{
		ID: game.ID, Houses: game.Houses, Captured: game.Captured,
		Turn: turn, YourPlayer: player, YourTurn: game.Turn == game.YourPlayer,
		Status: string(game.Status), Winner: game.Winner, Revision: game.Revision,
		MoveDeadline: game.MoveDeadline, ServerTime: game.ServerTime,
	}
}

func tournamentOwareCommand(r *http.Request, actorID, commandID string) competition.TournamentOwareCommand {
	return competition.TournamentOwareCommand{
		ID: commandID, CohortID: r.PathValue("cohortId"),
		CompetitionID: r.PathValue("competitionId"), MatchID: r.PathValue("matchId"),
		SessionID: r.PathValue("gameId"), ActorID: actorID,
	}
}

func launchTournamentOwareHandler(games TournamentOwareGames, auth SessionAuthenticator) http.Handler {
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
		game, err := games.Launch(r.Context(), tournamentOwareCommand(r, memberID, commandID), 24*time.Hour)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectTournamentOware(game))
	})
}

func viewTournamentOwareHandler(games TournamentOwareGames, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		game, err := games.View(r.Context(), tournamentOwareCommand(r, memberID, ""))
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectTournamentOware(game))
	})
}

func moveTournamentOwareHandler(games TournamentOwareGames, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			Pit              int    `json:"pit"`
			ExpectedRevision uint64 `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command := tournamentOwareCommand(r, memberID, commandID)
		command.ExpectedRevision = body.ExpectedRevision
		game, err := games.Move(r.Context(), command, body.Pit)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectTournamentOware(game))
	})
}

func finalizeTournamentOwareHandler(games TournamentOwareGames, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			ExpectedCompetitionRevision uint64 `json:"expectedCompetitionRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command := tournamentOwareCommand(r, memberID, commandID)
		command.ExpectedCompetitionRevision = body.ExpectedCompetitionRevision
		value, err := games.Finalize(r.Context(), command)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPrivateCompetition(value))
	})
}

func openCompetitionReviewHandler(competitions PrivateCompetitions, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			EvidenceRef      string `json:"evidenceRef"`
			ExpectedRevision uint64 `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		value, err := competitions.OpenReviewPrivate(r.Context(), competitionapp.Command{
			ID: commandID, CompetitionID: r.PathValue("competitionId"),
			CohortID: r.PathValue("cohortId"), ActorID: memberID,
			ExpectedRevision: body.ExpectedRevision,
		}, r.PathValue("matchId"), body.EvidenceRef)
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectPrivateCompetition(value))
	})
}

func appealCompetitionReviewHandler(competitions PrivateCompetitions, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			ExpectedRevision uint64 `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		value, err := competitions.AppealPrivate(r.Context(), competitionapp.Command{
			ID: commandID, CompetitionID: r.PathValue("competitionId"),
			CohortID: r.PathValue("cohortId"), ActorID: memberID,
			ExpectedRevision: body.ExpectedRevision,
		}, r.PathValue("reviewId"))
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPrivateCompetition(value))
	})
}

func viewCompetitionReviewDeskHandler(reviews CompetitionReviewDesk, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		value, err := reviews.ViewForReviewer(r.Context(), competitionapp.Command{
			CompetitionID: r.PathValue("competitionId"),
			CohortID:      r.PathValue("cohortId"), ActorID: principal.ActorID,
		})
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPrivateCompetition(value))
	})
}

func resolveCompetitionReviewHandler(reviews CompetitionReviewDesk, resolve AdminPrincipalResolver, appeal bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !principal.MFAVerified {
			writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before resolving competition reviews."})
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			Decision         competitiondomain.Decision `json:"decision"`
			ExpectedRevision uint64                     `json:"expectedRevision"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command := competitionapp.Command{
			ID: commandID, CompetitionID: r.PathValue("competitionId"),
			CohortID: r.PathValue("cohortId"), ActorID: principal.ActorID,
			ExpectedRevision: body.ExpectedRevision,
		}
		var (
			value competitionapp.PrivateProjection
			err   error
		)
		if appeal {
			value, err = reviews.ResolveAppealPrivate(r.Context(), command, r.PathValue("reviewId"), body.Decision)
		} else {
			value, err = reviews.ResolveReviewPrivate(r.Context(), command, r.PathValue("reviewId"), body.Decision)
		}
		if err != nil {
			writeCompetitionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPrivateCompetition(value))
	})
}

func writeCompetitionError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, cohortapp.ErrConflict) || errors.Is(err, competitionapp.ErrConflict) {
		writeError(w, r, http.StatusConflict, APIError{Code: "competition_conflict", Message: "The cohort or bracket changed. Refresh and try again."})
		return
	}
	writeError(w, r, http.StatusNotFound, APIError{Code: "competition_not_available", Message: "That private competition reference is not available."})
}
