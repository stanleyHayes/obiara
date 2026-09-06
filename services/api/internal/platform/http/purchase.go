package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	momoapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
)

// Purchases starts a membership purchase and settles the provider's callback.
type Purchases interface {
	Start(context.Context, purchase.StartCommand) (purchase.Started, error)
	Settle(context.Context, momoapplication.Callback) error
}

// RegisterPurchaseRoutes exposes buying a membership.
//
// Registered only when a collection provider is configured. Without one there
// is no way to take money, and a purchase route that always failed would be
// worse than a surface that is plainly absent — the same rule the Voice of
// Introduction follows about object storage.
//
// The callback carries no session. It is authenticated by the HMAC the
// payment context verifies over the whole payload, because the caller is a
// provider and not a member.
func RegisterPurchaseRoutes(
	mux *http.ServeMux, purchases Purchases, sessions SessionAuthenticator,
) {
	mux.Handle("POST /v1/membership/purchases", startPurchaseHandler(purchases, sessions))
	mux.Handle("POST /v1/payments/momo/callback", momoCallbackHandler(purchases))
}

type startPurchaseRequest struct {
	SKUID      string `json:"skuId"`
	SKUVersion uint64 `json:"skuVersion"`
	// Phone is the number the provider prompts. It is keyed before it reaches
	// storage and is never written down as given.
	Phone string `json:"phone"`
	// Code is optional. One that does not apply is not an error: the member
	// came to buy a membership and a typo should not stop them.
	Code string `json:"code,omitempty"`
}

type startPurchaseResponse struct {
	PurchaseID string `json:"purchaseId"`
	Status     string `json:"status"`
	// AmountPesewas is what the member is about to be asked for, echoed back
	// so a client shows the real number rather than one it assumed.
	AmountPesewas uint64 `json:"amountPesewas"`
	// DiscountPesewas and Code say what a code took off, so a member sees the
	// discount rather than inferring it from a smaller number.
	DiscountPesewas uint64 `json:"discountPesewas,omitempty"`
	Code            string `json:"code,omitempty"`
}

func startPurchaseHandler(purchases Purchases, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		// The buyer is the session. Nobody buys a membership for somebody
		// else here: the pass is granted to whoever paid.
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if commandID == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
			})
			return
		}
		var body startPurchaseRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if purchases == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		started, err := purchases.Start(r.Context(), purchase.StartCommand{
			CommandID: commandID, MemberID: memberID,
			SKUID: body.SKUID, SKUVersion: body.SKUVersion, Phone: body.Phone,
			Code: strings.ToUpper(strings.TrimSpace(body.Code)),
		})
		if err != nil {
			writePurchaseError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusAccepted, startPurchaseResponse{
			PurchaseID:      started.IntentID,
			Status:          started.Status,
			AmountPesewas:   started.AmountPesewas,
			DiscountPesewas: started.DiscountPesewas,
			Code:            started.Code,
		})
	})
}

type momoCallbackRequest struct {
	CallbackID  string `json:"callbackId"`
	IntentID    string `json:"intentId"`
	ProviderRef string `json:"providerRef"`
	Success     bool   `json:"success"`
	OccurredAt  int64  `json:"occurredAt"`
	Signature   string `json:"signature"`
}

// momoCallbackHandler settles a payment on the provider's word.
//
// It answers 204 for anything it could act on, including a callback it has
// already seen. Providers retry, and an error on a retry makes them retry
// harder — the payment context is idempotent by callback id, so saying "done"
// to the second one is both true and the only way to make retries stop.
func momoCallbackHandler(purchases Purchases) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		var body momoCallbackRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if purchases == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		err := purchases.Settle(r.Context(), momoapplication.Callback{
			CallbackID:   strings.TrimSpace(body.CallbackID),
			IntentID:     strings.TrimSpace(body.IntentID),
			ProviderRef:  strings.TrimSpace(body.ProviderRef),
			Success:      body.Success,
			OccurredUnix: body.OccurredAt,
			Signature:    strings.TrimSpace(body.Signature),
		})
		switch {
		case err == nil, errors.Is(err, purchase.ErrAlreadySettled):
			w.WriteHeader(http.StatusNoContent)
		default:
			// Deliberately says nothing about why. The caller is a provider
			// and a signature failure is the same shape as an unknown intent:
			// telling them apart would let anybody probe which intents exist.
			logServerError(r.Context(), r, http.StatusBadRequest, "callback_rejected", err)
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "callback_rejected", Message: "That callback could not be accepted.",
			})
		}
	})
}

func writePurchaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, purchase.ErrNotPurchasable):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "not_purchasable",
			Message: "That is not something you can buy right now.",
		})
	case errors.Is(err, purchase.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code:    "payment_unavailable",
			Message: "We could not reach mobile money. Please try again shortly.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
