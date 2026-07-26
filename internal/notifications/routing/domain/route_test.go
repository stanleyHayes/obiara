package domain

import (
	"testing"

	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
)

func TestLadders(t *testing.T) {
	otp := LadderFor(PurposeOtp)
	if len(otp) != 2 || otp[0] != ChannelSMS || otp[1] != ChannelWhatsApp {
		t.Fatalf("otp ladder = %v, want SMS-primary WhatsApp-fallback", otp)
	}
	safety := LadderFor(PurposeSafety)
	if len(safety) != 4 || safety[0] != ChannelPush {
		t.Fatalf("safety ladder = %v", safety)
	}
	pods := LadderFor(PurposePods)
	if len(pods) != 3 || pods[2] != ChannelInApp {
		t.Fatalf("pods ladder = %v", pods)
	}
	if got := LadderFor(Purpose("unknown")); len(got) != 1 || got[0] != ChannelInApp {
		t.Fatalf("unknown purpose ladder = %v, want in-app only", got)
	}
}

func TestBypassesPreferences(t *testing.T) {
	for purpose, want := range map[Purpose]bool{
		PurposeOtp:    true,
		PurposeSafety: true,
		PurposePods:   false,
		PurposeRitual: false,
		PurposeRooms:  false,
	} {
		if got := BypassesPreferences(purpose); got != want {
			t.Fatalf("BypassesPreferences(%s) = %v, want %v", purpose, got, want)
		}
	}
}

func TestCategoryFor(t *testing.T) {
	if CategoryFor(PurposePods) != notificationdomain.CategoryPods {
		t.Fatal("pods maps to pods")
	}
	if CategoryFor(PurposeSafety) != notificationdomain.CategorySafety {
		t.Fatal("safety maps to safety")
	}
	if CategoryFor(PurposeOtp) != "" {
		t.Fatal("otp has no preference category")
	}
}
