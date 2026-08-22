package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
)

// CommunityAudit is the inbound port for the community audit desk.
type CommunityAudit interface {
	Queue(ctx context.Context, actor string, limit int) ([]application.Summary, error)
	Detail(ctx context.Context, actor, id string) (application.Summary, error)
	Evidence(ctx context.Context, actor, id, purpose, correlation string) (application.Evidence, error)
	Decide(ctx context.Context, actor, id, command, why, correlation string, approve bool, expected uint64) (domain.Case, error)
}

// RegisterCommunityAuditRoutes adds the desk's routes. Every one is an
// operator surface; opening evidence and deciding additionally require a
// recent MFA step-up, which the service enforces.
func RegisterCommunityAuditRoutes(mux *http.ServeMux, audit CommunityAudit) {
	mux.Handle("GET /v1/admin/community-audit/cases", auditQueueHandler(audit))
	mux.Handle("GET /v1/admin/community-audit/cases/{id}", auditDetailHandler(audit))
	mux.Handle("POST /v1/admin/community-audit/cases/{id}/evidence", auditEvidenceHandler(audit))
	mux.Handle("POST /v1/admin/community-audit/cases/{id}/decision", auditDecideHandler(audit))
}

type auditSummaryResponse struct {
	CaseID      string `json:"caseId"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	SummaryCode string `json:"summaryCode"`
	CreatedAt   string `json:"createdAt"`
}

func toAuditSummary(summary application.Summary) auditSummaryResponse {
	return auditSummaryResponse{
		CaseID: summary.CaseID, Kind: string(summary.Kind), Status: string(summary.Status),
		// The summary is a code, not prose: the desk renders it from the
		// localization registry so a case listing carries no free text about
		// a member.
		SummaryCode: summary.SummaryCode,
		CreatedAt:   summary.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type auditEvidenceRequest struct {
	Purpose string `json:"purpose"`
}

type auditDecisionRequest struct {
	CommandID        string `json:"commandId"`
	Reason           string `json:"reason"`
	Approve          bool   `json:"approve"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

func auditQueueHandler(audit CommunityAudit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		limit := 50
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 200 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "limit", Reason: "must be between 1 and 200"}},
				})
				return
			}
			limit = parsed
		}
		summaries, err := audit.Queue(r.Context(), session, limit)
		if err != nil {
			writeCommunityAuditError(w, r, err)
			return
		}
		cases := make([]auditSummaryResponse, 0, len(summaries))
		for _, summary := range summaries {
			cases = append(cases, toAuditSummary(summary))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"cases": cases})
	})
}

func auditDetailHandler(audit CommunityAudit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		id := r.PathValue("id")
		if !validOpaqueID(id) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		summary, err := audit.Detail(r.Context(), session, id)
		if err != nil {
			writeCommunityAuditError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toAuditSummary(summary))
	})
}

// auditEvidenceHandler opens evidence about a member. It is a POST rather
// than a GET because it is not a safe read: every access is recorded against
// the operator and the stated purpose.
func auditEvidenceHandler(audit CommunityAudit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		id := r.PathValue("id")
		if !validOpaqueID(id) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body auditEvidenceRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.Purpose = strings.TrimSpace(body.Purpose)
		if body.Purpose == "" || len(body.Purpose) > 200 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "purpose", Reason: "must be 1-200 characters"}},
			})
			return
		}

		evidence, err := audit.Evidence(r.Context(), session, id, body.Purpose, CorrelationID(r.Context()))
		if err != nil {
			writeCommunityAuditError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]string{
			"reference": evidence.Reference, "kind": string(evidence.Kind),
		})
	})
}

func auditDecideHandler(audit CommunityAudit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		id := r.PathValue("id")
		if !validOpaqueID(id) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body auditDecisionRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.Reason = strings.TrimSpace(body.Reason)
		var details []FieldError
		if !validOpaqueID(body.CommandID) {
			details = append(details, FieldError{Field: "commandId", Reason: "must be an opaque identifier"})
		}
		// A decision about a member is auditable, and an audit entry with no
		// stated reason is not one.
		if body.Reason == "" || len(body.Reason) > 500 {
			details = append(details, FieldError{Field: "reason", Reason: "must be 1-500 characters"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		decided, err := audit.Decide(r.Context(), session, id, body.CommandID,
			body.Reason, CorrelationID(r.Context()), body.Approve, body.ExpectedRevision)
		if err != nil {
			writeCommunityAuditError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{
			"caseId": decided.ID(), "status": string(decided.Status()),
		})
	})
}

// writeCommunityAuditError maps the desk's failures. A missing step-up is
// distinguished from a missing role: the first is something the operator can
// fix in a second, the second is not.
func writeCommunityAuditError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrMFA):
		writeError(w, r, http.StatusForbidden, APIError{
			Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before opening this.",
		})
	case errors.Is(err, application.ErrDenied):
		writeError(w, r, http.StatusForbidden, APIError{
			Code: "community_audit_role_required", Message: "This desk requires the trust and safety or admin role.",
		})
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "case_not_found", Message: "That case is not available.",
		})
	case errors.Is(err, application.ErrConflict), errors.Is(err, domain.ErrStale), errors.Is(err, domain.ErrClosed):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "case_conflict", Message: "That case changed. Refresh and try again.",
		})
	case errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
