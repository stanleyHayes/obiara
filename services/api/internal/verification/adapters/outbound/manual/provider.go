// Package manual is the identity provider used when no automated Ghana Card
// vendor is contracted.
//
// It exists because the alternative was shipping the simulator, which
// answers "match" for any card number that does not end in a scripted
// suffix. In production that auto-approved every submission on fabricated
// evidence and promoted the account a tier — turning the platform's central
// identity control into a formality.
//
// This provider never approves and never rejects. It reports every case as
// uncertain, which the verification service already routes to the human
// review desk (E03-S11). Members can still submit, operators still decide,
// and no account is promoted without a person having looked at it.
package manual

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

// Reason is recorded on each queued case so the desk can tell a
// policy-driven referral from a provider outage.
const Reason = "no automated identity provider is configured; manual review required"

// Provider routes every verification to the manual desk.
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

// Verify always reports uncertain. It returns no error: an error would be
// recorded as a provider outage, and this is a deliberate policy, not a
// fault.
func (Provider) Verify(_ context.Context, request application.ProviderRequest) (application.ProviderResult, error) {
	return application.ProviderResult{
		Outcome:     "uncertain",
		ProviderRef: "manual:" + request.CaseID,
		Reason:      Reason,
	}, nil
}
