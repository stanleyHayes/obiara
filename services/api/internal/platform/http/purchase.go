package apihttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/adapters/outbound/paystack"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
)

// Purchases starts a membership purchase and settles a verified outcome.
type Purchases interface {
	Start(context.Context, purchase.StartCommand) (purchase.Started, error)
	Settle(context.Context, purchase.Outcome) error
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
	mux *http.ServeMux, purchases Purchases, sessions SessionAuthenticator, webhookSecret string,
) {
	mux.Handle("POST /v1/membership/purchases", startPurchaseHandler(purchases, sessions))
	mux.Handle("POST /v1/payments/paystack/webhook", paystackWebhookHandler(purchases, webhookSecret))
}

type startPurchaseRequest struct {
	SKUID      string `json:"skuId"`
	SKUVersion uint64 `json:"skuVersion"`
	// Phone is the number the provider prompts. It is keyed before it reaches
	// storage and is never written down as given.
	Phone string `json:"phone"`
	// Network is which mobile money provider the number is on. The processor
	// needs it and cannot reliably infer it from the number.
	Network string `json:"network"`
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
			Network: strings.TrimSpace(body.Network),
			Code:    strings.ToUpper(strings.TrimSpace(body.Code)),
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

// paystackWebhookHandler settles a payment on the processor's word.
//
// It reads the raw bytes and verifies the HMAC over exactly those bytes before
// anything parses them. That ordering is the whole security of this endpoint:
// the signature covers what was sent, so decoding first and re-serialising
// would hash something the processor never produced. It is also why this does
// not use decodeJSON — that consumes the body and rejects unknown fields, and
// a webhook must tolerate a processor adding one.
//
// It answers 200 for anything it could act on, including a webhook it has
// already seen. Processors retry, the payment context is idempotent by
// callback id, and an error on a retry only makes them retry harder.
func paystackWebhookHandler(purchases Purchases, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if purchases == nil || strings.TrimSpace(secret) == "" {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBytes))
		if err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "callback_rejected", Message: "That callback could not be accepted.",
			})
			return
		}
		event, err := paystack.VerifyWebhook(secret, body, r.Header.Get(paystack.SignatureHeader))
		if err != nil {
			// Says nothing about why. A bad signature and an unreadable body
			// answer identically, or this would tell somebody probing which
			// of the two they had got wrong.
			logServerError(r.Context(), r, http.StatusBadRequest, "callback_rejected", err)
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "callback_rejected", Message: "That callback could not be accepted.",
			})
			return
		}
		// Only the event that says money arrived is acted on. Paystack sends
		// many others — transfers, invoices, subscriptions — and answering
		// them with 200 without acting is correct: they are not this
		// endpoint's business, and an error would make Paystack retry them
		// forever.
		if event.Name != "charge.success" {
			w.WriteHeader(http.StatusOK)
			return
		}
		settleErr := purchases.Settle(r.Context(), purchase.Outcome{
			// The event carries no id of its own, so the reference and the
			// status together name this settlement. A retry of the same
			// outcome therefore carries the same callback id and settles once.
			CallbackID:    "paystack:" + event.Reference + ":" + event.Status,
			Reference:     event.Reference,
			Success:       event.Succeeded(),
			AmountPesewas: event.AmountPesewas,
			Currency:      event.Currency,
		})
		switch {
		case settleErr == nil, errors.Is(settleErr, purchase.ErrAlreadySettled):
			w.WriteHeader(http.StatusOK)
		case errors.Is(settleErr, purchase.ErrWrongAmount):
			// Authentic and wrong. Answered 200 so the processor stops
			// retrying something retrying will never fix, and logged loudly
			// because somebody paid an amount nobody asked for.
			logServerError(r.Context(), r, http.StatusOK, "settlement_amount_mismatch", settleErr)
			w.WriteHeader(http.StatusOK)
		default:
			// Something on this side failed. A non-2xx asks the processor to
			// try again, which is exactly what should happen.
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "settlement_failed", settleErr)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "settlement_failed", Message: "That callback could not be settled.",
			})
		}
	})
}

// maxWebhookBytes bounds what is read before anything is verified. Hashing an
// unbounded body is a way to make the server read forever.
const maxWebhookBytes = 1 << 20

func writePurchaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, purchase.ErrSponsorshipUnavailable):
		// The sponsorship is refused, not the purchase. The member can still
		// buy their own membership, and the message says so rather than
		// reading as their account being at fault.
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "sponsorship_unavailable",
			Message: "Your organisation's sponsorship is not available right now. You can still buy a membership.",
		})
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
