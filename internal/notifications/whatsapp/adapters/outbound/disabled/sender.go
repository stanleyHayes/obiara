// Package disabled is the explicit "this channel is not in service" WhatsApp
// adapter.
//
// It exists so that turning a channel off is not the same as the simulator.
// The simulator reports every send as delivered, which is correct for tests
// and catastrophic in production: a message the caller believes was sent but
// which reached nobody. This adapter fails every send loudly instead, so the
// delivery log records the failure, the caller's retry or dead-letter path
// runs, and operators can see that the channel is dark.
//
// Use it when WhatsApp Business is not yet provisioned but the rest of the
// platform must ship.
package disabled

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
)

// ErrChannelDisabled reports that WhatsApp is deliberately out of service.
var ErrChannelDisabled = errors.New("whatsapp channel is disabled: set WHATSAPP_PROVIDER=meta and provision the Cloud API to enable it")

// Sender rejects every message.
type Sender struct{}

func NewSender() *Sender { return &Sender{} }

func (Sender) Send(_ context.Context, _ domain.Message) (string, error) {
	return "", ErrChannelDisabled
}
