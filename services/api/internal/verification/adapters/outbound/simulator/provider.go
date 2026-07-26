// Package simulator provides the development/test identity provider. It
// produces deterministic, scripted outcomes behind the provider port until
// a scored Ghana Card vendor is selected (agent_plan.md §23: port-first
// adapters with simulators and a dual shortlist).
package simulator

import (
	"context"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

// Provider is a deterministic VerificationProvider for dev and tests.
// Outcome selection is data-driven so tests can script every branch:
// card numbers ending in "U" simulate provider outage, ending in "?" an
// uncertain result, ending in "X" a mismatch; anything else matches.
type Provider struct {
	requests []application.ProviderRequest
}

func NewProvider() *Provider {
	return &Provider{}
}

func (provider *Provider) Verify(_ context.Context, request application.ProviderRequest) (application.ProviderResult, error) {
	provider.requests = append(provider.requests, request)
	switch {
	case strings.HasSuffix(request.CardNumber, "U"):
		return application.ProviderResult{}, context.DeadlineExceeded
	case strings.HasSuffix(request.CardNumber, "?"):
		return application.ProviderResult{Outcome: "uncertain", ProviderRef: "sim:" + request.CaseID, Reason: "document unreadable"}, nil
	case strings.HasSuffix(request.CardNumber, "X"):
		return application.ProviderResult{Outcome: "mismatch", ProviderRef: "sim:" + request.CaseID, Reason: "issuer records do not match"}, nil
	default:
		return application.ProviderResult{Outcome: "match", ProviderRef: "sim:" + request.CaseID, Reason: "issuer match"}, nil
	}
}

// Requests returns every submitted request, for test assertions.
func (provider *Provider) Requests() []application.ProviderRequest {
	return provider.requests
}
