package manual

import (
	"context"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

// TestNeverApproves is the whole point of this adapter: no card number, however
// well formed, may be auto-approved without a person deciding.
func TestNeverApproves(t *testing.T) {
	provider := NewProvider()
	for _, cardNumber := range []string{
		"GHA-123456789-0", "", "GHA-000000000-0U", "GHA-000000000-0X", "anything",
	} {
		result, err := provider.Verify(context.Background(), application.ProviderRequest{
			CaseID: "vc_1", CardNumber: cardNumber,
		})
		if err != nil {
			t.Fatalf("Verify(%q) returned %v; a deliberate policy must not look like an outage", cardNumber, err)
		}
		if result.Outcome != "uncertain" {
			t.Errorf("Verify(%q) outcome = %q, want uncertain so the case reaches the review desk", cardNumber, result.Outcome)
		}
	}
}

func TestProviderRefIdentifiesTheCase(t *testing.T) {
	result, err := NewProvider().Verify(context.Background(), application.ProviderRequest{CaseID: "vc_42"})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.ProviderRef != "manual:vc_42" {
		t.Errorf("ProviderRef = %q", result.ProviderRef)
	}
	if result.Reason == "" {
		t.Error("Reason is empty; the desk cannot tell a policy referral from an outage")
	}
}
