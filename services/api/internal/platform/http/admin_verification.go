package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type AdminVerification interface {
	ListQueue(context.Context, admin.Principal, int) ([]admin.CaseSummary, error)
	Detail(context.Context, admin.Principal, string) (admin.CaseDetail, error)
	OpenEvidence(context.Context, admin.Principal, string, string, string, string) (admin.Evidence, error)
	Decide(context.Context, admin.Principal, string, admin.Outcome, string, string, string, int64) (admin.DecisionResult, error)
}

type AdminPrincipalResolver func(*http.Request) (admin.Principal, error)

const adminOperationsScope = admin.ScopeOperations

var errAdminOperationsForbidden = admin.ErrForbidden

func RegisterAdminVerificationRoutes(mux *http.ServeMux, service AdminVerification, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/verifications", adminVerificationQueueHandler(service, resolve))
	mux.Handle("GET /v1/admin/verifications/{id}", adminVerificationDetailHandler(service, resolve))
	mux.Handle("POST /v1/admin/verifications/{id}/evidence-access", adminVerificationEvidenceHandler(service, resolve))
	mux.Handle("POST /v1/admin/verifications/{id}/decisions", adminVerificationDecisionHandler(service, resolve))
}

type adminCaseResponse struct {
	CaseID      string `json:"caseId"`
	SubjectRef  string `json:"subjectRef"`
	ReasonCode  string `json:"reasonCode"`
	SubmittedAt string `json:"submittedAt"`
	Status      string `json:"status,omitempty"`
	Version     int64  `json:"version"`
}

type adminQueueResponse struct {
	Cases []adminCaseResponse `json:"cases"`
}

type evidenceAccessBody struct {
	Purpose string `json:"purpose"`
	Reason  string `json:"reason"`
}

type evidenceResponse struct {
	CaseID     string `json:"caseId"`
	MaskedCard string `json:"maskedCard"`
	AgeBand    string `json:"ageBand"`
	// These outlive the birth date the band comes from. After retention has
	// stripped it the band reads "unknown", and these are what still show
	// that the check happened and how strong it was.
	AgeAssuranceMethod string `json:"ageAssuranceMethod,omitempty"`
	AgeAssuredAt       string `json:"ageAssuredAt,omitempty"`
	ProviderStatus     string `json:"providerStatus"`
}

type decisionBody struct {
	Outcome         admin.Outcome `json:"outcome"`
	Reason          string        `json:"reason"`
	ExpectedVersion int64         `json:"expectedVersion"`
}

type decisionResponse struct {
	Case     adminCaseResponse `json:"case"`
	Outcome  admin.Outcome     `json:"outcome"`
	Replayed bool              `json:"replayed"`
}

func adminVerificationQueueHandler(service AdminVerification, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		cases, err := service.ListQueue(r.Context(), principal, limit)
		if err != nil {
			writeAdminVerificationError(w, r, err)
			return
		}
		response := adminQueueResponse{Cases: make([]adminCaseResponse, 0, len(cases))}
		for _, item := range cases {
			response.Cases = append(response.Cases, toAdminCaseResponse(admin.CaseDetail{CaseSummary: item}))
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func adminVerificationDetailHandler(service AdminVerification, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		detail, err := service.Detail(r.Context(), principal, r.PathValue("id"))
		if err != nil {
			writeAdminVerificationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toAdminCaseResponse(detail))
	})
}

func adminVerificationEvidenceHandler(service AdminVerification, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !requireJSON(w, r) {
			return
		}
		var body evidenceAccessBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		evidence, err := service.OpenEvidence(
			r.Context(), principal, r.PathValue("id"), strings.TrimSpace(body.Purpose),
			strings.TrimSpace(body.Reason), r.Header.Get("X-Correlation-ID"),
		)
		if err != nil {
			writeAdminVerificationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, evidenceResponse{
			CaseID: evidence.CaseID, MaskedCard: evidence.MaskedCard,
			AgeBand:            evidence.AgeBand,
			AgeAssuranceMethod: evidence.AgeAssuranceMethod,
			AgeAssuredAt:       evidence.AgeAssuredAt,
			ProviderStatus:     evidence.ProviderStatus,
		})
	})
}

func adminVerificationDecisionHandler(service AdminVerification, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !requireJSON(w, r) {
			return
		}
		var body decisionBody
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		result, err := service.Decide(
			r.Context(), principal, r.PathValue("id"), body.Outcome,
			strings.TrimSpace(body.Reason), r.Header.Get("Idempotency-Key"),
			r.Header.Get("X-Correlation-ID"), body.ExpectedVersion,
		)
		if err != nil {
			writeAdminVerificationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, decisionResponse{
			Case: toAdminCaseResponse(result.Case), Outcome: result.Outcome, Replayed: result.Replayed,
		})
	})
}

func resolveAdminPrincipal(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver) (admin.Principal, bool) {
	if resolve == nil {
		writeAdminVerificationError(w, r, admin.ErrForbidden)
		return admin.Principal{}, false
	}
	principal, err := resolve(r)
	if err != nil {
		writeAdminVerificationError(w, r, admin.ErrForbidden)
		return admin.Principal{}, false
	}
	return principal, true
}

func requireJSON(w http.ResponseWriter, r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{Code: "unsupported_media_type", Message: "Content-Type must be application/json."})
		return false
	}
	return true
}

func toAdminCaseResponse(detail admin.CaseDetail) adminCaseResponse {
	return adminCaseResponse{
		CaseID: detail.ID, SubjectRef: detail.SubjectRef, ReasonCode: detail.ReasonCode,
		SubmittedAt: detail.SubmittedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Status:      detail.Status, Version: detail.Version,
	}
}

func writeAdminVerificationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, admin.ErrForbidden):
		writeError(w, r, http.StatusForbidden, APIError{Code: "forbidden", Message: "This verification action is not permitted."})
	case errors.Is(err, admin.ErrMFARequired):
		writeError(w, r, http.StatusForbidden, APIError{Code: "mfa_required", Message: "Recent MFA is required before evidence can be accessed."})
	case errors.Is(err, admin.ErrCaseNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "verification_case_not_found", Message: "No such verification case."})
	case errors.Is(err, admin.ErrCaseClosed), errors.Is(err, admin.ErrStaleCase), errors.Is(err, admin.ErrIdempotencyConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "verification_case_conflict", Message: "The verification case changed before this action completed."})
	case errors.Is(err, admin.ErrReasonRequired), errors.Is(err, admin.ErrPurposeRequired),
		errors.Is(err, admin.ErrIdempotencyRequired), errors.Is(err, admin.ErrInvalidOutcome):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "One or more fields are invalid."})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
