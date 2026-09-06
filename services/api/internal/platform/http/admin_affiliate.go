package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	affiliateapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/application"
	affiliatedomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

// AdminAffiliates is the operator surface for the referral scheme.
type AdminAffiliates interface {
	Register(context.Context, affiliateapp.RegisterCommand, func(context.Context, string) (bool, error)) (affiliatedomain.Affiliate, error)
	List(ctx context.Context, limit int) ([]affiliatedomain.Affiliate, error)
	Find(ctx context.Context, id string) (affiliatedomain.Affiliate, error)
}

// AdminPayouts is the desk that pays them.
type AdminPayouts interface {
	Request(ctx context.Context, affiliateID string) (affiliatedomain.Payout, error)
	Approve(ctx context.Context, payoutID, approverKey, phone, network string) (affiliatedomain.Payout, error)
	Refuse(ctx context.Context, payoutID, approverKey string) (affiliatedomain.Payout, error)
	Pending(ctx context.Context, limit int) ([]affiliatedomain.Payout, error)
}

// MemberCheck answers whether an identifier belongs to a member. Affiliates
// are outside parties, and this is the one rule that keeps them so.
type MemberCheck func(context.Context, string) (bool, error)

// RegisterAdminAffiliateRoutes exposes the scheme and its payout desk.
//
// Step-up on everything that moves money or admits somebody to the scheme.
// Approving a payout is the first outbound money this product sends, and it
// has a person's name against it.
func RegisterAdminAffiliateRoutes(
	mux *http.ServeMux, affiliates AdminAffiliates, payouts AdminPayouts,
	isMember MemberCheck, keyer OperatorKeyer, resolve AdminPrincipalResolver,
) {
	mux.Handle("GET /v1/admin/affiliates", adminListAffiliatesHandler(affiliates, resolve))
	mux.Handle("POST /v1/admin/affiliates",
		adminRegisterAffiliateHandler(affiliates, isMember, resolve))
	mux.Handle("GET /v1/admin/affiliates/payouts",
		adminPendingPayoutsHandler(payouts, resolve))
	mux.Handle("POST /v1/admin/affiliates/{id}/payouts",
		adminRequestPayoutHandler(payouts, resolve))
	mux.Handle("POST /v1/admin/affiliates/payouts/{payoutId}/decision",
		adminDecidePayoutHandler(payouts, keyer, resolve))
}

// OperatorKeyer turns an operator id into the digest a payout records.
type OperatorKeyer interface {
	Key(namespace, value string) (string, error)
}

