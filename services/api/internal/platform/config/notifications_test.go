package config

import (
	"strings"
	"testing"
)

// productionBase is a complete non-development notification configuration.
// Individual tests remove or override one key to assert a single rule.
func productionBase() map[string]string {
	return map[string]string{
		"OTP_PROVIDERS":       "arkesel",
		"ARKESEL_API_KEY":     "synthetic-test-only",
		"ARKESEL_SENDER_ID":   "Obiara",
		"EMAIL_PROVIDER":      "resend",
		"RESEND_API_KEY":      "synthetic-test-only",
		"RESEND_FROM_ADDRESS": "no-reply@obiara.test",
		// WhatsApp is not provisioned yet; "disabled" is the explicit
		// out-of-service choice production must make.
		"WHATSAPP_PROVIDER": "disabled",
	}
}

func loadNotificationsFor(values map[string]string) NotificationsConfig {
	return loadNotifications(envWith(values))
}

// TestSimulatorsRejectedOutsideDevelopment pins the regression that shipped
// OTP registration to production against an in-memory sender: every code was
// minted, stored and never delivered, and the API still answered 202.
func TestSimulatorsRejectedOutsideDevelopment(t *testing.T) {
	t.Run("otp ladder", func(t *testing.T) {
		values := productionBase()
		values["OTP_PROVIDERS"] = "simulator"
		err := validateNotifications(loadNotificationsFor(values), false)
		if err == nil {
			t.Fatal("production accepted the simulator OTP sender")
		}
		if !strings.Contains(err.Error(), "OTP_PROVIDERS") {
			t.Fatalf("error should name the offending variable, got %v", err)
		}
	})

	t.Run("email", func(t *testing.T) {
		values := productionBase()
		values["EMAIL_PROVIDER"] = "simulator"
		if err := validateNotifications(loadNotificationsFor(values), false); err == nil {
			t.Fatal("production accepted the simulator email sender")
		}
	})

	t.Run("whatsapp rung", func(t *testing.T) {
		values := productionBase()
		values["OTP_PROVIDERS"] = "whatsapp"
		values["WHATSAPP_PROVIDER"] = "simulator"
		if err := validateNotifications(loadNotificationsFor(values), false); err == nil {
			t.Fatal("production accepted a whatsapp OTP rung backed by the simulator")
		}
	})

	t.Run("development still allows simulators", func(t *testing.T) {
		if err := validateNotifications(loadNotificationsFor(nil), true); err != nil {
			t.Fatalf("development rejected the default simulator configuration: %v", err)
		}
	})
}

func TestValidateNotificationsRequiresCredentials(t *testing.T) {
	cases := map[string]struct {
		mutate func(map[string]string)
		want   string
	}{
		"arkesel without key": {
			mutate: func(v map[string]string) { delete(v, "ARKESEL_API_KEY") },
			want:   "ARKESEL_API_KEY",
		},
		"arkesel without sender": {
			mutate: func(v map[string]string) { delete(v, "ARKESEL_SENDER_ID") },
			want:   "ARKESEL_SENDER_ID",
		},
		"twilio without credentials": {
			mutate: func(v map[string]string) { v["OTP_PROVIDERS"] = "twilio" },
			want:   "TWILIO_ACCOUNT_SID",
		},
		"twilio without an originator": {
			mutate: func(v map[string]string) {
				v["OTP_PROVIDERS"] = "twilio"
				v["TWILIO_ACCOUNT_SID"] = "synthetic-test-only"
				v["TWILIO_AUTH_TOKEN"] = "synthetic-test-only"
			},
			want: "TWILIO_FROM_NUMBER",
		},
		"resend without key": {
			mutate: func(v map[string]string) { delete(v, "RESEND_API_KEY") },
			want:   "RESEND_API_KEY",
		},
		"resend without from": {
			mutate: func(v map[string]string) { delete(v, "RESEND_FROM_ADDRESS") },
			want:   "RESEND_FROM_ADDRESS",
		},
		"meta without phone number id": {
			mutate: func(v map[string]string) { v["WHATSAPP_PROVIDER"] = "meta" },
			want:   "META_WHATSAPP_PHONE_NUMBER_ID",
		},
		"unknown otp provider": {
			mutate: func(v map[string]string) { v["OTP_PROVIDERS"] = "carrier-pigeon" },
			want:   "carrier-pigeon",
		},
		"template without a verb": {
			mutate: func(v map[string]string) { v["OTP_SMS_TEMPLATE"] = "your code is here" },
			want:   "OTP_SMS_TEMPLATE",
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			values := productionBase()
			testCase.mutate(values)
			err := validateNotifications(loadNotificationsFor(values), false)
			if err == nil {
				t.Fatal("configuration was accepted, want error")
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error %v should mention %q", err, testCase.want)
			}
		})
	}
}

