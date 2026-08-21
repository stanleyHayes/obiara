// Package manual is the liveness provider used when no automated liveness
// vendor is contracted.
//
// The simulator it replaces answered "live" for any attempt whose artifact
// reference did not end in a scripted suffix, so in production every
// liveness check passed automatically and the anti-impersonation control
// did nothing.
//
// This provider reports every attempt as uncertain, which the liveness
// service already routes to the human review queue. It never examines
// biometric bytes.
package manual

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

// Provider routes every liveness attempt to human review.
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

// Assess always reports uncertain. It returns no error, because a deliberate
// policy must not be recorded as a provider outage.
func (Provider) Assess(_ context.Context, request application.ProviderRequest) (application.ProviderResult, error) {
	return application.ProviderResult{
		Outcome:     application.OutcomeUncertain,
		ProviderRef: "manual:" + request.AttemptID,
	}, nil
}
