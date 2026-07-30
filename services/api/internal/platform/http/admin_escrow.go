package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	escrowapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
	escrowdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	matchmakerapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	matchmakerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type OperationsEngagements interface {
	FindForOperations(context.Context, string) (matchmakerdomain.Engagement, error)
}

type OperationsEscrows interface {
	FundAudited(context.Context, string, string, string, uint64, escrowdomain.Terms, string, string) (escrowdomain.Escrow, error)
	AddEvidenceAudited(context.Context, string, string, escrowdomain.EvidenceRole, string, string) (escrowdomain.Escrow, error)
	SettleAudited(context.Context, string, string, string, string) (escrowdomain.Escrow, escrowdomain.Statement, error)
}

func RegisterAdminEscrowRoutes(mux *http.ServeMux, escrows OperationsEscrows, engagements OperationsEngagements, resolve AdminPrincipalResolver) {
	mux.Handle("POST /v1/admin/escrows", adminFundEscrowHandler(escrows, engagements, resolve))
	mux.Handle("POST /v1/admin/escrows/{id}/milestones/{milestoneId}/delivery", adminEscrowDeliveryHandler(escrows, resolve))
	mux.Handle("POST /v1/admin/escrows/{id}/milestones/{milestoneId}/settlement", adminEscrowSettlementHandler(escrows, resolve))
}

type adminFundEscrowInput struct {
	EngagementID string `json:"engagementId"`
	FundingRef   string `json:"fundingRef"`
}

func adminFundEscrowHandler(escrows OperationsEscrows, engagements OperationsEngagements, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		var body adminFundEscrowInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		command := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		engagement, err := engagements.FindForOperations(r.Context(), strings.TrimSpace(body.EngagementID))
		if err != nil {
			writeError(w, r, http.StatusNotFound, APIError{Code: "engagement_not_found", Message: "The booked engagement was not found."})
			return
		}
		state := engagement.State()
		terms := escrowTermsFromEngagement(state.Terms)
		escrow, err := escrows.FundAudited(r.Context(), state.MemberKey, state.ID, strings.TrimSpace(body.FundingRef), state.Terms.TotalFeePesewas, terms, command, actorID)
		if err != nil {
			writeAdminEscrowError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, escrowView(escrow))
	})
}

func escrowTermsFromEngagement(terms matchmakerdomain.Terms) escrowdomain.Terms {
	milestones := make([]escrowdomain.MilestoneTerm, len(terms.Milestones))
	for i, milestone := range terms.Milestones {
		milestones[i] = escrowdomain.MilestoneTerm{ID: milestone.ID, GrossPesewas: milestone.FeePesewas}
	}
	return escrowdomain.Terms{ID: terms.ID, Version: terms.Version, Milestones: milestones}
}

func adminEscrowDeliveryHandler(escrows OperationsEscrows, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		escrow, err := escrows.AddEvidenceAudited(r.Context(), r.PathValue("id"), r.PathValue("milestoneId"), escrowdomain.DeliveryEvidence, strings.TrimSpace(r.Header.Get("Idempotency-Key")), actorID)
		if err != nil {
			writeAdminEscrowError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, escrowView(escrow))
	})
}

type escrowSettlementResponse struct {
	Escrow       escrowResponse `json:"escrow"`
	StatementRef string         `json:"statementRef"`
	GrossPesewas uint64         `json:"grossPesewas"`
	FeePesewas   uint64         `json:"feePesewas"`
	NetPesewas   uint64         `json:"netPesewas"`
	SettledAt    string         `json:"settledAt"`
}

func requireFinanceAdmin(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver) (string, bool) {
	principal, ok := resolveAdminPrincipal(w, r, resolve)
	if !ok {
		return "", false
	}
	if !principal.Has(admin.ScopeFinance) {
		writeAdminVerificationError(w, r, errAdminOperationsForbidden)
		return "", false
	}
	if !principal.MFAVerified {
		writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before settling escrow."})
		return "", false
	}
	return principal.ActorID, true
}

func adminEscrowSettlementHandler(escrows OperationsEscrows, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireFinanceAdmin(w, r, resolve)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		escrow, statement, err := escrows.SettleAudited(
			r.Context(), r.PathValue("id"), r.PathValue("milestoneId"),
			strings.TrimSpace(r.Header.Get("Idempotency-Key")), actorID,
		)
		if err != nil {
			writeAdminEscrowError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, escrowSettlementResponse{
			Escrow: escrowView(escrow), StatementRef: statement.Ref,
			GrossPesewas: statement.GrossPesewas, FeePesewas: statement.FeePesewas,
			NetPesewas: statement.NetPesewas, SettledAt: statement.SettledAt.UTC().Format(time.RFC3339),
		})
	})
}

func writeAdminEscrowError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, escrowapp.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "escrow_transition_invalid", Message: "The escrow action is invalid for the current state."})
	case errors.Is(err, escrowapp.ErrNotFound), errors.Is(err, matchmakerapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "escrow_not_found", Message: "The protected engagement was not found."})
	case errors.Is(err, escrowapp.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "escrow_conflict", Message: "The escrow changed. Refresh and try again."})
	default:
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "escrow_unavailable", Message: "The protected engagement is temporarily unavailable."})
	}
}
