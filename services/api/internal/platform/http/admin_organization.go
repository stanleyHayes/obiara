package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	organizationapp "github.com/stanleyHayes/obiara/services/api/internal/organization/application"
	organizationdomain "github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
)

// AdminOrganizations is the operator surface for the bodies Obiara has a
// commercial relationship with.
type AdminOrganizations interface {
	Register(context.Context, organizationapp.RegisterCommand) (organizationapp.Result, error)
	Suspend(context.Context, organizationapp.ChangeCommand) (organizationapp.Result, error)
	Restore(context.Context, organizationapp.ChangeCommand) (organizationapp.Result, error)
	Rename(context.Context, organizationapp.ChangeCommand) (organizationapp.Result, error)
	List(ctx context.Context, limit int) ([]organizationdomain.Organization, error)
}

// RegisterAdminOrganizationRoutes exposes registering and managing them.
//
// An operator surface and not a member one. Organizations do not sign in:
// codes are issued by staff on their behalf (agent_plan.md §41), which is why
// there is no organization principal anywhere in this codebase.
//
// Step-up on every write, because these are commercial relationships and the
// audit trail is the whole point of the context — an entry naming an operator
// who had not re-asserted who they were would be worth less than no entry.
func RegisterAdminOrganizationRoutes(
	mux *http.ServeMux, organizations AdminOrganizations, resolve AdminPrincipalResolver,
) {
	mux.Handle("GET /v1/admin/organizations", adminListOrganizationsHandler(organizations, resolve))
	mux.Handle("POST /v1/admin/organizations", adminRegisterOrganizationHandler(organizations, resolve))
	mux.Handle("POST /v1/admin/organizations/{id}/suspend",
		adminChangeOrganizationHandler(organizations, resolve, "suspend"))
	mux.Handle("POST /v1/admin/organizations/{id}/restore",
		adminChangeOrganizationHandler(organizations, resolve, "restore"))
	mux.Handle("PUT /v1/admin/organizations/{id}/name",
		adminChangeOrganizationHandler(organizations, resolve, "rename"))
}

type organizationResponse struct {
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	BillingEmail   string `json:"billingEmail"`
	Status         string `json:"status"`
	Revision       uint64 `json:"revision"`
	RegisteredAt   string `json:"registeredAt"`
}

type organizationListResponse struct {
	Organizations []organizationResponse `json:"organizations"`
}

type registerOrganizationInput struct {
	Name         string `json:"name"`
	BillingEmail string `json:"billingEmail"`
	ReasonCode   string `json:"reasonCode"`
}

type changeOrganizationInput struct {
	ReasonCode string `json:"reasonCode"`
	// ExpectedRevision is what the operator decided against. Without it two
	// operators acting at once would both succeed and the second would erase
	// the first's audit entry.
	ExpectedRevision uint64 `json:"expectedRevision"`
	// Name is read only when renaming.
	Name string `json:"name,omitempty"`
}

func projectOrganization(organization organizationdomain.Organization) organizationResponse {
	return organizationResponse{
		OrganizationID: organization.ID(),
		Name:           organization.Name(),
		BillingEmail:   organization.BillingEmail(),
		Status:         string(organization.Status()),
		Revision:       organization.Revision(),
		RegisteredAt:   organization.RegisteredAt().UTC().Format(time.RFC3339),
	}
}

func adminListOrganizationsHandler(
	organizations AdminOrganizations, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No step-up to read the roster. It holds no member data at all — an
		// organization is a body, not a person — and requiring a fresh MFA
		// to look at a list of universities would train operators to step up
		// out of habit, which is what makes step-up mean nothing.
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if organizations == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		found, err := organizations.List(r.Context(), limit)
		if err != nil {
			writeOrganizationError(w, r, err)
			return
		}
		response := organizationListResponse{
			Organizations: make([]organizationResponse, 0, len(found)),
		}
		for _, organization := range found {
			response.Organizations = append(response.Organizations, projectOrganization(organization))
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func adminRegisterOrganizationHandler(
	organizations AdminOrganizations, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		var body registerOrganizationInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if organizations == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		result, err := organizations.Register(r.Context(), organizationapp.RegisterCommand{
			CommandID: commandID, OperatorID: actorID,
			ReasonCode:   strings.TrimSpace(body.ReasonCode),
			Name:         body.Name,
			BillingEmail: body.BillingEmail,
		})
		if err != nil {
			writeOrganizationError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, projectOrganization(result.Organization))
	})
}

func adminChangeOrganizationHandler(
	organizations AdminOrganizations, resolve AdminPrincipalResolver, action string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		var body changeOrganizationInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if organizations == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		command := organizationapp.ChangeCommand{
			CommandID: commandID, OperatorID: actorID,
			ReasonCode:       strings.TrimSpace(body.ReasonCode),
			OrganizationID:   r.PathValue("id"),
			ExpectedRevision: body.ExpectedRevision,
			Name:             body.Name,
		}
		var result organizationapp.Result
		var err error
		switch action {
		case "suspend":
			result, err = organizations.Suspend(r.Context(), command)
		case "restore":
			result, err = organizations.Restore(r.Context(), command)
		default:
			result, err = organizations.Rename(r.Context(), command)
		}
		if err != nil {
			writeOrganizationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectOrganization(result.Organization))
	})
}

func writeOrganizationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, organizationapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "organization_not_found", Message: "No such organization.",
		})
	case errors.Is(err, organizationapp.ErrNameTaken):
		// Two bodies with one name make the audit trail ambiguous about which
		// of them a code was issued for, which is the one thing this context
		// exists to keep straight.
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "organization_name_taken",
			Message: "An organization with that name is already registered.",
		})
	case errors.Is(err, organizationdomain.ErrInvalidTransition):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "organization_transition_invalid", Message: "That organization is already in this state.",
		})
	case errors.Is(err, organizationdomain.ErrStaleRevision),
		errors.Is(err, organizationapp.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "organization_conflict",
			Message: "That organization changed while you were working. Refresh and try again.",
		})
	case errors.Is(err, organizationdomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, organizationdomain.ErrInvalidOrganization):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, organizationapp.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "feature_unavailable", Message: "This is not available right now.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}

// organizationCommandID reads the retry key.
//
// Required rather than generated, because it is what makes a retried
// suspension one suspension. Generating one here would make every retry a new
// command and put a second entry in the audit trail for one decision.
func organizationCommandID(w http.ResponseWriter, r *http.Request) (string, bool) {
	commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if commandID == "" {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
			Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
		})
		return "", false
	}
	return commandID, true
}
