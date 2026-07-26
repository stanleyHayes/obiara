// Package application routes notifications across channels with fallback
// (E13-S03). The E13-S01 decision boundary runs once before the ladder;
// safety and OTP bypass it. The first channel that accepts wins; a full
// ladder failure is a hard error for the caller's retry/dead-letter path.
package application

import (
	"context"
	"errors"
	"time"

	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
)

var ErrAllChannelsFailed = errors.New("all channels in the ladder failed")

// Outbound is one routed notification. Reference is opaque; template keys
// render through the localization registry downstream.
type Outbound struct {
	MemberID  string
	Phone     string
	DeviceRef string
	Purpose   domain.Purpose
	Template  string
	Reference string
}

// Result reports which channel delivered.
type Result struct {
	Channel     domain.Channel
	ProviderRef string
}

// ChannelSender delivers via one channel.
type ChannelSender interface {
	Channel() domain.Channel
	Send(context.Context, Outbound) (providerRef string, err error)
}

// PreferenceDecider is the E13-S01 decision boundary.
type PreferenceDecider interface {
	Decide(ctx context.Context, memberID string, category notificationdomain.Category) (notificationdomain.Decision, error)
}

// Router delivers through the ladder.
type Router struct {
	senders map[domain.Channel]ChannelSender
	decider PreferenceDecider
	now     func() time.Time
}

func NewRouter(senders []ChannelSender, decider PreferenceDecider, now func() time.Time) Router {
	byChannel := make(map[domain.Channel]ChannelSender, len(senders))
	for _, sender := range senders {
		byChannel[sender.Channel()] = sender
	}
	return Router{senders: byChannel, decider: decider, now: now}
}

// Deliver tries the purpose's ladder in order.
func (router Router) Deliver(ctx context.Context, outbound Outbound) (Result, error) {
	if !domain.BypassesPreferences(outbound.Purpose) && router.decider != nil {
		decision, err := router.decider.Decide(ctx, outbound.MemberID, domain.CategoryFor(outbound.Purpose))
		if err != nil {
			return Result{}, err
		}
		if !decision.Allowed {
			return Result{}, nil // suppressed by preferences
		}
	}

	for _, channel := range domain.LadderFor(outbound.Purpose) {
		sender, ok := router.senders[channel]
		if !ok {
			continue
		}
		providerRef, err := sender.Send(ctx, outbound)
		if err == nil {
			return Result{Channel: channel, ProviderRef: providerRef}, nil
		}
	}
	return Result{}, ErrAllChannelsFailed
}
