// Package failover composes OTP delivery adapters into the ladder described
// in agent_plan.md §11: SMS primary, WhatsApp fallback. The first adapter
// that accepts the message wins; the caller sees an error only when every
// rung fails.
package failover

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
)

// ErrNoSenders reports a ladder built with no rungs. Constructing one is a
// composition bug, so it surfaces at startup rather than at send time.
var ErrNoSenders = errors.New("otp failover ladder needs at least one sender")

// Observer receives per-rung outcomes so composition can meter delivery
// without this package importing a telemetry dependency. index is the
// zero-based rung. A nil Observer disables reporting.
type Observer func(ctx context.Context, index int, err error)

// Sender tries each rung in order.
type Sender struct {
	rungs   []application.OtpSender
	observe Observer
}

// NewSender builds the ladder. Rungs are tried in the given order.
func NewSender(observe Observer, rungs ...application.OtpSender) (*Sender, error) {
	present := make([]application.OtpSender, 0, len(rungs))
	for _, rung := range rungs {
		if rung != nil {
			present = append(present, rung)
		}
	}
	if len(present) == 0 {
		return nil, ErrNoSenders
	}
	return &Sender{rungs: present, observe: observe}, nil
}

// Send walks the ladder. A cancelled context stops the walk immediately
// rather than burning the remaining rungs on a request the caller has
// already abandoned.
func (sender *Sender) Send(ctx context.Context, phone, code string) error {
	var failures []error
	for index, rung := range sender.rungs {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := rung.Send(ctx, phone, code)
		if sender.observe != nil {
			sender.observe(ctx, index, err)
		}
		if err == nil {
			return nil
		}
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}