type affiliateResponse struct {
	AffiliateID string `json:"affiliateId"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Status      string `json:"status"`
	// Conversions is how many referrals counted. Which members they were is
	// never said: the referrals are one-way digests.
	Conversions    int   `json:"conversions"`
	AccruedPesewas int64 `json:"accruedPesewas"`
	PaidPesewas    int64 `json:"paidPesewas"`
	BalancePesewas int64 `json:"balancePesewas"`
}

type payoutResponse struct {
	PayoutID        string `json:"payoutId"`
	AffiliateID     string `json:"affiliateId"`
	GrossPesewas    int64  `json:"grossPesewas"`
	WithheldPesewas int64  `json:"withheldPesewas"`
	NetPesewas      int64  `json:"netPesewas"`
	Status          string `json:"status"`
	RequestedAt     string `json:"requestedAt"`
	TransferCode    string `json:"transferCode,omitempty"`
}

func projectAffiliate(affiliate affiliatedomain.Affiliate) affiliateResponse {
	return affiliateResponse{
		AffiliateID: affiliate.ID(), Name: affiliate.Name(), Code: affiliate.Code(),
		Status: string(affiliate.Status()), Conversions: affiliate.Conversions(),
		AccruedPesewas: affiliate.AccruedPesewas(), PaidPesewas: affiliate.PaidPesewas(),
		BalancePesewas: affiliate.Balance(),
	}
}

func projectPayout(payout affiliatedomain.Payout) payoutResponse {
	return payoutResponse{
		PayoutID: payout.ID(), AffiliateID: payout.AffiliateID(),
		GrossPesewas: payout.GrossPesewas(), WithheldPesewas: payout.WithheldPesewas(),
		NetPesewas: payout.NetPesewas(), Status: string(payout.Status()),
		RequestedAt:  payout.RequestedAt().UTC().Format(time.RFC3339),
		TransferCode: payout.TransferCode(),
	}
}

func adminListAffiliatesHandler(
	affiliates AdminAffiliates, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if affiliates == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		found, err := affiliates.List(r.Context(), limit)
		if err != nil {
			writeAffiliateError(w, r, err)
			return
		}
		items := make([]affiliateResponse, 0, len(found))
		for _, affiliate := range found {
			items = append(items, projectAffiliate(affiliate))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Affiliates []affiliateResponse `json:"affiliates"`
		}{items})
	})
}

type registerAffiliateInput struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	Email      string `json:"email"`
	ReasonCode string `json:"reasonCode"`
}

func adminRegisterAffiliateHandler(
	affiliates AdminAffiliates, isMember MemberCheck, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, true); !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		var body registerAffiliateInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if affiliates == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		affiliate, err := affiliates.Register(r.Context(), affiliateapp.RegisterCommand{
			CommandID: commandID, ReasonCode: strings.TrimSpace(body.ReasonCode),
			Name: body.Name, Code: body.Code, Email: body.Email,
		}, isMember)
		if err != nil {
			writeAffiliateError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectAffiliate(affiliate))
	})
}

func adminPendingPayoutsHandler(
	payouts AdminPayouts, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if payouts == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		pending, err := payouts.Pending(r.Context(), limit)
		if err != nil {
			writeAffiliateError(w, r, err)
			return
		}
		items := make([]payoutResponse, 0, len(pending))
		for _, payout := range pending {
			items = append(items, projectPayout(payout))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Payouts []payoutResponse `json:"payouts"`
		}{items})
	})
}

func adminRequestPayoutHandler(
	payouts AdminPayouts, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, true); !ok {
			return
		}
		if payouts == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		payout, err := payouts.Request(r.Context(), r.PathValue("id"))
		if err != nil {
			writeAffiliateError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectPayout(payout))
	})
}

type payoutDecisionInput struct {
	// Approve is required rather than defaulted: a missing field must not
	// mean "send the money".
	Approve *bool `json:"approve"`
	// Phone and Network are where an approved payout goes. Read only on an
	// approval, and never stored: they are the destination for one transfer.
	Phone   string `json:"phone,omitempty"`
	Network string `json:"network,omitempty"`
}

func adminDecidePayoutHandler(
	payouts AdminPayouts, keyer OperatorKeyer, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		var body payoutDecisionInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if payouts == nil || keyer == nil || body.Approve == nil {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "approve", Reason: "is required"}},
			})
			return
		}
		approverKey, err := keyer.Key("affiliate_payout_approver", actorID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		payoutID := r.PathValue("payoutId")
		var payout affiliatedomain.Payout
		if *body.Approve {
			payout, err = payouts.Approve(
				r.Context(), payoutID, approverKey,
				strings.TrimSpace(body.Phone), strings.TrimSpace(body.Network))
		} else {
			payout, err = payouts.Refuse(r.Context(), payoutID, approverKey)
		}
		if err != nil {
			writeAffiliateError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPayout(payout))
	})
}

func writeAffiliateError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, affiliateapp.ErrMemberAffiliate):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "member_cannot_be_affiliate",
			Message: "A member cannot be an affiliate.",
		})
	case errors.Is(err, affiliateapp.ErrCodeTaken):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "affiliate_code_taken", Message: "That code is already in use.",
		})
	case errors.Is(err, affiliateapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "affiliate_not_found", Message: "No such affiliate.",
		})
	case errors.Is(err, affiliatedomain.ErrBelowMinimum):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "below_payout_minimum",
			Message: "That balance is below the payout minimum.",
		})
	case errors.Is(err, affiliatedomain.ErrWithholdingUnset):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "withholding_unset", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code:    "withholding_unset",
			Message: "No withholding rate is configured, so nothing can be paid out.",
		})
	case errors.Is(err, affiliatedomain.ErrInsufficientBalance):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "insufficient_balance", Message: "That is more than this affiliate has earned.",
		})
	case errors.Is(err, affiliatedomain.ErrPayoutTransition), errors.Is(err, affiliateapp.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "payout_conflict",
			Message: "That payout changed while you were working. Refresh and try again.",
		})
	case errors.Is(err, affiliatedomain.ErrInvalidAffiliate), errors.Is(err, affiliatedomain.ErrInvalidPayout):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, affiliateapp.ErrUnavailable):
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
