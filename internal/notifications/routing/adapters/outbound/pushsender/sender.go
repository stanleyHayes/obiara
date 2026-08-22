// Package pushsender delivers the push rung of the routing ladder.
package pushsender

import (
	"context"
	"errors"

	pushapp "github.com/stanleyHayes/obiara/internal/notifications/push/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
)

// Deliverer is the push service surface this rung needs.
type Deliverer interface {
	Deliver(ctx context.Context, memberID, template, reference string) (string, error)
}

type Sender struct {
	push Deliverer
}

func New(push Deliverer) *Sender { return &Sender{push: push} }

func (sender *Sender) Channel() domain.Channel { return domain.ChannelPush }

// Send pushes to the member's devices. A member with no registered device is
// reported as a failure so the ladder falls through to the next channel —
// most members will not have the app installed, and the in-app inbox below
// this rung is the one that always works.
func (sender *Sender) Send(ctx context.Context, outbound application.Outbound) (string, error) {
	reference := outbound.Reference
	providerRef, err := sender.push.Deliver(ctx, outbound.MemberID, outbound.Template, reference)
	if err != nil {
		if errors.Is(err, pushapp.ErrNoDevices) {
			// Not a fault. The ladder treats any error as "try the next
			// rung", which is exactly right here.
			return "", err
		}
		return "", err
	}
	return providerRef, nil
}
