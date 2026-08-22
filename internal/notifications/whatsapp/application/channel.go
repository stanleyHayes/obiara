// Package application runs the WhatsApp channel (E13-S05). OTP delivery is
// identity-safety class and bypasses preference caps; pod alerts pass
// through the E13-S01 decision boundary. Every send is delivery-logged
// (E13-S08 observability groundwork).
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/domain"
	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
)

var ErrDeliveryFailed = errors.New("whatsapp delivery failed")

// Sender is the outbound WhatsApp Business provider port. Production
// adapters are scored separately (agent_plan.md §11); the simulator serves
// development and tests.
type Sender interface {
	Send(context.Context, whatsappdomain.Message) (providerRef string, err error)
}

// DeliveryLog records every attempted send.
type DeliveryLog interface {
	Record(ctx context.Context, entry DeliveryEntry) error
}

// DeliveryEntry is one logged send. Params are template-bounded already.
type DeliveryEntry struct {
	To          string
	Template    whatsappdomain.Template
	ProviderRef string
	Status      string
	At          time.Time
}

// PreferenceDecider is the E13-S01 decision boundary for pod alerts.
type PreferenceDecider interface {
	Decide(ctx context.Context, memberID string, category domain.Category) (domain.Decision, error)
}

// ChannelService sends WhatsApp messages.
type ChannelService struct {
	sender  Sender
	log     DeliveryLog
	decider PreferenceDecider
	now     func() time.Time
}

func NewChannelService(sender Sender, log DeliveryLog, decider PreferenceDecider, now func() time.Time) ChannelService {
	return ChannelService{sender: sender, log: log, decider: decider, now: now}
}

// SendOtp delivers an OTP code (identity-safety class: no preference gate,
// matching the OTP-resend throttle in the identity context).
func (service ChannelService) SendOtp(ctx context.Context, phone, code string) (string, error) {
	message, err := whatsappdomain.NewOtpMessage(phone, code)
	if err != nil {
		return "", err
	}
	return service.deliver(ctx, message)
}

// SendPodAlert delivers a pod-arrival alert when the member's preferences
// allow it. The memberID drives the E13-S01 decision; the phone is the
// member's verified number (FR-701).
func (service ChannelService) SendPodAlert(ctx context.Context, memberID, phone, podRef string) (string, error) {
	message, err := whatsappdomain.NewPodAlertMessage(phone, podRef)
	if err != nil {
		return "", err
	}
	if service.decider != nil {
		decision, err := service.decider.Decide(ctx, memberID, domain.CategoryPods)
		if err != nil {
			return "", err
		}
		if !decision.Allowed {
			return "", nil // suppressed by preferences; nothing logged or sent
		}
	}
	return service.deliver(ctx, message)
}

// SendNnoboaConsent delivers the Nnoboa kin-consent invite (E13-S06). Like
// OTP it is an account-companionship message, not marketing: no preference
// gate, every attempt delivery-logged.
func (service ChannelService) SendNnoboaConsent(ctx context.Context, phone, kinName, nominationID, consentToken string) (string, error) {
	message, err := whatsappdomain.NewNnoboaConsentMessage(phone, kinName, nominationID, consentToken)
	if err != nil {
		return "", err
	}
	return service.deliver(ctx, message)
}

func (service ChannelService) deliver(ctx context.Context, message whatsappdomain.Message) (string, error) {
	providerRef, err := service.sender.Send(ctx, message)
	status := "sent"
	if err != nil {
		status = "failed"
	}
	entry := DeliveryEntry{
		To:          message.To(),
		Template:    message.Template(),
		ProviderRef: providerRef,
		Status:      status,
		At:          service.now().UTC(),
	}
	if logErr := service.log.Record(ctx, entry); logErr != nil && err == nil {
		return "", logErr
	}
	if err != nil {
		// Keep the provider cause for triage; callers match the sentinel.
		return "", fmt.Errorf("%w: %w", ErrDeliveryFailed, err)
	}
	return providerRef, nil
}
