// Package simulator provides the development/test OTP sender. It delivers
// nothing: codes are retrievable only through the test hook, never logged
// (OTP codes are secrets; agent_plan.md §14). Production SMS/WhatsApp
// adapters replace this behind the same port after provider scoring.
package simulator

import (
	"context"
	"sync"
)

// Sender is an in-memory OtpSender for local development and tests.
type Sender struct {
	mu          sync.Mutex
	lastByPhone map[string]string
}

func NewSender() *Sender {
	return &Sender{lastByPhone: make(map[string]string)}
}

func (sender *Sender) Send(_ context.Context, phone, code string) error {
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
