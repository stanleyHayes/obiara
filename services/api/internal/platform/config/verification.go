package config

import "fmt"

const (
	// ProviderManual routes every case to the human review desk. It is the
	// correct posture while no automated vendor is contracted: nothing is
	// auto-approved, and the feature still works through operators.
	ProviderManual Provider = "manual"
)

// VerificationConfig selects the identity and liveness provider adapters.
//
// Both default to the simulator so local development needs no vendor, and
// both reject the simulator outside development. That guard matters more
// here than anywhere else in the configuration: the identity simulator
// answers "match" and the liveness simulator answers "live" for any input
// that does not carry a scripted suffix, so shipping either to production
// turns the platform's two central safety controls into formalities that
// promote accounts on fabricated evidence.
type VerificationConfig struct {
	IdentityProvider Provider
	LivenessProvider Provider
}

func loadVerification(getenv func(string) string) VerificationConfig {
	return VerificationConfig{
		IdentityProvider: provider(getenv("IDENTITY_VERIFICATION_PROVIDER"), ProviderSimulator),
		LivenessProvider: provider(getenv("LIVENESS_PROVIDER"), ProviderSimulator),
	}
}

func validateVerification(config VerificationConfig, simulatorsAllowed bool) error {
	for _, selected := range []struct {
		variable string
		value    Provider
		effect   string
	}{
		{"IDENTITY_VERIFICATION_PROVIDER", config.IdentityProvider,
			"every Ghana Card submission would be auto-approved and promote the account a tier"},
		{"LIVENESS_PROVIDER", config.LivenessProvider,
			"every liveness attempt would pass automatically"},
	} {
		switch selected.value {
		case ProviderManual:
			// Always acceptable: cases go to the human desk.
		case ProviderSimulator:
			if !simulatorsAllowed {
				return fmt.Errorf(
					"%s may not be %q outside development: %s. Use %q to route cases to the review desk",
					selected.variable, ProviderSimulator, selected.effect, ProviderManual)
			}
		default:
			return fmt.Errorf("%s must be %q or %q, got %q",
				selected.variable, ProviderManual, ProviderSimulator, selected.value)
		}
	}
	return nil
}
