// Package email is the composition root of the transactional email
// channel (E13-S04). The production Resend adapter replaces the simulator
// after vendor setup; webhook secrets arrive only via the environment.
package email

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
)

type Module struct {
	Email   application.EmailService
	Webhook application.WebhookService
	Sender  application.Sender
}

// ErrSenderRequired reports a module built without an email transport.
// Admin MFA codes ride this channel, so a nil sender locks every operator
// out of the console; it fails at startup rather than at first sign-in.
var ErrSenderRequired = errors.New("email module requires a sender")

// NewModule builds the channel. sender is the provider adapter chosen by the
// composition root from EMAIL_PROVIDER and must not be nil. webhookSecret is
// the svix whsec_ signing secret (empty disables the webhook with an
// explicit configuration error on each call).
func NewModule(ctx context.Context, database *mongo.Database, sender application.Sender, webhookSecret string) (Module, error) {
	if sender == nil {
		return Module{}, ErrSenderRequired
	}
	deliveryLog := mongodb.NewDeliveryLog(database)
	if err := deliveryLog.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Email:   application.NewEmailService(sender, deliveryLog, time.Now),
		Webhook: application.NewWebhookService(deliveryLog, webhookSecret, time.Now),
		Sender:  sender,
	}, nil
}
