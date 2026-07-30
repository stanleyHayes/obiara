package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
)

type MemberEscrows interface {
	ForOwner(context.Context, string) ([]domain.Escrow, error)
	FindForOwner(context.Context, string, string) (domain.Escrow, error)
	AddEvidence(context.Context, string, string, string, domain.EvidenceRole, string) (domain.Escrow, error)
	Dispute(context.Context, string, string, string) (domain.Escrow, error)
}

func RegisterEscrowRoutes(mux *http.ServeMux, escrows MemberEscrows, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/escrows", listEscrowsHandler(escrows, keyer, sessions))
	mux.Handle("GET /v1/escrows/{id}", getEscrowHandler(escrows, keyer, sessions))
	mux.Handle("POST /v1/escrows/{id}/milestones/{milestoneId}/acceptance", acceptEscrowMilestoneHandler(escrows, keyer, sessions))
	mux.Handle("POST /v1/escrows/{id}/disputes", disputeEscrowHandler(escrows, keyer, sessions))
}

type escrowMilestoneResponse struct {
	ID                  string `json:"id"`
	GrossPesewas        uint64 `json:"grossPesewas"`
	FeePesewas          uint64 `json:"feePesewas"`
	DeliveryConfirmed   bool   `json:"deliveryConfirmed"`
	AcceptanceConfirmed bool   `json:"acceptanceConfirmed"`
	Settled             bool   `json:"settled"`
	StatementRef        string `json:"statementRef,omitempty"`
}

type escrowResponse struct {
	EscrowID       string                    `json:"escrowId"`
	EngagementID   string                    `json:"engagementId"`
	FundedPesewas  uint64                    `json:"fundedPesewas"`
	SettledPesewas uint64                    `json:"settledPesewas"`
	TermsID        string                    `json:"termsId"`
	TermsVersion   uint64                    `json:"termsVersion"`
	Milestones     []escrowMilestoneResponse `json:"milestones"`
	Disputed       bool                      `json:"disputed"`
	EscalationRef  string                    `json:"escalationRef,omitempty"`
	Revision       uint64                    `json:"revision"`
}

func escrowView(escrow domain.Escrow) escrowResponse {
	state := escrow.State()
	milestones := make([]escrowMilestoneResponse, 0, len(state.Milestones))
	for _, milestone := range state.Milestones {
		item := escrowMilestoneResponse{
			ID: milestone.Term.ID, GrossPesewas: milestone.Term.GrossPesewas,
			FeePesewas: milestone.Term.FeePesewas, Settled: milestone.Settled,
			StatementRef: milestone.StatementRef,
		}
		for _, evidence := range milestone.Evidence {
			item.DeliveryConfirmed = item.DeliveryConfirmed || evidence.Role == domain.DeliveryEvidence
			item.AcceptanceConfirmed = item.AcceptanceConfirmed || evidence.Role == domain.AcceptanceEvidence
		}
		milestones = append(milestones, item)
	}
	response := escrowResponse{
		EscrowID: state.ID, EngagementID: state.EngagementID,
		FundedPesewas: state.FundedPesewas, SettledPesewas: state.SettledPesewas,
		TermsID: state.TermsID, TermsVersion: state.TermsVersion,
		Milestones: milestones, Disputed: state.Dispute != nil, Revision: state.Revision,
	}
	if state.Dispute != nil {
		response.EscalationRef = state.Dispute.EscalationRef
	}
	return response
}

func escrowOwner(w http.ResponseWriter, r *http.Request, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) (string, bool) {
	return matchmakerMember(w, r, keyer, sessions)
}

func listEscrowsHandler(escrows MemberEscrows, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := escrowOwner(w, r, keyer, sessions)
		if !ok {
			return
		}
		found, err := escrows.ForOwner(r.Context(), owner)
		if err != nil {
			writeEscrowError(w, r, err)
			return
		}
		items := make([]escrowResponse, 0, len(found))
		for _, escrow := range found {
			items = append(items, escrowView(escrow))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"items": items})
	})
}

func getEscrowHandler(escrows MemberEscrows, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := escrowOwner(w, r, keyer, sessions)
		if !ok {
			return
		}
		escrow, err := escrows.FindForOwner(r.Context(), r.PathValue("id"), owner)
		if err != nil {
			writeEscrowError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, escrowView(escrow))
	})
}

func acceptEscrowMilestoneHandler(escrows MemberEscrows, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return mutateMemberEscrowHandler(escrows, keyer, sessions, func(ctx context.Context, id, owner, command string, r *http.Request) (domain.Escrow, error) {
		return escrows.AddEvidence(ctx, id, owner, r.PathValue("milestoneId"), domain.AcceptanceEvidence, command)
	})
}

func disputeEscrowHandler(escrows MemberEscrows, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return mutateMemberEscrowHandler(escrows, keyer, sessions, func(ctx context.Context, id, owner, command string, _ *http.Request) (domain.Escrow, error) {
		return escrows.Dispute(ctx, id, owner, command)
	})
}

func mutateMemberEscrowHandler(
	escrows MemberEscrows,
	keyer MatchmakerMemberKeyer,
	sessions SessionAuthenticator,
	mutate func(context.Context, string, string, string, *http.Request) (domain.Escrow, error),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := escrowOwner(w, r, keyer, sessions)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		command := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		escrow, err := mutate(r.Context(), r.PathValue("id"), owner, command, r)
		if err != nil {
			writeEscrowError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, escrowView(escrow))
	})
}

func writeEscrowError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "escrow_not_found", Message: "The protected engagement was not found."})
	case errors.Is(err, application.ErrInvalid):
		writeError(w, r, http.StatusConflict, APIError{Code: "escrow_transition_invalid", Message: "That escrow action is not available in the current state."})
	case errors.Is(err, application.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "escrow_conflict", Message: "The escrow changed. Refresh and try again."})
	default:
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "escrow_unavailable", Message: "The protected engagement is temporarily unavailable."})
	}
}
