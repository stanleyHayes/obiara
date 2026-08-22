// Package application runs the email channel: outbound sending with
// delivery logging, and inbound delivery-status webhooks with svix
// signature verification, timestamp tolerance and dedup (agent_plan.md
// §11: webhooks are signature-verified, timestamp/replay-checked,
// persisted, deduplicated and dead-lettered).
package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

var (
	ErrDeliveryFailed   = errors.New("email delivery failed")
	ErrSignatureInvalid = errors.New("webhook signature invalid")
	ErrTimestampStale   = errors.New("webhook timestamp outside tolerance")
	ErrSecretMissing    = errors.New("webhook secret is not configured")
	ErrDeliveryNotFound = errors.New("email delivery not found")
)

// timestampTolerance bounds webhook replay windows (agent_plan.md §11).
const timestampTolerance = 5 * time.Minute

// Sender is the outbound Resend provider port.
type Sender interface {
	Send(context.Context, domain.Message) (providerRef string, err error)
}

// DeliveryLog records and updates deliveries by provider reference.
type DeliveryLog interface {
	Record(ctx context.Context, entry DeliveryEntry) error
	UpdateStatus(ctx context.Context, providerRef, status string, at time.Time) error
}

// DeliveryEntry is one logged send.
type DeliveryEntry struct {
	To          string
	Template    domain.Template
	ProviderRef string
	Status      string
	At          time.Time
}

// EmailService sends transactional email.
type EmailService struct {
	sender Sender
	log    DeliveryLog
	now    func() time.Time
}

func NewEmailService(sender Sender, log DeliveryLog, now func() time.Time) EmailService {
	return EmailService{sender: sender, log: log, now: now}
}

// Send delivers one message and logs the outcome.
func (service EmailService) Send(ctx context.Context, to string, template domain.Template, params map[string]string) (string, error) {
	message, err := domain.NewMessage(to, template, params)
	if err != nil {
		return "", err
	}
	providerRef, sendErr := service.sender.Send(ctx, message)
	status := "sent"
	if sendErr != nil {
		status = "failed"
	}
	if err := service.log.Record(ctx, DeliveryEntry{
		To: message.To(), Template: message.Template(), ProviderRef: providerRef, Status: status, At: service.now().UTC(),
	}); err != nil && sendErr == nil {
		return "", err
	}
	if sendErr != nil {
		// The provider cause is kept in the chain. Callers still match on
		// ErrDeliveryFailed, but an operator reading the logs can tell a
		// rejected key from an unverified sender domain from a rate limit.
		return "", fmt.Errorf("%w: %w", ErrDeliveryFailed, sendErr)
	}
	return providerRef, nil
}

// WebhookService verifies and applies Resend delivery-status webhooks.
type WebhookService struct {
	log       DeliveryLog
	secret    string
	tolerance time.Duration
	now       func() time.Time
}

func NewWebhookService(log DeliveryLog, secret string, now func() time.Time) WebhookService {
	return WebhookService{log: log, secret: secret, tolerance: timestampTolerance, now: now}
}

// VerifySignature checks the svix signature over "id.timestamp.body" using
// HMAC-SHA256 with the whsec_ base64 secret. Any matching v1 signature
// counts (svix may send several during rotation).
func (service WebhookService) VerifySignature(svixID, svixTimestamp, svixSignature string, body []byte) error {
	if service.secret == "" {
		return ErrSecretMissing
	}
	if svixID == "" || svixTimestamp == "" || svixSignature == "" {
		return ErrSignatureInvalid
	}

	unix, err := strconv.ParseInt(svixTimestamp, 10, 64)
	if err != nil {
		return ErrSignatureInvalid
	}
	sentAt := time.Unix(unix, 0)
	if skew := service.now().Sub(sentAt); skew > service.tolerance || skew < -service.tolerance {
		return ErrTimestampStale
	}

	secret := strings.TrimPrefix(service.secret, "whsec_")
	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return ErrSignatureInvalid
	}
	signed := svixID + "." + svixTimestamp + "." + string(body)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signed))
	expected := mac.Sum(nil)

	for _, candidate := range strings.Split(svixSignature, " ") {
		version, signature, found := strings.Cut(candidate, ",")
		if !found || version != "v1" {
			continue
		}
		given, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			continue
		}
		if hmac.Equal(given, expected) {
			return nil
		}
	}
	return ErrSignatureInvalid
}

// ApplyStatus updates a delivery by provider reference.
func (service WebhookService) ApplyStatus(ctx context.Context, providerRef, status string) error {
	return service.log.UpdateStatus(ctx, providerRef, status, service.now().UTC())
}
