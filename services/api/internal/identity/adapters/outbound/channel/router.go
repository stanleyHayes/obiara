// Package channel routes an OTP to the transport its contact names.
//
// It exists because the channel is part of the identity, not a delivery
// preference: a member who verified an email address must receive their
// code by email, and routing it anywhere else would send a sign-in secret
// to an address they never proved they control. Keeping that decision in
// one adapter means the registration service never learns what a provider
// is, and adding a channel later is a map entry rather than a branch spread
// through the application layer.
package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var (
	// ErrNoRoutes reports a router built with no channels at all.
	// Constructing one is a composition bug, so it surfaces at startup.
	ErrNoRoutes = errors.New("otp channel router needs at least one route")
	// ErrChannelUnavailable is the application's sentinel, re-exported so
	// this adapter does not invent a second one for the same condition.
	ErrChannelUnavailable = application.ErrChannelUnavailable
)

// Router dispatches by the contact's channel.
type Router struct {
	routes map[domain.Channel]application.OtpSender
}

// NewRouter builds the dispatch table, ignoring nil senders so a deployment
// can configure only the channels it actually has credentials for.
func NewRouter(routes map[domain.Channel]application.OtpSender) (*Router, error) {
	present := make(map[domain.Channel]application.OtpSender, len(routes))
	for channel, sender := range routes {
		if sender != nil {
			present[channel] = sender
		}
	}
	if len(present) == 0 {
		return nil, ErrNoRoutes
	}
	return &Router{routes: present}, nil
}

// Supports reports whether this deployment can deliver on a channel, so a
// transport layer can refuse a request up front rather than minting a code
// nobody will ever receive.
func (router *Router) Supports(channel domain.Channel) bool {
	_, ok := router.routes[channel]
	return ok
}

// Channels lists the routable channels, for boot logging and for telling a
// caller what it may ask for.
func (router *Router) Channels() []domain.Channel {
	// Fixed order rather than map order, so boot logs are comparable
	// between restarts.
	ordered := make([]domain.Channel, 0, len(router.routes))
	for _, channel := range []domain.Channel{domain.ChannelSMS, domain.ChannelEmail} {
		if _, ok := router.routes[channel]; ok {
			ordered = append(ordered, channel)
		}
	}
	return ordered
}

// Send delivers the code on the contact's own channel.
func (router *Router) Send(ctx context.Context, contact domain.Contact, code string) error {
	sender, ok := router.routes[contact.Channel()]
	if !ok {
		return fmt.Errorf("%w: %s", ErrChannelUnavailable, contact.Channel())
	}
	return sender.Send(ctx, contact, code)
}
