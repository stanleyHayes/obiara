// Package simulator provides development/test SMS and push senders. They
// record sends and can be scripted to fail (fallback tests). Nothing is
// transmitted.
package simulator

import (
	"context"
	"errors"
	"sync"

	"github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
)

type Sender struct {
	channel  domain.Channel
	mu       sync.Mutex
	sent     []application.Outbound
	FailNext bool
}

func NewSMS() *Sender  { return &Sender{channel: domain.ChannelSMS} }
func NewPush() *Sender { return &Sender{channel: domain.ChannelPush} }

func (sender *Sender) Channel() domain.Channel { return sender.channel }

func (sender *Sender) Send(_ context.Context, outbound application.Outbound) (string, error) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.FailNext {
		sender.FailNext = false
		return "", errors.New("simulated provider failure")
	}
	sender.sent = append(sender.sent, outbound)
	return "sim_" + string(sender.channel) + "_" + outbound.Reference, nil
}

func (sender *Sender) Sent() []application.Outbound {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	out := make([]application.Outbound, len(sender.sent))
	copy(out, sender.sent)
	return out
}
