package manual

import (
	"context"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

// TestNeverPasses pins the anti-impersonation control: no attempt passes
// without a person reviewing it.
func TestNeverPasses(t *testing.T) {
	provider := NewProvider()
	for _, artifact := range []string{"art_1", "", "art_1:fail", "art_1:outage"} {
		result, err := provider.Assess(context.Background(), application.ProviderRequest{
			AttemptID: "live_1", VoiceArtifactRef: artifact, FaceArtifactRef: artifact,
		})
		if err != nil {
			t.Fatalf("Assess(%q) returned %v; a deliberate policy must not look like an outage", artifact, err)
		}
		if result.Outcome != application.OutcomeUncertain {
			t.Errorf("Assess(%q) outcome = %q, want uncertain", artifact, result.Outcome)
		}
	}
}

func TestProviderRefIdentifiesTheAttempt(t *testing.T) {
	result, err := NewProvider().Assess(context.Background(), application.ProviderRequest{AttemptID: "live_7"})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if result.ProviderRef != "manual:live_7" {
		t.Errorf("ProviderRef = %q", result.ProviderRef)
	}
}
