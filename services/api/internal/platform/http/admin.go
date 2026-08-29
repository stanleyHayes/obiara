package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

// Admin is the inbound port for admin auth (E16-S01).
type Admin interface {
	Enroll(ctx context.Context, sessionID, email string, roles []domain.Role) (domain.Principal, bool, error)
	ListPrincipals(ctx context.Context, sessionID string) ([]domain.Principal, error)
	ChangeStatus(ctx context.Context, sessionID, targetID string, status domain.Status) (domain.Principal, error)
	ChangeRoles(ctx context.Context, sessionID, targetID string, roles []domain.Role, expectedVersion *int64) (domain.Principal, error)
	ProposeAdminRoleChange(ctx context.Context, sessionID, targetID string, roles []domain.Role, reason string) (domain.RoleChange, error)
	ListPendingRoleChanges(ctx context.Context, sessionID string) ([]domain.RoleChange, error)
	ApproveAdminRoleChange(ctx context.Context, sessionID, changeID string) (domain.Principal, error)
	StartLogin(ctx context.Context, email, password string) error
	CompleteLogin(ctx context.Context, email, code string) (domain.Session, error)
	StepUpStart(ctx context.Context, sessionID string) error
	StepUpComplete(ctx context.Context, sessionID, code string) (domain.Session, error)
	Authenticate(ctx context.Context, sessionID string) (domain.Session, domain.Principal, error)
}

// RegisterAdminRoutes adds the admin auth baseline routes.
func RegisterAdminRoutes(mux *http.ServeMux, admin Admin) {
	mux.Handle("GET /v1/admin/account", adminAccountHandler(admin))
	mux.Handle("POST /v1/admin/principals", enrollHandler(admin))
	mux.Handle("GET /v1/admin/principals", listPrincipalsHandler(admin))
	mux.Handle("PATCH /v1/admin/principals/{id}", updatePrincipalHandler(admin))
	mux.Handle("POST /v1/admin/principals/{id}/role-changes", proposeRoleChangeHandler(admin))
	mux.Handle("GET /v1/admin/role-changes", listRoleChangesHandler(admin))
	mux.Handle("POST /v1/admin/role-changes/{id}/approve", approveRoleChangeHandler(admin))
	mux.Handle("POST /v1/admin/login/start", loginStartHandler(admin))
	mux.Handle("POST /v1/admin/login/complete", loginCompleteHandler(admin))
	mux.Handle("POST /v1/admin/sessions/{id}/step-up/start", stepUpStartHandler(admin))
	mux.Handle("POST /v1/admin/sessions/{id}/step-up/complete", stepUpCompleteHandler(admin))
}

type AdminAccountAuthenticator interface {
	Authenticate(context.Context, string) (domain.Session, domain.Principal, error)
}

type adminAccountResponse struct {
	Email          string    `json:"email"`
	Roles          []string  `json:"roles"`
	Status         string    `json:"status"`
	OperatorSince  time.Time `json:"operatorSince"`
	SessionCreated time.Time `json:"sessionCreated"`
	SessionExpires time.Time `json:"sessionExpires"`
	SteppedUp      bool      `json:"steppedUp"`
}

func adminAccountHandler(admin AdminAccountAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			writeError(w, r, http.StatusUnauthorized, APIError{Code: "admin_authentication_required", Message: "A valid admin session is required."})
			return
		}
		session, principal, err := admin.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, APIError{Code: "admin_authentication_required", Message: "A valid admin session is required."})
			return
		}
		roles := make([]string, 0, len(principal.Roles()))
		for _, role := range principal.Roles() {
			roles = append(roles, string(role))
		}
		writeSuccess(w, r, http.StatusOK, adminAccountResponse{
			Email: principal.Email(), Roles: roles, Status: string(principal.Status()),
			OperatorSince: principal.CreatedAt(), SessionCreated: session.CreatedAt(),
			SessionExpires: session.ExpiresAt(), SteppedUp: session.SteppedUp(),
		})
	})
}

