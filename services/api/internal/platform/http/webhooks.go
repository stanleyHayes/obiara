package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
)

// ResendWebhook is the inbound port for Resend delivery webhooks.
type ResendWebhook interface {
	VerifySignature(svixID, svixTimestamp, svixSignature string, body []byte) error
	ApplyStatus(ctx context.Context, providerRef, status string) error
}

const maxWebhookBodyBytes = 64 << 10

// RegisterResendWebhookRoute adds the signed Resend webhook ingress
// (E13-S04; agent_plan.md §11 webhook standards). The route is
// unauthenticated but HMAC-signed; replays are deduplicated.
func RegisterResendWebhookRoute(mux *http.ServeMux, webhook ResendWebhook, inboxStore *inbox.Store) {
	mux.Handle("POST /webhooks/resend", resendWebhookHandler(webhook, inboxStore))
}

type resendPayload struct {
	Type string `json:"type"`
	Data struct {
		EmailID string `json:"email_id"`
	} `json:"data"`
}

// statusByEvent maps Resend event types to delivery statuses.
var statusByEvent = map[string]string{
	"email.sent":       "sent",
	"email.delivered":  "delivered",
	"email.bounced":    "bounced",
	"email.complained": "complained",
	"email.failed":     "failed",
}

func resendWebhookHandler(webhook ResendWebhook, inboxStore *inbox.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes))
		if err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		svixID := r.Header.Get("svix-id")
		if err := webhook.VerifySignature(svixID, r.Header.Get("svix-timestamp"), r.Header.Get("svix-signature"), body); err != nil {
			status := http.StatusUnauthorized
			code := "signature_invalid"
			if errors.Is(err, application.ErrTimestampStale) {
				code = "timestamp_stale"
			}
			writeError(w, r, status, APIError{
				Code:    code,
				Message: "The webhook signature could not be verified.",
			})
			return
		}

		if inboxStore != nil {
			seen, err := inboxStore.AlreadyProcessed(r.Context(), "resend.webhook", svixID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
				return
			}
			if seen {
				writeSuccess(w, r, http.StatusOK, map[string]string{"status": "duplicate"})
				return
			}
		}

		var payload resendPayload
		if err := json.Unmarshal(body, &payload); err != nil || payload.Data.EmailID == "" {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		status, ok := statusByEvent[payload.Type]
		if !ok {
			// Unknown event types are acknowledged without action (forward
			// compatible with new Resend events).
			writeSuccess(w, r, http.StatusOK, map[string]string{"status": "ignored"})
			return
		}
		if err := webhook.ApplyStatus(r.Context(), payload.Data.EmailID, status); err != nil {
			if errors.Is(err, application.ErrDeliveryNotFound) {
				writeError(w, r, http.StatusNotFound, APIError{
					Code:    "delivery_not_found",
					Message: "No delivery with that reference.",
				})
				return
			}
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]string{"status": "applied"})
	})
}
