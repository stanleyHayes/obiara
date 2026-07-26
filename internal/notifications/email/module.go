// Package email is the composition root of the transactional email
// channel (E13-S04). The production Resend adapter replaces the simulator
// after vendor setup; webhook secrets arrive only via the environment.
package email

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
)

type Module struct {
	Email   application.EmailService
	Webhook application.WebhookService
	Sender  application.Sender
}

// NewModule builds the channel. webhookSecret is the svix whsec_ signing
// secret (empty disables the webhook with an explicit configuration error
// on each call).
func NewModule(ctx context.Context, database *mongo.Database, webhookSecret string) (Module, error) {
	deliveryLog := mongodb.NewDeliveryLog(database)
	if err := deliveryLog.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	sender := simulator.NewSender()
	return Module{
		Email:   application.NewEmailService(sender, deliveryLog, time.Now),
		Webhook: application.NewWebhookService(deliveryLog, webhookSecret, time.Now),
		Sender:  sender,
	}, nil
}
