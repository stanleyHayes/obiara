package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

// IntroductionSource is the inbound port for asking to be introduced through
// a circle the member already belongs to (E06-S01).
type IntroductionSource interface {
	Open(context.Context, application.Command, application.Proposal) (application.Result, error)
	Withdraw(context.Context, application.Command) (application.Result, error)
	FindForRequester(ctx context.Context, requestID, requesterID string) (domain.Request, error)
}

// sourceTTL is how long an open request stands. The domain refuses anything
// beyond an hour (`expiresAt.After(command.At.Add(time.Hour))`), and it is
// right to: the request carries a snapshot of a circle's roster, and a roster
// is only true for as long as nobody joins or leaves. A day-long request would
// keep offering people who have since walked out.
//
// A first draft of this file set fourteen days and every real request would
// have been refused as invalid.
const sourceTTL = time.Hour

func RegisterSeedSourceRoutes(mux *http.ServeMux, sources IntroductionSource, sessions SessionAuthenticator, gate MemberGate) {
	// Asking to be introduced through a circle is a romantic surface.
	// Withdrawing the ask is not — a member may always take it back.
	mux.Handle("POST /v1/seed/sources", gate.guard(sessions, "introductions.view", "introduction", openSeedSourceHandler(sources, sessions)))
	mux.Handle("GET /v1/seed/sources/{id}", gate.guard(sessions, "introductions.view", "introduction", readSeedSourceHandler(sources, sessions)))
	mux.Handle("DELETE /v1/seed/sources/{id}", withdrawSeedSourceHandler(sources, sessions))
}

type openSeedSourceRequest struct {
	CircleID string `json:"circleId"`
}

// seedSourceResponse deliberately carries a count and not the candidates.
//
// The service keys every candidate before storage so that who reached toward
// whom is not legible at rest, and handing those keys back would undo that —
// a caller could correlate the same person across requests without ever
// learning a name. What a member needs to know is that their ask found
// people; who those people are is the next stage's to reveal, one at a time.
type seedSourceResponse struct {
	RequestID  string `json:"requestId"`
	SourceType string `json:"sourceType"`
	Status     string `json:"status"`
	Candidates int    `json:"candidateCount"`
	ExpiresAt  string `json:"expiresAt"`
	Revision   uint64 `json:"revision"`
}

func toSeedSourceResponse(request domain.Request) seedSourceResponse {
	return seedSourceResponse{
		RequestID:  request.ID(),
		SourceType: string(request.Source().Type),
		Status:     string(request.Status()),
		Candidates: len(request.CandidateIDs()),
		ExpiresAt:  request.ExpiresAt().Format(time.RFC3339),
		Revision:   request.Revision(),
	}
}

func openSeedSourceHandler(sources IntroductionSource, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if sources == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_source_unavailable", Message: "Introductions are unavailable.",
			})
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		var body openSeedSourceRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}

		result, err := sources.Open(r.Context(),
			application.Command{ID: commandID, ActorID: memberID, ReasonCode: "member_request"},
			application.Proposal{
				RequesterID: memberID,
				SourceRef:   strings.TrimSpace(body.CircleID),
				SourceType:  domain.SourceCircle,
				TTL:         sourceTTL,
			})
		if err != nil {
			writeSeedSourceError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, toSeedSourceResponse(result.Request))
	})
}

func readSeedSourceHandler(sources IntroductionSource, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if sources == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_source_unavailable", Message: "Introductions are unavailable.",
			})
			return
		}
		request, err := sources.FindForRequester(r.Context(), r.PathValue("id"), memberID)
		if err != nil {
			writeSeedSourceError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSeedSourceResponse(request))
	})
}

func withdrawSeedSourceHandler(sources IntroductionSource, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if sources == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_source_unavailable", Message: "Introductions are unavailable.",
			})
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		result, err := sources.Withdraw(r.Context(), application.Command{
			ID: commandID, RequestID: r.PathValue("id"),
			ActorID: memberID, ReasonCode: "member_withdrew",
		})
		if err != nil {
			writeSeedSourceError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSeedSourceResponse(result.Request))
	})
}

func writeSeedSourceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "introduction_source_not_found", Message: "That introduction request was not found.",
		})
	case errors.Is(err, application.ErrNotAvailable):
		// Covers "not a member of that circle" and "not your request". Both
		// answer the same way: a distinct refusal would confirm the circle or
		// the request is real.
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "introduction_source_not_found", Message: "That introduction request was not found.",
		})
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrInvalidRequest):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "That introduction request cannot change in the way asked.",
		})
	case errors.Is(err, domain.ErrStaleRevision), errors.Is(err, domain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "introduction_source_conflict", Message: "That request changed while this one was in flight.",
		})
	case errors.Is(err, application.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "introduction_source_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "introduction_source_unavailable", Message: "Introductions are temporarily unavailable.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
