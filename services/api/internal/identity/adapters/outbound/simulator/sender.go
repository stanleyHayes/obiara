// Package simulator provides the development/test OTP sender. It delivers
// nothing: codes are retrievable only through the test hook, never logged
// (OTP codes are secrets; agent_plan.md §14). Production SMS/WhatsApp
// adapters replace this behind the same port after provider scoring.
package simulator

import (
	"context"
	"fmt"
	"sync"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// Sender is an in-memory OtpSender for local development and tests.
type Sender struct {
	mu          sync.Mutex
	lastByPhone map[string]string
}

func NewSender() *Sender {
	return &Sender{lastByPhone: make(map[string]string)}
}

func (sender *Sender) Send(_ context.Context, contact domain.Contact, code string) error {
	// The router only sends SMS contacts here; a different channel means a
	// composition mistake, and delivering a sign-in code to the wrong
	// transport is exactly what must not happen quietly.
	if contact.Channel() != domain.ChannelSMS {
		return fmt.Errorf("%s: unexpected contact channel %q", "simulator", contact.Channel())
	}
	phone := contact.Value()

	sender.mu.Lock()
	defer sender.mu.Unlock()
	sender.lastByPhone[phone] = code
	return nil
}

// LastCode returns the most recent code "sent" to a phone. Test hook only.
func (sender *Sender) LastCode(phone string) (string, bool) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	code, ok := sender.lastByPhone[phone]
	return code, ok
}
