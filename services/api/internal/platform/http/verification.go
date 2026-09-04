package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

// Verification is the inbound port for identity verification (E03-S03).
type Verification interface {
	SubmitGhanaCard(ctx context.Context, accountID, cardNumber string, dateOfBirth time.Time) (domain.VerificationCase, error)
	SubmitDocuments(context.Context, application.SubmitDocumentsRequest) (application.SubmitDocumentsResult, error)
}

// RegisterVerificationRoutes adds the verification baseline routes to mux.
func RegisterVerificationRoutes(mux *http.ServeMux, verification Verification, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/verifications/ghana-card", submitGhanaCardHandler(verification, sessions))
	mux.Handle("POST /v1/verifications/ghana-card/documents", submitCardDocumentsHandler(verification, sessions))
}

type cardDocumentsRequest struct {
	CardNumber     string `json:"cardNumber"`
	DateOfBirth    string `json:"dateOfBirth"`
	FrontMediaType string `json:"frontMediaType"`
	FrontBase64    string `json:"frontBase64"`
	BackMediaType  string `json:"backMediaType"`
	BackBase64     string `json:"backBase64"`
}

// maxCardDocumentsRequestBytes bounds the whole envelope: two images at their
// individual cap, plus base64 expansion and the small JSON around them.
const maxCardDocumentsRequestBytes = 12 << 20

// submitCardDocumentsHandler takes both sides of a card for human review.
//
// This replaced the issuer lookup inside signing up. That call was a third
// party, and while it was unreachable nobody could finish creating an account
// at all — an outage at a vendor closed the front door. A member now uploads
// the card after they are already in, and the result decides a badge.
func submitCardDocumentsHandler(verification Verification, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || sessions == nil || verification == nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(r.Context(), token)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		if mediaType, _, mediaErr := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaErr != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxCardDocumentsRequestBytes)

		var body cardDocumentsRequest
		if decodeErr := decodeJSON(w, r, &body); decodeErr != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		dateOfBirth, dobErr := time.Parse("2006-01-02", strings.TrimSpace(body.DateOfBirth))
		if dobErr != nil {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "dateOfBirth", Reason: "must be YYYY-MM-DD"}},
			})
			return
		}

		result, err := verification.SubmitDocuments(r.Context(), application.SubmitDocumentsRequest{
			AccountID: session.MemberID(), CardNumber: body.CardNumber, DateOfBirth: dateOfBirth,
			FrontMediaType: body.FrontMediaType, FrontBase64: body.FrontBase64,
			BackMediaType: body.BackMediaType, BackBase64: body.BackBase64,
		})
		if errors.Is(err, application.ErrInvalidDocument) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "Both sides of the card are required as JPEG, PNG or WebP images under 4MB.",
			})
			return
		}
		if errors.Is(err, application.ErrIdentityAlreadyVerified) {
			writeError(w, r, http.StatusConflict, APIError{
				Code:    "identity_already_verified",
				Message: "This identity is already verified on another account.",
			})
			return
		}
		if err != nil {
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "document_store_unavailable", err)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "document_store_unavailable", Message: "Secure storage is unavailable. Please try again shortly.",
			})
			return
		}
		writeSuccess(w, r, http.StatusAccepted, verificationCaseResponse{
			CaseID: result.CaseID, Status: result.Status,
		})
	})
}

type ghanaCardRequest struct {
	CardNumber  string `json:"cardNumber"`
	DateOfBirth string `json:"dateOfBirth"`
}

type verificationCaseResponse struct {
	CaseID string `json:"caseId"`
	Status string `json:"status"`
}

func submitGhanaCardHandler(verification Verification, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || sessions == nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(r.Context(), token)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body ghanaCardRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.CardNumber = strings.TrimSpace(body.CardNumber)
		dateOfBirth, dobErr := time.Parse("2006-01-02", strings.TrimSpace(body.DateOfBirth))

		var details []FieldError
		if body.CardNumber == "" || len(body.CardNumber) > 32 {
			details = append(details, FieldError{Field: "cardNumber", Reason: "must be 1-32 characters"})
		}
		if dobErr != nil {
			details = append(details, FieldError{Field: "dateOfBirth", Reason: "must be YYYY-MM-DD"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		verificationCase, err := verification.SubmitGhanaCard(r.Context(), session.MemberID(), body.CardNumber, dateOfBirth)
		if err != nil {
			writeVerificationError(w, r, err, verificationCase)
			return
		}

		status := http.StatusCreated
		if verificationCase.Status() == domain.StatusQueuedManual {
			status = http.StatusAccepted
		}
		writeSuccess(w, r, status, verificationCaseResponse{
			CaseID: verificationCase.ID(),
			Status: string(verificationCase.Status()),
		})
	})
}

func writeVerificationError(w http.ResponseWriter, r *http.Request, err error, verificationCase domain.VerificationCase) {
	switch {
	case errors.Is(err, application.ErrIdentityAlreadyVerified):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "identity_already_verified",
			Message: "This identity is already verified on another account. Sign in to that account, or contact support if you believe this is wrong.",
		})
	case errors.Is(err, application.ErrProviderRejected):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "verification_rejected",
			Message: "The document could not be verified with the issuer.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
