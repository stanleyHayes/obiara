package domain

import (
	"errors"
	"strings"
	"testing"
)

func testModel() Model {
	return Model{
		Vendor: "approved-vendor", Name: "approved-model", Regions: []string{"GH"},
		Capabilities:  []Capability{CapabilityResonance},
		DataClasses:   []DataClass{DataConsentedText},
		MaxInputBytes: 4096, MaxOutputTokens: 512,
	}
}

func TestPolicyFailsClosedOutsideAllowlist(t *testing.T) {
	t.Parallel()
	model := testModel()
	if _, err := NewPolicy("explain resonance", "US", CapabilityResonance, []DataClass{DataConsentedText}, model); !errors.Is(err, ErrDenied) {
		t.Fatalf("region denial = %v", err)
	}
	if _, err := NewPolicy("counsel", "GH", CapabilityOkyeame, []DataClass{DataConsentedText}, model); !errors.Is(err, ErrDenied) {
		t.Fatalf("capability denial = %v", err)
	}
	if _, err := NewPolicy("explain", "GH", CapabilityResonance, []DataClass{DataDerivedProfile}, model); !errors.Is(err, ErrDenied) {
		t.Fatalf("class denial = %v", err)
	}
}

func TestPolicyBoundsInputAndOutput(t *testing.T) {
	t.Parallel()
	policy, err := NewPolicy("explain resonance", "GH", CapabilityResonance, []DataClass{DataConsentedText}, testModel())
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.Authorize(4096, 512); err != nil {
		t.Fatal(err)
	}
	if err := policy.Authorize(4097, 1); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("input bound = %v", err)
	}
	if err := policy.Authorize(1, 513); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("output bound = %v", err)
	}
}

func TestPolicyCopiesCallerSlices(t *testing.T) {
	t.Parallel()
	model := testModel()
	classes := []DataClass{DataConsentedText}
	policy, err := NewPolicy("explain", "GH", CapabilityResonance, classes, model)
	if err != nil {
		t.Fatal(err)
	}
	classes[0] = DataTranscript
	model.Regions[0] = "US"
	if policy.DataClasses[0] != DataConsentedText || policy.Model.Regions[0] != "GH" {
		t.Fatal("policy retained caller-owned slices")
	}
}

func FuzzPolicyNeverAllowsUnlistedRoute(f *testing.F) {
	f.Add("GH", string(CapabilityResonance), string(DataConsentedText))
	f.Add("US", string(CapabilityOkyeame), string(DataDerivedProfile))
	f.Fuzz(func(t *testing.T, region, capability, class string) {
		model := testModel()
		_, err := NewPolicy(
			"explain", region, Capability(capability), []DataClass{DataClass(class)}, model,
		)
		allowed := strings.EqualFold(strings.TrimSpace(region), "GH") &&
			Capability(capability) == CapabilityResonance &&
			DataClass(class) == DataConsentedText
		if allowed && err != nil {
			t.Fatalf("allowlisted route rejected: %v", err)
		}
		if !allowed && err == nil {
			t.Fatalf("unlisted route allowed: %q %q %q", region, capability, class)
		}
	})
}
