package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	safetyapp "github.com/stanleyHayes/obiara/internal/safety/application"
	safetydomain "github.com/stanleyHayes/obiara/internal/safety/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type AdminSafetyCases interface {
	NextQueued(context.Context, safetydomain.Queue, int) ([]safetydomain.Case, error)
	Find(context.Context, string) (safetydomain.Case, error)
	Assign(context.Context, string, string) (safetydomain.Case, error)
}

type AdminSafetyEvidence interface {
	View(context.Context, string, string, safetydomain.Purpose) (safetydomain.Bundle, error)
}

type SafetySubjectKeyer interface {
	MemberKey(string) (string, error)
}

// AdminSafetyActions applies the T&S ladder. The ladder itself is enforced
// inside the safety context; this transport only carries the decision and
// the request that made it.
type AdminSafetyActions interface {
	Apply(ctx context.Context, caseID string, action safetydomain.Action, actorID, commandID string) error
}

func RegisterAdminSafetyRoutes(mux *http.ServeMux, cases AdminSafetyCases, evidence AdminSafetyEvidence, actions AdminSafetyActions, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/safety/cases", adminSafetyQueueHandler(cases, keyer, resolve))
	mux.Handle("POST /v1/admin/safety/cases/{id}/assignment", adminSafetyAssignHandler(cases, keyer, resolve))
	mux.Handle("POST /v1/admin/safety/cases/{id}/evidence-access", adminSafetyEvidenceHandler(cases, evidence, keyer, resolve))
	// Until this existed a case could be queued, assigned and its evidence
	// read, and then nothing could happen to it. The ladder was enforced in
	// a service no route reached.
	mux.Handle("POST /v1/admin/safety/cases/{id}/actions", adminSafetyActionHandler(actions, resolve))
}

type safetyActionInput struct {
	Action string `json:"action"`
}

// adminSafetyActionHandler applies one ladder action to one case.
//
// Step-up is required for every action, not only bans. A suspension takes a
// member off the product for weeks, and the desk should have to re-assert who
// it is before it does that.
func adminSafetyActionHandler(actions AdminSafetyActions, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		// The request id is what makes the action idempotent. Without one, a
		// double submission would be logged twice and escalate this member's
		// next action for a decision taken once, so it is required rather
		// than defaulted.
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if commandID == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
			})
			return
		}
		var body safetyActionInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		action := safetydomain.Action(strings.TrimSpace(body.Action))
		if !knownSafetyAction(action) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "action", Reason: "must be a ladder action"}},
			})
			return
		}
		if actions == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		if err := actions.Apply(r.Context(), r.PathValue("id"), action, principal.ActorID, commandID); err != nil {
			writeSafetyActionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]string{"action": string(action)})
	})
}

func knownSafetyAction(action safetydomain.Action) bool {
	switch action {
	case safetydomain.ActionWarning, safetydomain.ActionSuspend14d,
		safetydomain.ActionSuspend30d, safetydomain.ActionSuspend90d, safetydomain.ActionBan:
		return true
	}
	return false
}

func writeSafetyActionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, safetydomain.ErrActionNotOnLadder):
		// A refusal of the operator's choice, not of the operator. The
		// message says the ladder decided, because that is the point of
		// having one: the desk does not get to pick the severity freely.
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "action_not_on_ladder",
			Message: "That action is not the one this case's tier allows next.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}

func requireSafetyAdmin(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver, stepUp bool) (admin.Principal, bool) {
	principal, ok := resolveAdminPrincipal(w, r, resolve)
	if !ok {
		return admin.Principal{}, false
	}
	if !principal.Has(admin.ScopeSafety) {
		writeAdminVerificationError(w, r, admin.ErrForbidden)
		return admin.Principal{}, false
	}
	if stepUp && !principal.MFAVerified {
		writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before opening safety evidence."})
		return admin.Principal{}, false
	}
	return principal, true
}