func TestParseProvidersBuildsAnOrderedLadder(t *testing.T) {
	values := productionBase()
	values["OTP_PROVIDERS"] = " arkesel , twilio ,, arkesel "
	values["TWILIO_ACCOUNT_SID"] = "synthetic-test-only"
	values["TWILIO_AUTH_TOKEN"] = "synthetic-test-only"
	values["TWILIO_FROM_NUMBER"] = "+15550000000"

	cfg := loadNotificationsFor(values)
	want := []Provider{ProviderArkesel, ProviderTwilio}
	if len(cfg.OtpProviders) != len(want) {
		t.Fatalf("OtpProviders = %v, want %v", cfg.OtpProviders, want)
	}
	for i, provider := range want {
		if cfg.OtpProviders[i] != provider {
			t.Fatalf("OtpProviders = %v, want %v (order matters: rung 0 is tried first)", cfg.OtpProviders, want)
		}
	}
	if err := validateNotifications(cfg, false); err != nil {
		t.Fatalf("a fully credentialed ladder was rejected: %v", err)
	}
}

func TestUsesOtpProvider(t *testing.T) {
	cfg := loadNotificationsFor(map[string]string{"OTP_PROVIDERS": "arkesel,whatsapp"})
	if !cfg.UsesOtpProvider(ProviderWhatsApp) {
		t.Error("UsesOtpProvider(whatsapp) = false, want true")
	}
	if cfg.UsesOtpProvider(ProviderTwilio) {
		t.Error("UsesOtpProvider(twilio) = true, want false")
	}
}

// TestDisabledWhatsAppIsAllowedInProduction covers shipping without WhatsApp
// provisioned. "disabled" must be acceptable where "simulator" is not,
// because it fails sends loudly instead of reporting them delivered.
func TestDisabledWhatsAppIsAllowedInProduction(t *testing.T) {
	values := productionBase()
	values["WHATSAPP_PROVIDER"] = "disabled"

	if err := validateNotifications(loadNotificationsFor(values), false); err != nil {
		t.Fatalf("production rejected a deliberately disabled whatsapp channel: %v", err)
	}
}

// TestSimulatorWhatsAppIsRejectedEvenWhenOtpDoesNotUseIt closes the loophole
// where an SMS-only OTP ladder left the WhatsApp channel on the simulator and
// Nnoboa consent invites vanished silently.
func TestSimulatorWhatsAppIsRejectedEvenWhenOtpDoesNotUseIt(t *testing.T) {
	values := productionBase()
	values["OTP_PROVIDERS"] = "arkesel"
	values["WHATSAPP_PROVIDER"] = "simulator"

	err := validateNotifications(loadNotificationsFor(values), false)
	if err == nil {
		t.Fatal("production accepted a simulator whatsapp channel on an SMS-only ladder")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("error should point operators at the disabled provider: %v", err)
	}
}

func TestDisabledWhatsAppCannotBackAnOtpRung(t *testing.T) {
	values := productionBase()
	values["OTP_PROVIDERS"] = "whatsapp"
	values["WHATSAPP_PROVIDER"] = "disabled"

	if err := validateNotifications(loadNotificationsFor(values), false); err == nil {
		t.Fatal("accepted an OTP ladder resting on a disabled channel")
	}
}

func TestPushProviderSelection(t *testing.T) {
	t.Run("defaults to disabled", func(t *testing.T) {
		cfg := loadNotificationsFor(nil)
		if cfg.PushProvider != ProviderDisabled {
			t.Errorf("PushProvider = %q, want disabled", cfg.PushProvider)
		}
	})

	t.Run("expo is accepted without a token", func(t *testing.T) {
		values := productionBase()
		values["PUSH_PROVIDER"] = "expo"
		if err := validateNotifications(loadNotificationsFor(values), false); err != nil {
			t.Fatalf("production rejected expo: %v", err)
		}
	})

	// A push simulator reports delivery without delivering, which is the
	// failure this codebase has already paid for once.
	t.Run("simulator is rejected in production", func(t *testing.T) {
		values := productionBase()
		values["PUSH_PROVIDER"] = "simulator"
		err := validateNotifications(loadNotificationsFor(values), false)
		if err == nil {
			t.Fatal("production accepted the push simulator")
		}
		if !strings.Contains(err.Error(), "disabled") {
			t.Errorf("error should point at the disabled provider: %v", err)
		}
	})

	t.Run("unknown provider is rejected", func(t *testing.T) {
		values := productionBase()
		values["PUSH_PROVIDER"] = "firebase"
		if err := validateNotifications(loadNotificationsFor(values), false); err == nil {
			t.Fatal("accepted an unknown push provider")
		}
	})
}
