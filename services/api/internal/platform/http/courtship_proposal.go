package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
)

// Proposals is the inbound port for courtship proposals: one member offering
// a call, a meeting or exclusivity, and the other deciding.
type Proposals interface {
	Create(context.Context, application.CreateCommand) (application.Result, error)
	List(ctx context.Context, memberID string, limit int) ([]application.Summary, error)
	Accept(context.Context, application.DecisionCommand) (application.Result, error)
	Reject(context.Context, application.DecisionCommand) (application.Result, error)
	Withdraw(context.Context, application.DecisionCommand) (application.Result, error)
}

// RegisterCourtshipProposalRoutes adds the proposal routes.
//
// Every write carries a command id and, for decisions, the revision the
// caller believes it is acting on. A proposal is a two-party negotiation over
// an unreliable mobile network: without those, a retried tap could create a
// second proposal, and two devices could race a decision.
func RegisterCourtshipProposalRoutes(mux *http.ServeMux, proposals Proposals, sessions SessionAuthenticator, gate MemberGate) {
	mux.Handle("GET /v1/courtship/proposals", listProposalsHandler(proposals, sessions))
	// Making a proposal and accepting one advance a courtship. Rejecting and
	// withdrawing end one, so they stay open at every rung.
	mux.Handle("POST /v1/courtship/proposals", gate.guard(sessions, "rooms.participate", "room", createProposalHandler(proposals, sessions)))
	mux.Handle("POST /v1/courtship/proposals/{id}/accept", gate.guard(sessions, "rooms.participate", "room", decideProposalHandler(proposals, sessions, "accept")))
	mux.Handle("POST /v1/courtship/proposals/{id}/reject", decideProposalHandler(proposals, sessions, "reject"))
	mux.Handle("POST /v1/courtship/proposals/{id}/withdraw", decideProposalHandler(proposals, sessions, "withdraw"))
}

type createProposalRequest struct {
	CommandID   string `json:"commandId"`
	RecipientID string `json:"recipientId"`
	Kind        string `json:"kind"`
	Detail      string `json:"detail"`
	ExpiresAt   string `json:"expiresAt"`
}

