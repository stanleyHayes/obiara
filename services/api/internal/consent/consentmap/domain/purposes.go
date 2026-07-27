// Package domain is the consent map: the master switchboard (Doc 08 §8;
// NFR-402 purpose-bound processing). Every purpose has a default posture
// and a member control level; identity & safety processing is required
// and cannot be disabled.
package domain

import (
	"errors"
	"time"
)

// Purpose is a consent-map row.
type Purpose string

const (
	PurposeIdentitySafety   Purpose = "identity_safety"
	PurposeMatching         Purpose = "matching_personalization"
	PurposeScamArc          Purpose = "scam_arc_monitoring"
	PurposePlayPortraits    Purpose = "play_portraits"
	PurposeProductAnalytics Purpose = "product_analytics"
)

// control is the member's control level over a purpose.
type control int

const (
	controlNone   control = iota // required; cannot be changed
	controlToggle control = 1    // member may toggle either way
	controlOptIn  control = 2    // member may only enable (default off)
	controlOptOut control = 3    // member may only disable (default on)
)

// purposeSpec declares the default and control for a purpose (Doc 08 §8).
type purposeSpec struct {
	defaultOn bool
	control   control
}

var specs = map[Purpose]purposeSpec{
	PurposeIdentitySafety:   {defaultOn: true, control: controlNone},
	PurposeMatching:         {defaultOn: false, control: controlOptIn},
	PurposeScamArc:          {defaultOn: true, control: controlOptOut},
	PurposePlayPortraits:    {defaultOn: false, control: controlOptIn},
	PurposeProductAnalytics: {defaultOn: true, control: controlOptOut},
}

var (
	ErrInvalidPurpose = errors.New("unknown consent purpose")
	ErrPurposeLocked  = errors.New("this purpose cannot be changed")
	ErrWrongDirection = errors.New("this purpose only allows one direction of change")
)

// Purposes lists every consent-map row with its default posture, for the
// member-facing switchboard.
func Purposes() map[Purpose]bool {
	out := make(map[Purpose]bool, len(specs))
	for purpose, spec := range specs {
		out[purpose] = spec.defaultOn
	}
	return out
}

// State resolves a member's effective consent for a purpose: explicit
// choice when present, otherwise the default.
func State(purpose Purpose, explicit *bool) (bool, error) {
	spec, ok := specs[purpose]
	if !ok {
		return false, ErrInvalidPurpose
	}
	if explicit != nil {
		return *explicit, nil
	}
	return spec.defaultOn, nil
}

// ValidateChange checks a member's requested change against the purpose's
// control level.
func ValidateChange(purpose Purpose, enable bool) error {
	spec, ok := specs[purpose]
	if !ok {
		return ErrInvalidPurpose
	}
	switch spec.control {
	case controlNone:
		return ErrPurposeLocked
	case controlOptIn:
		if !enable {
			return ErrWrongDirection
		}
	case controlOptOut:
		if enable {
			return ErrWrongDirection
		}
	}
	return nil
}

// Receipt is the immutable record of a consent change (plan §15: consent
// receipts).
type Receipt struct {
	ID        string
	MemberID  string
	Purpose   Purpose
	Enabled   bool
	CreatedAt time.Time
}
