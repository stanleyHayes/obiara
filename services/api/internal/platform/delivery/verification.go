package delivery

import (
	"fmt"

	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	verificationmanual "github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/manual"
	verificationsimulator "github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/simulator"
	verificationapp "github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	livenessmanual "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/manual"
	livenesssimulator "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/simulator"
	livenessapp "github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

// IdentityProvider builds the Ghana Card verification adapter.
func IdentityProvider(cfg config.VerificationConfig) (verificationapp.VerificationProvider, error) {
	switch cfg.IdentityProvider {
	case config.ProviderManual:
		return verificationmanual.NewProvider(), nil
	case config.ProviderSimulator:
		return verificationsimulator.NewProvider(), nil
	default:
		return nil, fmt.Errorf("unknown identity verification provider %q", cfg.IdentityProvider)
	}
}

// LivenessProvider builds the liveness assessment adapter.
func LivenessProvider(cfg config.VerificationConfig) (livenessapp.Provider, error) {
	switch cfg.LivenessProvider {
	case config.ProviderManual:
		return livenessmanual.NewProvider(), nil
	case config.ProviderSimulator:
		return livenesssimulator.NewProvider(), nil
	default:
		return nil, fmt.Errorf("unknown liveness provider %q", cfg.LivenessProvider)
	}
}