type decideProposalRequest struct {
	CommandID        string `json:"commandId"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

type proposalResponse struct {
	ProposalID string `json:"proposalId"`
	Status     string `json:"status"`
	Revision   uint64 `json:"revision"`
	// Replayed reports that this command had already been applied, so the
	// caller can distinguish a successful retry from a fresh action.
	Replayed bool `json:"replayed"`
}

func toProposalResponse(result application.Result) proposalResponse {
	return proposalResponse{
		ProposalID: result.Proposal.ID(),
		Status:     string(result.Proposal.Status()),
		Revision:   result.Proposal.Revision(),
		Replayed:   result.Replayed,
	}
}

type proposalSummaryResponse struct {
	ProposalID string `json:"proposalId"`
	Kind       string `json:"kind"`
	Status     string `json:"status"`
	Revision   uint64 `json:"revision"`
	ExpiresAt  string `json:"expiresAt"`
	Outgoing   bool   `json:"outgoing"`
}

type proposalListResponse struct {
	Proposals []proposalSummaryResponse `json:"proposals"`
}

// listProposalsHandler returns the caller's own proposals. The member comes
// from the session, so there is no path by which one member can read
// another's list.
func listProposalsHandler(proposals Proposals, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		limit := 50
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 100 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "limit", Reason: "must be between 1 and 100"}},
				})
				return
			}
			limit = parsed
		}

		summaries, err := proposals.List(r.Context(), memberID, limit)
		if err != nil {
			writeProposalError(w, r, err)
			return
		}
		response := proposalListResponse{Proposals: make([]proposalSummaryResponse, 0, len(summaries))}
		for _, summary := range summaries {
			response.Proposals = append(response.Proposals, proposalSummaryResponse{
				ProposalID: summary.ID, Kind: string(summary.Kind), Status: string(summary.Status),
				Revision: summary.Revision, ExpiresAt: summary.ExpiresAt.UTC().Format(time.RFC3339),
				Outgoing: summary.Outgoing,
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func createProposalHandler(proposals Proposals, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		// The sender is the session, never the body: a member must not be
		// able to send a proposal as somebody else.
		senderID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}

		var body createProposalRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}

		var details []FieldError
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.RecipientID = strings.TrimSpace(body.RecipientID)
		body.Detail = strings.TrimSpace(body.Detail)
		if !validOpaqueID(body.CommandID) {
			details = append(details, FieldError{Field: "commandId", Reason: "must be an opaque identifier"})
		}
		if !validOpaqueID(body.RecipientID) {
			details = append(details, FieldError{Field: "recipientId", Reason: "must be an opaque identifier"})
		}
		if body.RecipientID == senderID {
			details = append(details, FieldError{Field: "recipientId", Reason: "cannot be yourself"})
		}
		kind := domain.Type(strings.ToLower(strings.TrimSpace(body.Kind)))
		switch kind {
		case domain.TypeCall, domain.TypeMeeting, domain.TypeExclusivity:
		default:
			details = append(details, FieldError{Field: "kind", Reason: "must be call, meeting or exclusivity"})
		}
		if body.Detail == "" || len(body.Detail) > 500 {
			details = append(details, FieldError{Field: "detail", Reason: "must be 1-500 characters"})
		}
		expiresAt, err := time.Parse(time.RFC3339, strings.TrimSpace(body.ExpiresAt))
		if err != nil {
			details = append(details, FieldError{Field: "expiresAt", Reason: "must be an RFC3339 timestamp"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		result, err := proposals.Create(r.Context(), application.CreateCommand{
			ID: body.CommandID, SenderID: senderID, RecipientID: body.RecipientID,
			DetailRef: body.Detail, Kind: kind, ExpiresAt: expiresAt.UTC(),
		})
		if err != nil {
			writeProposalError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, toProposalResponse(result))
	})
}

func decideProposalHandler(proposals Proposals, sessions SessionAuthenticator, action string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		proposalID := r.PathValue("id")
		if !validOpaqueID(proposalID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}

		var body decideProposalRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		if !validOpaqueID(body.CommandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "commandId", Reason: "must be an opaque identifier"}},
			})
			return
		}

		command := application.DecisionCommand{
			ID: body.CommandID, ProposalID: proposalID, ActorID: actorID,
			ExpectedRevision: body.ExpectedRevision,
		}
		var result application.Result
		var err error
		switch action {
		case "accept":
			result, err = proposals.Accept(r.Context(), command)
		case "reject":
			result, err = proposals.Reject(r.Context(), command)
		default:
			result, err = proposals.Withdraw(r.Context(), command)
		}
		if err != nil {
			writeProposalError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toProposalResponse(result))
	})
}

func proposalJSONGuard(w http.ResponseWriter, r *http.Request) bool {
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{
			Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
		})
		return false
	}
	return true
}

// writeProposalError maps the slice's failures.
//
// A proposal that does not exist and one the caller is not part of answer
// identically. Separating them would let anyone enumerate other members'
// proposals by identifier, which is exactly the private negotiation this
// context is built to protect.
func writeProposalError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotAvailable), errors.Is(err, domain.ErrNotParticipant):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "proposal_not_found", Message: "That proposal is not available.",
		})
	case errors.Is(err, domain.ErrTerminal):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "proposal_decided", Message: "That proposal has already been decided.",
		})
	case errors.Is(err, domain.ErrExpired):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "proposal_expired", Message: "That proposal has expired.",
		})
	case errors.Is(err, domain.ErrStaleRevision), errors.Is(err, application.ErrConcurrentChange):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "proposal_conflict", Message: "That proposal changed. Refresh and try again.",
		})
	case errors.Is(err, domain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, application.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "feature_unavailable", Message: "Courtship proposals are not available right now.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
