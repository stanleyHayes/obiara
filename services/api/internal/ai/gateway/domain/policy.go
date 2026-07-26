package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Capability string
type DataClass string

const (
	CapabilityResonance   Capability = "resonance_explanation"
	CapabilityOkyeame     Capability = "okyeame_counsel"
	CapabilitySowScreen   Capability = "sow_screening"
	CapabilitySikaShield  Capability = "sika_shield"
	CapabilityLivenessAid Capability = "liveness_assist"

	DataConsentedText  DataClass = "consented_text"
	DataDerivedProfile DataClass = "derived_profile"
	DataTranscript     DataClass = "consented_transcript"
)

var (
	ErrDenied       = errors.New("ai gateway policy denied")
	ErrInvalid      = errors.New("invalid ai gateway input")
	ErrTooLarge     = errors.New("ai gateway input exceeds policy")
	ErrNotConsented = errors.New("ai gateway consent unavailable")
)

type Model struct {
	Vendor          string
	Name            string
	Regions         []string
	Capabilities    []Capability
	DataClasses     []DataClass
	MaxInputBytes   int
	MaxOutputTokens int
}

type Policy struct {
	Purpose     string
	Region      string
	Capability  Capability
	DataClasses []DataClass
	Model       Model
}

func NewPolicy(purpose, region string, capability Capability, classes []DataClass, model Model) (Policy, error) {
	purpose = strings.TrimSpace(purpose)
	region = strings.ToUpper(strings.TrimSpace(region))
	model.Vendor = strings.TrimSpace(model.Vendor)
	model.Name = strings.TrimSpace(model.Name)
	if purpose == "" || len(purpose) > 80 || region == "" || model.Vendor == "" || model.Name == "" {
		return Policy{}, ErrInvalid
	}
	if model.MaxInputBytes < 1 || model.MaxInputBytes > 1_000_000 ||
		model.MaxOutputTokens < 1 || model.MaxOutputTokens > 8_192 {
		return Policy{}, ErrInvalid
	}
	if !validCapability(capability) || !slices.Contains(model.Regions, region) ||
		!slices.Contains(model.Capabilities, capability) {
		return Policy{}, ErrDenied
	}
	if len(classes) == 0 || len(classes) > 3 {
		return Policy{}, ErrInvalid
	}
	seen := map[DataClass]struct{}{}
	for _, class := range classes {
		if !validDataClass(class) || !slices.Contains(model.DataClasses, class) {
			return Policy{}, fmt.Errorf("%w: data class", ErrDenied)
		}
		if _, exists := seen[class]; exists {
			return Policy{}, ErrInvalid
		}
		seen[class] = struct{}{}
	}
	return Policy{
		Purpose: purpose, Region: region, Capability: capability,
		DataClasses: slices.Clone(classes), Model: cloneModel(model),
	}, nil
}

func (p Policy) Authorize(inputBytes, outputTokens int) error {
	if inputBytes < 1 || outputTokens < 1 {
		return ErrInvalid
	}
	if inputBytes > p.Model.MaxInputBytes || outputTokens > p.Model.MaxOutputTokens {
		return ErrTooLarge
	}
	return nil
}

func cloneModel(model Model) Model {
	model.Regions = slices.Clone(model.Regions)
	model.Capabilities = slices.Clone(model.Capabilities)
	model.DataClasses = slices.Clone(model.DataClasses)
	return model
}

func validCapability(value Capability) bool {
	return slices.Contains([]Capability{
		CapabilityResonance, CapabilityOkyeame, CapabilitySowScreen,
		CapabilitySikaShield, CapabilityLivenessAid,
	}, value)
}

func validDataClass(value DataClass) bool {
	return slices.Contains([]DataClass{DataConsentedText, DataDerivedProfile, DataTranscript}, value)
}
