package apihttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/domain"
)

type EbeCatalog interface {
	Approve(context.Context, string, application.PromptApproval) (application.PromptProjection, error)
}

type EbeDuels interface {
	Create(context.Context, application.Command) (application.Projection, error)
	Answer(context.Context, application.Command, string) (application.Projection, error)
	View(context.Context, application.Command) (application.Projection, error)
}

func RegisterEbeRoutes(
	mux *http.ServeMux,
	catalog EbeCatalog,
	duels EbeDuels,
	pairs OwarePairResolver,
	memberAuth SessionAuthenticator,
	adminAuth AdminPrincipalResolver,
) {
	mux.Handle("POST /v1/admin/games/ebe/prompts", approveEbePromptHandler(catalog, adminAuth))
	mux.Handle("POST /v1/circles/{circleId}/ebe", createEbeHandler(duels, pairs, memberAuth))
	mux.Handle("GET /v1/circles/{circleId}/ebe/{duelId}", viewEbeHandler(duels, pairs, memberAuth))
	mux.Handle("POST /v1/circles/{circleId}/ebe/{duelId}/answers", answerEbeHandler(duels, pairs, memberAuth))
}

type ebePromptResponse struct {
	ID             string `json:"id"`
	Version        uint64 `json:"version"`
	Language       string `json:"language"`
	Cue            string `json:"cue"`
	SourceKind     string `json:"sourceKind"`
	SourceCitation string `json:"sourceCitation"`
	SourceLocator  string `json:"sourceLocator,omitempty"`
}

func projectEbePrompt(prompt application.PromptProjection) ebePromptResponse {
	return ebePromptResponse{
		ID: prompt.ID, Version: prompt.Version, Language: prompt.Language,
		Cue: prompt.Cue, SourceKind: string(prompt.SourceKind),
		SourceCitation: prompt.SourceCitation, SourceLocator: prompt.SourceLocator,
	}
}

type ebeTurnResponse struct {
	Number            uint64            `json:"number"`
	Prompt            ebePromptResponse `json:"prompt"`
	Yours             bool              `json:"yours"`
	YourAnswer        string            `json:"yourAnswer,omitempty"`
	YourAnswerCorrect *bool             `json:"yourAnswerCorrect,omitempty"`
}

type ebeResponse struct {
	ID            string             `json:"id"`
	Revision      uint64             `json:"revision"`
	Complete      bool               `json:"complete"`
	YourTurn      bool               `json:"yourTurn"`
	CurrentPrompt *ebePromptResponse `json:"currentPrompt,omitempty"`
	Turns         []ebeTurnResponse  `json:"turns"`
}

func projectEbe(duel application.Projection) ebeResponse {
	result := ebeResponse{
		ID: duel.ID, Revision: duel.Revision, Complete: duel.Complete,
		YourTurn: duel.YourTurn, Turns: make([]ebeTurnResponse, 0, len(duel.Turns)),
	}
	if duel.CurrentPrompt != nil {
		prompt := projectEbePrompt(*duel.CurrentPrompt)
		result.CurrentPrompt = &prompt
	}
	for _, turn := range duel.Turns {
		result.Turns = append(result.Turns, ebeTurnResponse{
			Number: turn.Number, Prompt: projectEbePrompt(turn.Prompt),
			Yours: turn.Yours, YourAnswer: turn.YourAnswer,
			YourAnswerCorrect: turn.YourAnswerCorrect,
		})
	}
	return result
}

type ebePromptApprovalRequest struct {
	ID              string        `json:"id"`
	Version         uint64        `json:"version"`
	Language        string        `json:"language"`
	Cue             string        `json:"cue"`
	AcceptedAnswers []string      `json:"acceptedAnswers"`
	Source          domain.Source `json:"source"`
}

func approveEbePromptHandler(catalog EbeCatalog, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !packJSONGuard(w, r) {
			return
		}
		var body ebePromptApprovalRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		prompt, err := catalog.Approve(r.Context(), principal.ActorID, application.PromptApproval{
			ID: body.ID, Version: body.Version, Language: body.Language,
			Cue: body.Cue, AcceptedAnswers: body.AcceptedAnswers, Source: body.Source,
		})
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectEbePrompt(prompt))
	})
}

func ebeCommand(r *http.Request, memberID, otherID, commandID string, revision uint64) application.Command {
	return application.Command{
		ID: commandID, DuelID: r.PathValue("duelId"), RoomID: r.PathValue("circleId"),
		ActorID: memberID, FirstPlayerID: memberID, SecondPlayerID: otherID,
		ExpectedRevision: revision,
	}
}

func createEbeHandler(duels EbeDuels, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
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
			writeEbeError(w, r, err)
			return
		}
		duel, err := duels.Create(r.Context(), ebeCommand(r, memberID, otherID, commandID, 0))
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectEbe(duel))
	})
}

func viewEbeHandler(duels EbeDuels, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		otherID, err := pairs.Pair(r.Context(), r.PathValue("circleId"), memberID)
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		duel, err := duels.View(r.Context(), ebeCommand(r, memberID, otherID, "", 0))
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectEbe(duel))
	})
}

type ebeAnswerRequest struct {
	Answer           string `json:"answer"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

func answerEbeHandler(duels EbeDuels, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body ebeAnswerRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		otherID, err := pairs.Pair(r.Context(), r.PathValue("circleId"), memberID)
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		duel, err := duels.Answer(r.Context(), ebeCommand(r, memberID, otherID, commandID, body.ExpectedRevision), body.Answer)
		if err != nil {
			writeEbeError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectEbe(duel))
	})
}

func writeEbeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, application.ErrInvalidPrompt) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "The reviewed prompt or its source is invalid."})
		return
	}
	if errors.Is(err, application.ErrConflict) {
		writeError(w, r, http.StatusConflict, APIError{Code: "ebe_conflict", Message: "The duel changed. Refresh and try again."})
		return
	}
	writeError(w, r, http.StatusNotFound, APIError{Code: "ebe_not_available", Message: "That reviewed private duel is not available."})
}
