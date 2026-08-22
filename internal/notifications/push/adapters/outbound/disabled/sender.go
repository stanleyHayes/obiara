// Package disabled is the explicit "push is not in service" adapter.
//
// It is distinct from a simulator: a simulator reports every notification as
// delivered, which in production means a channel that silently reaches
// nobody. This fails every send, so the routing ladder falls through to a
// channel that can actually deliver.
package disabled

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

// ErrChannelDisabled reports that push is deliberately out of service.
var ErrChannelDisabled = errors.New("push channel is disabled: set PUSH_PROVIDER=expo to enable it")

type Sender struct{}

func NewSender() *Sender { return &Sender{} }

func (Sender) Send(context.Context, []string, domain.Copy, string) ([]string, error) {
	return nil, ErrChannelDisabled
}
