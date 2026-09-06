package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	promotionapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
	promotiondomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
)

// AdminPromotions is the operator surface for discount codes.
type AdminPromotions interface {
	Issue(context.Context, promotionapp.IssueCommand) (promotiondomain.Promotion, error)
	Withdraw(ctx context.Context, code, commandID string) (promotiondomain.Promotion, error)
	ListByIssuer(ctx context.Context, issuerID string, limit int) ([]promotiondomain.Promotion, error)
}

// RegisterAdminPromotionRoutes exposes issuing codes for an organization.
//
// An operator surface, because codes are issued by staff on an organization's
// behalf (agent_plan.md §41). Step-up on the writes: a code is money coming
// off real revenue.
func RegisterAdminPromotionRoutes(
	mux *http.ServeMux, promotions AdminPromotions, resolve AdminPrincipalResolver,
) {
	mux.Handle("GET /v1/admin/organizations/{id}/promotions",
		adminListPromotionsHandler(promotions, resolve))
	mux.Handle("POST /v1/admin/organizations/{id}/promotions",
		adminIssuePromotionHandler(promotions, resolve))
	mux.Handle("POST /v1/admin/promotions/{code}/withdraw",
		adminWithdrawPromotionHandler(promotions, resolve))
}

type promotionResponse struct {
	Code     string `json:"code"`
	IssuerID string `json:"issuerId"`
	SKUID    string `json:"skuId"`
	Shape    string `json:"shape"`
	Amount   int64  `json:"amount"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Cap      uint32 `json:"cap"`
	// Redeemed is how many members have used it, and the only reporting there
	// is. Who used a code is never said: the redemptions are keyed.
	Redeemed  uint32 `json:"redeemed"`
	Withdrawn bool   `json:"withdrawn"`
}

type promotionListResponse struct {
	Promotions []promotionResponse `json:"promotions"`
}

type issuePromotionInput struct {
	Code  string `json:"code"`
	SKUID string `json:"skuId"`
	// Shape is "percentage" or "fixed"; Amount is a percentage or minor units
	// accordingly.
	Shape    string `json:"shape"`
	Amount   int64  `json:"amount"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	// Cap is required and cannot be zero. It is the only thing standing
	// between a leaked code and every membership being free.
	Cap uint32 `json:"cap"`
}

func projectPromotion(promotion promotiondomain.Promotion) promotionResponse {
	return promotionResponse{
		Code: promotion.Code(), IssuerID: promotion.IssuerID(), SKUID: promotion.SKUID(),
		Shape: string(promotion.Shape()), Amount: promotion.Amount(),
		StartsAt: promotion.StartsAt().UTC().Format(time.RFC3339),
		EndsAt:   promotion.EndsAt().UTC().Format(time.RFC3339),
		Cap:      promotion.Cap(), Redeemed: promotion.Redeemed(),
		Withdrawn: promotion.Withdrawn(),
	}
}

func adminListPromotionsHandler(
	promotions AdminPromotions, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if promotions == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		found, err := promotions.ListByIssuer(r.Context(), r.PathValue("id"), limit)
		if err != nil {
			writePromotionError(w, r, err)
			return
		}
		response := promotionListResponse{Promotions: make([]promotionResponse, 0, len(found))}
		for _, promotion := range found {
			response.Promotions = append(response.Promotions, projectPromotion(promotion))
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func adminIssuePromotionHandler(
	promotions AdminPromotions, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, true); !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		var body issuePromotionInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if promotions == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		startsAt, startErr := time.Parse(time.RFC3339, strings.TrimSpace(body.StartsAt))
		endsAt, endErr := time.Parse(time.RFC3339, strings.TrimSpace(body.EndsAt))
		if startErr != nil || endErr != nil {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "startsAt", Reason: "must be an RFC 3339 timestamp"}},
			})
			return
		}
		promotion, err := promotions.Issue(r.Context(), promotionapp.IssueCommand{
			CommandID: commandID, Code: body.Code, IssuerID: r.PathValue("id"),
			SKUID: body.SKUID, Shape: promotiondomain.Shape(strings.TrimSpace(body.Shape)),
			Amount: body.Amount, StartsAt: startsAt, EndsAt: endsAt, Cap: body.Cap,
		})
		if err != nil {
			writePromotionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectPromotion(promotion))
	})
}

func adminWithdrawPromotionHandler(
	promotions AdminPromotions, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, true); !ok {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		if promotions == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		promotion, err := promotions.Withdraw(r.Context(), r.PathValue("code"), commandID)
		if err != nil {
			writePromotionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectPromotion(promotion))
	})
}

func writePromotionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, promotionapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "promotion_not_found", Message: "No such code.",
		})
	case errors.Is(err, promotionapp.ErrCodeTaken):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "promotion_code_taken", Message: "That code is already in use.",
		})
	case errors.Is(err, promotionapp.ErrIssuerNotIssuing):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "organization_not_issuing",
			Message: "That organization is suspended, so no new codes can be issued for it.",
		})
	case errors.Is(err, promotiondomain.ErrNotRedeemable):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "promotion_already_withdrawn", Message: "That code was already withdrawn.",
		})
	case errors.Is(err, promotiondomain.ErrInvalidPromotion):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, promotionapp.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "promotion_conflict", Message: "That code changed. Try again.",
		})
	case errors.Is(err, promotionapp.ErrUnavailable):
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
