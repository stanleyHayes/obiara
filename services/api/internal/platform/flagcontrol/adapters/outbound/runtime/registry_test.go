package runtime

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flags"
	"testing"
)

func TestRegistryAppliesOnlyBoundScopeAndPreservesKernelKillPrecedence(t *testing.T) {
	config, _ := flags.NewConfiguration(nil, nil)
	kernel := flags.New(config, nil, nil)
	adapter := NewRegistry(kernel, domain.EnvironmentProduction, domain.MarketGH)
	if err := adapter.Apply(context.Background(), domain.EnvironmentProduction, domain.MarketGH, domain.RuntimeChange{Capability: domain.CapabilityPayments, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if got := kernel.Evaluate(flags.FlagPayments); !got.Enabled || got.Killed {
		t.Fatalf("%+v", got)
	}
	if err := adapter.Apply(context.Background(), domain.EnvironmentProduction, domain.MarketGH, domain.RuntimeChange{Capability: domain.CapabilityPayments, Killed: true}); err != nil {
		t.Fatal(err)
	}
	if got := kernel.Evaluate(flags.FlagPayments); got.Enabled || !got.Killed {
		t.Fatalf("%+v", got)
	}
	if err := adapter.Apply(context.Background(), domain.EnvironmentStaging, domain.MarketGH, domain.RuntimeChange{Capability: domain.CapabilityPayments, Enabled: true}); !errors.Is(err, ErrScope) {
		t.Fatalf("scope=%v", err)
	}
}