type adminSafetyCaseResponse struct {
	CaseID       string `json:"caseId"`
	SubjectRef   string `json:"subjectRef"`
	Tier         string `json:"tier"`
	Queue        string `json:"queue"`
	Status       string `json:"status"`
	SLADueAt     string `json:"slaDueAt"`
	Assigned     bool   `json:"assigned"`
	AssignedToMe bool   `json:"assignedToMe"`
	Version      int64  `json:"version"`
}

func safetyCaseView(value safetydomain.Case, keyer SafetySubjectKeyer, actorID string) (adminSafetyCaseResponse, error) {
	subject, err := keyer.MemberKey(value.SubjectID())
	if err != nil {
		return adminSafetyCaseResponse{}, err
	}
	return adminSafetyCaseResponse{
		CaseID: value.ID(), SubjectRef: subject, Tier: string(value.Tier()),
		Queue: string(value.Queue()), Status: string(value.Status()),
		SLADueAt: value.SLADueAt().UTC().Format(time.RFC3339),
		Assigned: value.AssignedTo() != "", AssignedToMe: value.AssignedTo() == actorID,
		Version: value.Version(),
	}, nil
}

func adminSafetyQueueHandler(cases AdminSafetyCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, false)
		if !ok {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 50
		}
		found, err := cases.NextQueued(r.Context(), safetydomain.QueueTriage, limit)
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		items := make([]adminSafetyCaseResponse, 0, len(found))
		for _, item := range found {
			view, viewErr := safetyCaseView(item, keyer, principal.ActorID)
			if viewErr != nil {
				writeAdminSafetyError(w, r, viewErr)
				return
			}
			items = append(items, view)
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"cases": items})
	})
}

func adminSafetyAssignHandler(cases AdminSafetyCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, false)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		assigned, err := cases.Assign(r.Context(), r.PathValue("id"), principal.ActorID)
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		view, err := safetyCaseView(assigned, keyer, principal.ActorID)
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, view)
	})
}

type safetyEvidenceInput struct {
	Purpose string `json:"purpose"`
}

type safetyEvidenceResponse struct {
	CaseID      string `json:"caseId"`
	SubjectRef  string `json:"subjectRef"`
	Tier        string `json:"tier"`
	Category    string `json:"category"`
	Surface     string `json:"surface"`
	ContextRef  string `json:"contextRef,omitempty"`
	Description string `json:"description,omitempty"`
}

func adminSafetyEvidenceHandler(cases AdminSafetyCases, evidence AdminSafetyEvidence, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requireSafetyAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		var body safetyEvidenceInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		current, err := cases.Find(r.Context(), r.PathValue("id"))
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		if current.AssignedTo() != principal.ActorID {
			writeError(w, r, http.StatusForbidden, APIError{Code: "safety_assignment_required", Message: "Assign this case to yourself before opening evidence."})
			return
		}
		purpose := safetydomain.Purpose(strings.TrimSpace(body.Purpose))
		bundle, err := evidence.View(r.Context(), current.ID(), principal.ActorID, purpose)
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		subject, err := keyer.MemberKey(bundle.SubjectID)
		if err != nil {
			writeAdminSafetyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, safetyEvidenceResponse{
			CaseID: current.ID(), SubjectRef: subject, Tier: string(bundle.Tier),
			Category: string(bundle.Category), Surface: string(bundle.Surface),
			ContextRef: bundle.ContextRef, Description: bundle.Description,
		})
	})
}

func writeAdminSafetyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, safetyapp.ErrCaseNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "safety_case_not_found", Message: "No such safety case."})
	case errors.Is(err, safetydomain.ErrCaseNotOpen):
		writeError(w, r, http.StatusConflict, APIError{Code: "safety_case_conflict", Message: "The safety case is no longer available for that action."})
	case errors.Is(err, safetydomain.ErrInvalidPurpose):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Choose triage, appeal or legal as the evidence purpose."})
	default:
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "safety_unavailable", Message: "The safety desk is temporarily unavailable."})
	}
}
