package runtime

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flags"
)

var ErrScope = errors.New("runtime flag scope mismatch")

type Registry struct {
	registry    *flags.Registry
	environment domain.Environment
	market      domain.Market
}

func NewRegistry(r *flags.Registry, environment domain.Environment, market domain.Market) Registry {
	return Registry{r, environment, market}
}
func (r Registry) Apply(_ context.Context, environment domain.Environment, market domain.Market, change domain.RuntimeChange) error {
	if r.registry == nil || environment != r.environment || market != r.market {
		return ErrScope
	}
	enabled, killed := change.Enabled, change.Killed
	return r.registry.Apply(flags.Change{Flag: flags.Flag(change.Capability), Enabled: &enabled, Killed: &killed})
}
