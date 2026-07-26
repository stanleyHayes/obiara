// Package simulator provides the development/test WhatsApp sender. It
// records sends in memory; nothing is transmitted.
package simulator

import (
	"context"
	"sync"

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
)

type Sender struct {
	mu     sync.Mutex
	sent   []domain.Message
	nextID int
}

func NewSender() *Sender {
	return &Sender{}
}

func (sender *Sender) Send(_ context.Context, message domain.Message) (string, error) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	sender.nextID++
	sender.sent = append(sender.sent, message)
	return "sim_wa_" + message.To(), nil
}

// Sent returns every recorded message, for test assertions.
func (sender *Sender) Sent() []domain.Message {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	out := make([]domain.Message, len(sender.sent))
	copy(out, sender.sent)
	return out
}
