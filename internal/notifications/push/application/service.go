// Package application runs device push registration and delivery.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

var (
	// ErrDeliveryFailed reports a provider-side rejection.
	ErrDeliveryFailed = errors.New("push delivery failed")
	// ErrNoDevices reports a member with no registered device. It is a
	// normal condition — most members will not have the app installed on
	// day one — so the router treats it as "this rung cannot serve" and
	// falls through rather than as a fault.
	ErrNoDevices = errors.New("member has no registered push device")
)

// Registry persists device registrations.
type Registry interface {
	// Register stores a device, replacing any prior registration of the
	// same token so a device that reinstalls does not accumulate rows.
	Register(context.Context, domain.Registration) error
	// TokensFor returns every live token for a member.
	TokensFor(ctx context.Context, memberID string) ([]string, error)
	// Forget removes tokens the provider reported as dead.
	Forget(ctx context.Context, tokens []string) error
	// ForgetMember removes every token for a member, for sign-out and
	// erasure.
	ForgetMember(ctx context.Context, memberID string) error
}

// Sender is the outbound push provider port.
type Sender interface {
	// Send delivers copy to the given tokens and returns the tokens the
	// provider reported as permanently dead, so the registry can prune them.
	Send(ctx context.Context, tokens []string, copy domain.Copy, reference string) (dead []string, err error)
}

// Service registers devices and delivers push notifications.
type Service struct {
	registry Registry
	sender   Sender
	now      func() time.Time
}

func NewService(registry Registry, sender Sender, now func() time.Time) Service {
	return Service{registry: registry, sender: sender, now: now}
}

// Register records a device for a member.
func (service Service) Register(ctx context.Context, memberID, token string, platform domain.Platform) error {
	registration, err := domain.NewRegistration(memberID, token, platform, service.now())
	if err != nil {
		return err
	}
	return service.registry.Register(ctx, registration)
}

// Forget removes every device for a member, used at sign-out so a shared
// handset stops receiving the previous member's notifications.
func (service Service) Forget(ctx context.Context, memberID string) error {
	return service.registry.ForgetMember(ctx, memberID)
}

// Deliver pushes one template to every device a member has registered.
//
// Tokens the provider reports as permanently dead are pruned. That matters:
// Expo rejects a whole request that contains too many stale tokens, so a
// registry that never forgets degrades into a channel that cannot deliver.
func (service Service) Deliver(ctx context.Context, memberID, template, reference string) (string, error) {
	tokens, err := service.registry.TokensFor(ctx, memberID)
	if err != nil {
		return "", err
	}
	if len(tokens) == 0 {
		return "", ErrNoDevices
	}

	copy, err := domain.CopyFor(template)
	if err != nil {
		// An unmapped template must not stop the notification; the fallback
		// says nothing specific, which is the safe default on a lock screen.
		copy = domain.FallbackCopy()
	}

	dead, sendErr := service.sender.Send(ctx, tokens, copy, reference)
	if len(dead) > 0 {
		// Pruning is best effort: failing the send because cleanup failed
		// would turn a delivered notification into a retried one.
		_ = service.registry.Forget(ctx, dead)
	}
	if sendErr != nil {
		return "", sendErr
	}
	return "push:" + memberID, nil
}
