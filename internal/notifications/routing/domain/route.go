// Package domain defines the notification channel ladder (E13-S03): which
// channels a category tries, in fallback order. Delivery always respects
// the E13-S01 preference boundary first; safety and OTP bypass it (member
// physical safety and identity processing outrank preferences).
package domain

import (
	"github.com/stanleyHayes/obiara/internal/notifications/domain"
)

// Channel is a delivery channel.
type Channel string

const (
	ChannelPush     Channel = "push"
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelInApp    Channel = "in_app"
	ChannelEmail    Channel = "email"
)

// Purpose extends the preference categories with the OTP flow, which is
// identity-safety class rather than a preference category.
type Purpose string

const (
	PurposeRitual Purpose = "ritual"
	PurposePods   Purpose = "pods"
	PurposeRooms  Purpose = "rooms"
	PurposeSafety Purpose = "safety"
	PurposeOtp    Purpose = "otp"
)

// LadderFor returns the ordered fallback channels for a purpose.
// OTP is SMS-primary with WhatsApp fallback (agent_plan.md §11).
func LadderFor(purpose Purpose) []Channel {
	switch purpose {
	case PurposeOtp:
		return []Channel{ChannelSMS, ChannelWhatsApp}
	case PurposeSafety:
		return []Channel{ChannelPush, ChannelSMS, ChannelWhatsApp, ChannelInApp}
	case PurposePods:
		return []Channel{ChannelPush, ChannelWhatsApp, ChannelInApp}
	case PurposeRitual, PurposeRooms:
		return []Channel{ChannelPush, ChannelInApp}
	default:
		return []Channel{ChannelInApp}
	}
}

// BypassesPreferences reports whether the purpose skips the E13-S01
// preference gate (safety and OTP always deliver).
func BypassesPreferences(purpose Purpose) bool {
	return purpose == PurposeSafety || purpose == PurposeOtp
}

// CategoryFor maps a purpose back to its preference category ("" for OTP,
// which has none).
func CategoryFor(purpose Purpose) domain.Category {
	switch purpose {
	case PurposePods:
		return domain.CategoryPods
	case PurposeRooms:
		return domain.CategoryRooms
	case PurposeSafety:
		return domain.CategorySafety
	case PurposeRitual:
		return domain.CategoryRitual
	default:
		return ""
	}
}
