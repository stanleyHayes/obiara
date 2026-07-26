// Package simulator provides the development/test Resend sender.
package simulator

import (
	"context"
	"sync"

	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

type Sender struct {
	mu   sync.Mutex
	sent []domain.Message
}

func NewSender() *Sender {
	return &Sender{}
}

func (sender *Sender) Send(_ context.Context, message domain.Message) (string, error) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	sender.sent = append(sender.sent, message)
	return "sim_email_" + string(message.Template()) + "_" + message.To(), nil
}

func (sender *Sender) Sent() []domain.Message {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	out := make([]domain.Message, len(sender.sent))
	copy(out, sender.sent)
	return out
}
