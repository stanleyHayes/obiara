package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	safetyapp "github.com/stanleyHayes/obiara/internal/safety/application"
	safetydomain "github.com/stanleyHayes/obiara/internal/safety/domain"
)

type AdminCareCases interface {
	NextOpen(context.Context, int) ([]safetydomain.CareCase, error)
	Engage(context.Context, string) (safetydomain.CareCase, error)
	Resolve(context.Context, string, []safetydomain.ScriptKey) (safetydomain.CareCase, error)
}

func RegisterAdminCareRoutes(mux *http.ServeMux, care AdminCareCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/care/cases", adminCareQueueHandler(care, keyer, resolve))
	mux.Handle("POST /v1/admin/care/cases/{id}/engagement", adminCareEngageHandler(care, keyer, resolve))
	mux.Handle("POST /v1/admin/care/cases/{id}/resolution", adminCareResolveHandler(care, keyer, resolve))
}

type adminCareCaseResponse struct {
	CaseID     string   `json:"caseId"`
	SubjectRef string   `json:"subjectRef"`
	Signal     string   `json:"signal"`
	Status     string   `json:"status"`
	Scripts    []string `json:"scripts"`
	CreatedAt  string   `json:"createdAt"`
	Version    int64    `json:"version"`
}

func careCaseView(value safetydomain.CareCase, keyer SafetySubjectKeyer) (adminCareCaseResponse, error) {
	subject, err := keyer.MemberKey(value.SubjectID())
	if err != nil {
		return adminCareCaseResponse{}, err
	}
	scripts := make([]string, 0, len(value.Scripts()))
	for _, script := range value.Scripts() {
		scripts = append(scripts, string(script))
	}
	return adminCareCaseResponse{
		CaseID: value.ID(), SubjectRef: subject, Signal: string(value.Signal()),
		Status: string(value.Status()), Scripts: scripts,
		CreatedAt: value.CreatedAt().UTC().Format(time.RFC3339), Version: value.Version(),
	}, nil
}

func adminCareQueueHandler(care AdminCareCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSafetyAdmin(w, r, resolve, false); !ok {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 50
		}
		found, err := care.NextOpen(r.Context(), limit)
		if err != nil {
			writeAdminCareError(w, r, err)
			return
		}
		items := make([]adminCareCaseResponse, 0, len(found))
		for _, item := range found {
			view, viewErr := careCaseView(item, keyer)
			if viewErr != nil {
				writeAdminCareError(w, r, viewErr)
				return
			}
			items = append(items, view)
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"cases": items})
	})
}

func adminCareEngageHandler(care AdminCareCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSafetyAdmin(w, r, resolve, false); !ok || !adminJSONGuard(w, r) {
			return
		}
		engaged, err := care.Engage(r.Context(), r.PathValue("id"))
		if err != nil {
			writeAdminCareError(w, r, err)
			return
		}
		view, err := careCaseView(engaged, keyer)
		if err != nil {
			writeAdminCareError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, view)
	})
}

type adminCareResolutionInput struct {
	Scripts []safetydomain.ScriptKey `json:"scripts"`
}

func adminCareResolveHandler(care AdminCareCases, keyer SafetySubjectKeyer, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSafetyAdmin(w, r, resolve, true); !ok || !adminJSONGuard(w, r) {
			return
		}
		var body adminCareResolutionInput
		if decodeJSON(w, r, &body) != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		resolved, err := care.Resolve(r.Context(), r.PathValue("id"), body.Scripts)
		if err != nil {
			writeAdminCareError(w, r, err)
			return
		}
		view, err := careCaseView(resolved, keyer)
		if err != nil {
			writeAdminCareError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, view)
	})
}

func writeAdminCareError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, safetyapp.ErrCaseNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "care_case_not_found", Message: "No such care case."})
	case errors.Is(err, safetydomain.ErrCaseNotOpen):
		writeError(w, r, http.StatusConflict, APIError{Code: "care_case_conflict", Message: "The care case is no longer available for that action."})
	case errors.Is(err, safetydomain.ErrScriptRequired), errors.Is(err, safetydomain.ErrInvalidScript):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Choose at least one approved care resource."})
	default:
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "care_unavailable", Message: "The care queue is temporarily unavailable."})
	}
}