type roleChangeResponse struct {
	ChangeID      string     `json:"changeId"`
	TargetID      string     `json:"targetId"`
	TargetVersion int64      `json:"targetVersion"`
	Roles         []string   `json:"roles"`
	Reason        string     `json:"reason"`
	ProposerID    string     `json:"proposerId"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	ApprovedAt    *time.Time `json:"approvedAt,omitempty"`
}

func toRoleChangeResponse(change domain.RoleChange) roleChangeResponse {
	roles := make([]string, 0, len(change.Roles()))
	for _, role := range change.Roles() {
		roles = append(roles, string(role))
	}
	return roleChangeResponse{
		ChangeID: change.ID(), TargetID: change.TargetID(), TargetVersion: change.TargetVersion(),
		Roles: roles, Reason: change.Reason(), ProposerID: change.ProposerID(),
		Status: string(change.Status()), CreatedAt: change.CreatedAt(), ApprovedAt: change.ApprovedAt(),
	}
}

func adminJSONGuard(w http.ResponseWriter, r *http.Request) bool {
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{
			Code:    "unsupported_media_type",
			Message: "Content-Type must be application/json.",
		})
		return false
	}
	return true
}

type enrollRequest struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

type principalResponse struct {
	PrincipalID string    `json:"principalId"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Status      string    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toPrincipalResponse(principal domain.Principal) principalResponse {
	roleNames := make([]string, 0, len(principal.Roles()))
	for _, role := range principal.Roles() {
		roleNames = append(roleNames, string(role))
	}
	return principalResponse{
		PrincipalID: principal.ID(), Email: principal.Email(), Roles: roleNames,
		Status: string(principal.Status()), Version: principal.Version(), CreatedAt: principal.CreatedAt(),
	}
}

// enrolledPrincipalResponse adds the one fact only enrollment knows: whether
// the new operator was actually told.
type enrolledPrincipalResponse struct {
	principalResponse
	Invited bool `json:"invited"`
}

func enrollHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		var body enrollRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		body.Email = strings.TrimSpace(body.Email)
		if len(body.Roles) == 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "roles", Reason: "at least one role is required"}},
			})
			return
		}
		roles := make([]domain.Role, 0, len(body.Roles))
		for _, role := range body.Roles {
			roles = append(roles, domain.Role(role))
		}
		principal, invited, err := admin.Enroll(r.Context(), session.ID(), body.Email, roles)
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		// invited is reported rather than folded into the status: the
		// operator exists and is audited either way, and an admin who is not
		// told the invitation bounced will assume the person was notified.
		writeSuccess(w, r, http.StatusCreated, enrolledPrincipalResponse{
			principalResponse: toPrincipalResponse(principal),
			Invited:           invited,
		})
	})
}

func listPrincipalsHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		principals, err := admin.ListPrincipals(r.Context(), session.ID())
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		items := make([]principalResponse, 0, len(principals))
		for _, principal := range principals {
			items = append(items, toPrincipalResponse(principal))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"items": items})
	})
}

type updatePrincipalRequest struct {
	Action string   `json:"action"`
	Status string   `json:"status"`
	Roles  []string `json:"roles"`
	Reason string   `json:"reason"`
	// ExpectedVersion is the principal revision the console displayed when
	// the operator chose this change. Absent means "apply regardless",
	// which is what pre-existing callers send.
	ExpectedVersion *int64 `json:"expectedVersion"`
}

func updatePrincipalHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		var body updatePrincipalRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if len(strings.TrimSpace(body.Reason)) < 12 || len(strings.TrimSpace(body.Reason)) > 240 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Record a reason between 12 and 240 characters."})
			return
		}
		var principal domain.Principal
		var err error
		switch body.Action {
		case "status":
			principal, err = admin.ChangeStatus(r.Context(), session.ID(), r.PathValue("id"), domain.Status(body.Status))
		case "roles":
			roles := make([]domain.Role, 0, len(body.Roles))
			for _, role := range body.Roles {
				roles = append(roles, domain.Role(role))
			}
			principal, err = admin.ChangeRoles(r.Context(), session.ID(), r.PathValue("id"), roles, body.ExpectedVersion)
		default:
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Choose a supported principal action."})
			return
		}
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPrincipalResponse(principal))
	})
}

type proposeRoleChangeRequest struct {
	Roles  []string `json:"roles"`
	Reason string   `json:"reason"`
}

func proposeRoleChangeHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		var body proposeRoleChangeRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		roles := make([]domain.Role, 0, len(body.Roles))
		for _, role := range body.Roles {
			roles = append(roles, domain.Role(role))
		}
		change, err := admin.ProposeAdminRoleChange(r.Context(), session.ID(), r.PathValue("id"), roles, body.Reason)
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toRoleChangeResponse(change))
	})
}

func listRoleChangesHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		changes, err := admin.ListPendingRoleChanges(r.Context(), session.ID())
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		items := make([]roleChangeResponse, 0, len(changes))
		for _, change := range changes {
			items = append(items, toRoleChangeResponse(change))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"items": items})
	})
}

func approveRoleChangeHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		principal, err := admin.ApproveAdminRoleChange(r.Context(), session.ID(), r.PathValue("id"))
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPrincipalResponse(principal))
	})
}

type emailRequest struct {
	Email string `json:"email"`
	// Password is optional on the wire so operators enrolled before
	// password support can still sign in on the code alone. The service
	// decides whether it is required for the named principal.
	Password string `json:"password"`
}

func loginStartHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body emailRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if strings.TrimSpace(body.Email) == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "One or more fields are invalid."})
			return
		}
		// The password is deliberately not trimmed: leading and trailing
		// whitespace are legitimate characters the operator chose.
		if err := admin.StartLogin(r.Context(), strings.TrimSpace(body.Email), body.Password); err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, map[string]string{"status": "code_sent"})
	})
}

type loginCompleteRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type adminSessionResponse struct {
	SessionID string    `json:"sessionId"`
	Roles     []string  `json:"roles"`
	SteppedUp bool      `json:"steppedUp"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func toSessionResponse(session domain.Session) adminSessionResponse {
	roles := make([]string, 0, len(session.Roles()))
	for _, role := range session.Roles() {
		roles = append(roles, string(role))
	}
	return adminSessionResponse{SessionID: session.ID(), Roles: roles, SteppedUp: session.SteppedUp(), ExpiresAt: session.ExpiresAt()}
}

func loginCompleteHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		var body loginCompleteRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		session, err := admin.CompleteLogin(r.Context(), strings.TrimSpace(body.Email), strings.TrimSpace(body.Code))
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSessionResponse(session))
	})
}

func stepUpStartHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if session.ID() != r.PathValue("id") {
			writeError(w, r, http.StatusForbidden, APIError{Code: "admin_session_mismatch", Message: "You can only step up your current session."})
			return
		}
		// No JSON guard: this endpoint reads no body. Requiring a JSON
		// Content-Type on a bodyless POST rejected every real caller with
		// 415 before it did anything, and because enrollment, role changes
		// and suspensions all require a stepped-up session, that one guard
		// made the whole privileged surface unreachable.
		if err := admin.StepUpStart(r.Context(), r.PathValue("id")); err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, map[string]string{"status": "code_sent"})
	})
}

type stepUpCompleteRequest struct {
	Code string `json:"code"`
}

func stepUpCompleteHandler(admin Admin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		if session.ID() != r.PathValue("id") {
			writeError(w, r, http.StatusForbidden, APIError{Code: "admin_session_mismatch", Message: "You can only step up your current session."})
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		var body stepUpCompleteRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		session, err := admin.StepUpComplete(r.Context(), r.PathValue("id"), strings.TrimSpace(body.Code))
		if err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSessionResponse(session))
	})
}

func authenticatedAdmin(w http.ResponseWriter, r *http.Request, admin Admin) (domain.Session, domain.Principal, bool) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "admin_authentication_required", Message: "A valid admin session is required."})
		return domain.Session{}, domain.Principal{}, false
	}
	session, principal, err := admin.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "admin_authentication_required", Message: "A valid admin session is required."})
		return domain.Session{}, domain.Principal{}, false
	}
	return session, principal, true
}

func writeAdminError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrCodeDeliveryFailed):
		// The code was minted but the email provider rejected it. Retrying
		// cannot help until the provider is fixed, so this is reported as an
		// unavailable dependency and the real cause goes to the logs.
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "mfa_delivery_failed", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code:    "mfa_delivery_failed",
			Message: "The sign-in code could not be delivered. The email provider is unavailable or misconfigured.",
		})
	case errors.Is(err, application.ErrNotAdmin):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "admin_role_required",
			Message: "Enrollment requires the admin role.",
		})
	case errors.Is(err, application.ErrStepUpRequired):
		writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before changing operator access."})
	case errors.Is(err, application.ErrSelfSuspension):
		writeError(w, r, http.StatusConflict, APIError{Code: "self_suspension_forbidden", Message: "You cannot suspend your own admin principal."})
	case errors.Is(err, application.ErrLastAdmin):
		writeError(w, r, http.StatusConflict, APIError{Code: "last_admin_required", Message: "The last active admin cannot be suspended."})
	case errors.Is(err, application.ErrFourEyesRequired):
		writeError(w, r, http.StatusConflict, APIError{Code: "four_eyes_required", Message: "Admin-role changes require approval by a distinct stepped-up administrator."})
	case errors.Is(err, application.ErrPrincipalConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "principal_conflict", Message: "The operator changed. Refresh and try again."})
	case errors.Is(err, application.ErrRoleChangeNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "role_change_not_found", Message: "The role-change proposal was not found."})
	case errors.Is(err, domain.ErrSameApprover):
		writeError(w, r, http.StatusConflict, APIError{Code: "distinct_approver_required", Message: "A different stepped-up administrator must approve this change."})
	case errors.Is(err, domain.ErrRoleChangeClosed):
		writeError(w, r, http.StatusConflict, APIError{Code: "role_change_closed", Message: "This role-change proposal is already closed."})
	case errors.Is(err, application.ErrPrincipalExists):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "principal_exists",
			Message: "A principal with that email already exists.",
		})
	case errors.Is(err, domain.ErrMfaMismatch), errors.Is(err, domain.ErrMfaConsumed), errors.Is(err, application.ErrChallengeNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_invalid",
			Message: "The code is not valid for this account.",
		})
	case errors.Is(err, domain.ErrMfaExpired):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_expired",
			Message: "The code has expired. Request a new one.",
		})
	case errors.Is(err, domain.ErrMfaAttempts):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "mfa_attempts_exceeded",
			Message: "Too many wrong attempts. Request a new code.",
		})
	case errors.Is(err, domain.ErrSessionClosed), errors.Is(err, application.ErrSessionNotFound):
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code:    "session_closed",
			Message: "The admin session is not active.",
		})
	case errors.Is(err, application.ErrPrincipalNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "principal_not_found",
			Message: "No admin principal for that account.",
		})
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrNoRoles), errors.Is(err, domain.ErrInvalidStatus), errors.Is(err, domain.ErrRoleChangeReason):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	default:
		// Unmapped errors are genuine faults. The envelope stays opaque; the
		// cause goes to the logs so a 500 is never a dead end during triage.
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
