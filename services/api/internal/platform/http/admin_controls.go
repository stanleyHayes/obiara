package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	flagapp "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/application"
	flagdomain "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type AdminFlagControls interface {
	Propose(context.Context, flagapp.ProposeCommand) (flagdomain.Proposal, error)
	Approve(context.Context, string, string) (flagdomain.Proposal, error)
	Apply(context.Context, string, string) (flagdomain.Proposal, error)
}

type AdminFlagControlReader interface {
	ListActive(context.Context, int64) ([]flagdomain.Proposal, error)
}

type ControlActorKeyer interface {
	Key(namespace, value string) (string, error)
}

func RegisterAdminControlRoutes(mux *http.ServeMux, controls AdminFlagControls, reader AdminFlagControlReader, keyer ControlActorKeyer, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/controls", adminControlListHandler(reader, keyer, resolve))
	mux.Handle("POST /v1/admin/controls", adminControlProposeHandler(controls, keyer, resolve))
	mux.Handle("POST /v1/admin/controls/{id}/approval", adminControlApproveHandler(controls, keyer, resolve))
	mux.Handle("POST /v1/admin/controls/{id}/application", adminControlApplyHandler(controls, keyer, resolve))
}

func requireControlAdmin(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver) (admin.Principal, string, bool) {
	principal, ok := resolveAdminPrincipal(w, r, resolve)
	if !ok {
		return admin.Principal{}, "", false
	}
	if !principal.Has(admin.ScopeOperations) {
		writeAdminVerificationError(w, r, admin.ErrForbidden)
		return admin.Principal{}, "", false
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		writeAdminVerificationError(w, r, admin.ErrForbidden)
		return admin.Principal{}, "", false
	}
	return principal, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")), true
}

func requireControlStepUp(w http.ResponseWriter, r *http.Request, principal admin.Principal) bool {
	if principal.MFAVerified {
		return true
	}
	writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before changing runtime controls."})
	return false
}

type adminControlResponse struct {
	ProposalID   string `json:"proposalId"`
	Capability   string `json:"capability"`
	Environment  string `json:"environment"`
	Market       string `json:"market"`
	Action       string `json:"action"`
	Reason       string `json:"reason"`
	Status       string `json:"status"`
	Version      uint64 `json:"version"`
	ExpiresAt    string `json:"expiresAt"`
	ProposedByMe bool   `json:"proposedByMe"`
	ApprovedByMe bool   `json:"approvedByMe"`
}

func controlView(value flagdomain.Proposal, actorKey string) adminControlResponse {
	state := value.State()
	return adminControlResponse{
		ProposalID: value.ID(), Capability: string(state.Capability),
		Environment: string(state.Environment), Market: string(state.Market),
		Action: string(state.Action), Reason: string(state.Reason),
		Status: string(state.Status), Version: state.Version,
		ExpiresAt:    value.ExpiresAt().UTC().Format(time.RFC3339),
		ProposedByMe: state.ProposerKey == actorKey,
		ApprovedByMe: state.ApproverKey == actorKey,
	}
}

func adminControlListHandler(reader AdminFlagControlReader, keyer ControlActorKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, _, ok := requireControlAdmin(w, r, resolve)
		if !ok {
			return
		}
		actorKey, err := keyer.Key("flag-controller", principal.ActorID)
		if err != nil {
			writeAdminControlError(w, r, err)
			return
		}
		found, err := reader.ListActive(r.Context(), 100)
		if err != nil {
			writeAdminControlError(w, r, err)
			return
		}
		items := make([]adminControlResponse, 0, len(found))
		for _, item := range found {
			items = append(items, controlView(item, actorKey))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"proposals": items})
	})
}

type adminControlProposalInput struct {
	CommandID   string `json:"commandId"`
	Capability  string `json:"capability"`
	Environment string `json:"environment"`
	Market      string `json:"market"`
	Action      string `json:"action"`
	Reason      string `json:"reason"`
}

func adminControlProposeHandler(controls AdminFlagControls, keyer ControlActorKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, session, ok := requireControlAdmin(w, r, resolve)
		if !ok || !requireControlStepUp(w, r, principal) || !adminJSONGuard(w, r) {
			return
		}
		var body adminControlProposalInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		proposal, err := controls.Propose(r.Context(), flagapp.ProposeCommand{
			CommandID: body.CommandID, SessionID: session,
			Capability:  flagdomain.Capability(body.Capability),
			Environment: flagdomain.Environment(body.Environment),
			Market:      flagdomain.Market(body.Market), Action: flagdomain.Action(body.Action),
			Reason: flagdomain.Reason(body.Reason),
		})
		if err != nil {
			writeAdminControlError(w, r, err)
			return
		}
		actorKey, _ := keyer.Key("flag-controller", principal.ActorID)
		writeSuccess(w, r, http.StatusCreated, controlView(proposal, actorKey))
	})
}

func adminControlApproveHandler(controls AdminFlagControls, keyer ControlActorKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, session, ok := requireControlAdmin(w, r, resolve)
		if !ok || !requireControlStepUp(w, r, principal) || !adminJSONGuard(w, r) {
			return
		}
		proposal, err := controls.Approve(r.Context(), r.PathValue("id"), session)
		if err != nil {
			writeAdminControlError(w, r, err)
			return
		}
		actorKey, _ := keyer.Key("flag-controller", principal.ActorID)
		writeSuccess(w, r, http.StatusOK, controlView(proposal, actorKey))
	})
}

func adminControlApplyHandler(controls AdminFlagControls, keyer ControlActorKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, session, ok := requireControlAdmin(w, r, resolve)
		if !ok || !requireControlStepUp(w, r, principal) || !adminJSONGuard(w, r) {
			return
		}
		proposal, err := controls.Apply(r.Context(), r.PathValue("id"), session)
		if err != nil {
			writeAdminControlError(w, r, err)
			return
		}
		actorKey, _ := keyer.Key("flag-controller", principal.ActorID)
		writeSuccess(w, r, http.StatusOK, controlView(proposal, actorKey))
	})
}

func writeAdminControlError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, flagapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "control_not_found", Message: "No such runtime-control proposal."})
	case errors.Is(err, flagdomain.ErrSameActor):
		writeError(w, r, http.StatusForbidden, APIError{Code: "distinct_approver_required", Message: "A different stepped-up administrator must approve this proposal."})
	case errors.Is(err, flagdomain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Choose one bounded capability, environment, market, action and reason."})
	case errors.Is(err, flagdomain.ErrExpired), errors.Is(err, flagdomain.ErrState), errors.Is(err, flagapp.ErrConflict), errors.Is(err, flagapp.ErrApplied):
		writeError(w, r, http.StatusConflict, APIError{Code: "control_conflict", Message: "The proposal is expired or no longer available for that transition."})
	default:
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "controls_unavailable", Message: "Runtime controls are temporarily unavailable."})
	}
}
