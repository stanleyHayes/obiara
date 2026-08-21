package config

import (
	"strings"
	"testing"
)

// TestVerificationSimulatorsRejectedOutsideDevelopment pins the two most
// dangerous simulators in the codebase. The identity simulator answers
// "match" and the liveness simulator answers "live" for any input without a
// scripted suffix, so shipping either auto-approves fabricated evidence and
// promotes the account a tier.
func TestVerificationSimulatorsRejectedOutsideDevelopment(t *testing.T) {
	for _, variable := range []string{"IDENTITY_VERIFICATION_PROVIDER", "LIVENESS_PROVIDER"} {
		t.Run(variable, func(t *testing.T) {
			values := map[string]string{
				"IDENTITY_VERIFICATION_PROVIDER": "manual",
				"LIVENESS_PROVIDER":              "manual",
			}
			values[variable] = "simulator"

			err := validateVerification(loadVerification(envWith(values)), false)
			if err == nil {
				t.Fatalf("production accepted the %s simulator", variable)
			}
			if !strings.Contains(err.Error(), variable) {
				t.Errorf("error should name the variable, got %v", err)
			}
			if !strings.Contains(err.Error(), "manual") {
				t.Errorf("error should point operators at the manual desk, got %v", err)
			}
		})
	}
}

func TestManualProvidersAreAcceptedInProduction(t *testing.T) {
	values := map[string]string{
		"IDENTITY_VERIFICATION_PROVIDER": "manual",
		"LIVENESS_PROVIDER":              "manual",
	}
	if err := validateVerification(loadVerification(envWith(values)), false); err != nil {
		t.Fatalf("production rejected manual review: %v", err)
	}
}

func TestVerificationDefaultsToSimulatorInDevelopment(t *testing.T) {
	cfg := loadVerification(envWith(nil))
	if cfg.IdentityProvider != ProviderSimulator || cfg.LivenessProvider != ProviderSimulator {
		t.Fatalf("development defaults = %+v, want simulators", cfg)
	}
	if err := validateVerification(cfg, true); err != nil {
		t.Fatalf("development rejected its own defaults: %v", err)
	}
}

func TestUnknownVerificationProviderIsRejected(t *testing.T) {
	values := map[string]string{
		"IDENTITY_VERIFICATION_PROVIDER": "ghanacard-vendor-x",
		"LIVENESS_PROVIDER":              "manual",
	}
	if err := validateVerification(loadVerification(envWith(values)), false); err == nil {
		t.Fatal("accepted an unknown identity provider")
	}
}
