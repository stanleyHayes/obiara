package config

import (
	"fmt"
	"slices"
	"strings"
)

// Provider names the outbound adapter selected for a channel.
type Provider string

const (
	// ProviderSimulator delivers nothing. It exists for local development
	// and tests and is rejected outside them.
	ProviderSimulator Provider = "simulator"
	// ProviderArkesel is the primary Ghanaian SMS transport for OTP.
	ProviderArkesel Provider = "arkesel"
	// ProviderTwilio is the secondary international SMS transport.
	ProviderTwilio Provider = "twilio"
	// ProviderWhatsApp routes OTP over the WhatsApp Business channel.
	ProviderWhatsApp Provider = "whatsapp"
	// ProviderMeta is the WhatsApp Business Cloud API adapter.
	ProviderMeta Provider = "meta"
	// ProviderResend is the transactional email adapter.
	ProviderResend Provider = "resend"
	// ProviderDisabled marks a channel as deliberately out of service. Unlike
	// the simulator it fails every send loudly, so a dark channel is visible
	// instead of silently swallowing messages.
	ProviderDisabled Provider = "disabled"
)

// NotificationsConfig selects and configures the outbound delivery adapters.
//
// Every channel defaults to the simulator so local development needs no
// credentials, and Validate rejects the simulator outside development and
// test. That pairing is deliberate: an unset provider in production used to
// mean "mint an OTP, store it in memory, and tell the member it was sent",
// which is precisely the failure this configuration exists to make
// impossible.
type NotificationsConfig struct {
	// OtpProviders is the ordered OTP delivery ladder. The first rung that
	// accepts the message wins.
	OtpProviders []Provider
	// OtpSmsTemplate renders the SMS body and must hold exactly one %s.
	OtpSmsTemplate string

	ArkeselAPIKey  string
	ArkeselSender  string
	ArkeselBaseURL string

	TwilioAccountSID          string
	TwilioAuthToken           string
	TwilioFromNumber          string
	TwilioMessagingServiceSID string
	TwilioBaseURL             string

	WhatsAppProvider   Provider
	MetaPhoneNumberID  string
	MetaAccessToken    string
	MetaAPIVersion     string
	MetaLanguageCode   string
	MetaOtpTemplate    string
	MetaPodTemplate    string
	MetaNnoboaTemplate string
	MetaBaseURL        string

	EmailProvider Provider
	ResendAPIKey  string
	ResendFrom    string
	ResendReplyTo string
	ResendBaseURL string
}

// UsesOtpProvider reports whether the ladder contains the given rung.
func (config NotificationsConfig) UsesOtpProvider(provider Provider) bool {
	return slices.Contains(config.OtpProviders, provider)
}

func loadNotifications(getenv func(string) string) NotificationsConfig {
	return NotificationsConfig{
		OtpProviders:   parseProviders(getenv("OTP_PROVIDERS"), ProviderSimulator),
		OtpSmsTemplate: strings.TrimSpace(getenv("OTP_SMS_TEMPLATE")),

		ArkeselAPIKey:  strings.TrimSpace(getenv("ARKESEL_API_KEY")),
		ArkeselSender:  strings.TrimSpace(getenv("ARKESEL_SENDER_ID")),
		ArkeselBaseURL: strings.TrimSpace(getenv("ARKESEL_BASE_URL")),

		TwilioAccountSID:          strings.TrimSpace(getenv("TWILIO_ACCOUNT_SID")),
		TwilioAuthToken:           strings.TrimSpace(getenv("TWILIO_AUTH_TOKEN")),
		TwilioFromNumber:          strings.TrimSpace(getenv("TWILIO_FROM_NUMBER")),
		TwilioMessagingServiceSID: strings.TrimSpace(getenv("TWILIO_MESSAGING_SERVICE_SID")),
		TwilioBaseURL:             strings.TrimSpace(getenv("TWILIO_BASE_URL")),

		WhatsAppProvider:   provider(getenv("WHATSAPP_PROVIDER"), ProviderSimulator),
		MetaPhoneNumberID:  strings.TrimSpace(getenv("META_WHATSAPP_PHONE_NUMBER_ID")),
		MetaAccessToken:    strings.TrimSpace(getenv("META_WHATSAPP_ACCESS_TOKEN")),
		MetaAPIVersion:     strings.TrimSpace(getenv("META_WHATSAPP_API_VERSION")),
		MetaLanguageCode:   strings.TrimSpace(getenv("META_WHATSAPP_LANGUAGE")),
		MetaOtpTemplate:    strings.TrimSpace(getenv("META_WHATSAPP_TEMPLATE_OTP")),
		MetaPodTemplate:    strings.TrimSpace(getenv("META_WHATSAPP_TEMPLATE_POD_ALERT")),
		MetaNnoboaTemplate: strings.TrimSpace(getenv("META_WHATSAPP_TEMPLATE_NNOBOA_CONSENT")),
		MetaBaseURL:        strings.TrimSpace(getenv("META_WHATSAPP_BASE_URL")),

		EmailProvider: provider(getenv("EMAIL_PROVIDER"), ProviderSimulator),
		ResendAPIKey:  strings.TrimSpace(getenv("RESEND_API_KEY")),
		ResendFrom:    strings.TrimSpace(getenv("RESEND_FROM_ADDRESS")),
		ResendReplyTo: strings.TrimSpace(getenv("RESEND_REPLY_TO")),
		ResendBaseURL: strings.TrimSpace(getenv("RESEND_BASE_URL")),
	}
}

func provider(value string, fallback Provider) Provider {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return fallback
	}
	return Provider(trimmed)
}

func parseProviders(value string, fallback Provider) []Provider {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []Provider{fallback}
	}
	var providers []Provider
	for _, part := range strings.Split(trimmed, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		candidate := Provider(name)
		// A ladder that tries the same transport twice is a typo, not a
		// retry policy; the adapters already retry internally.
		if !slices.Contains(providers, candidate) {
			providers = append(providers, candidate)
		}
	}
	if len(providers) == 0 {
		return []Provider{fallback}
	}
	return providers
}

// validateNotifications checks provider selection and the credentials each
// selected provider needs. simulatorsAllowed is false outside development
// and test, which is what stops a production deploy from silently going
// dark.
func validateNotifications(config NotificationsConfig, simulatorsAllowed bool) error {
	if len(config.OtpProviders) == 0 {
		return fmt.Errorf("OTP_PROVIDERS must name at least one provider")
	}
	for _, selected := range config.OtpProviders {
		switch selected {
		case ProviderSimulator:
			if !simulatorsAllowed {
				return fmt.Errorf(
					"OTP_PROVIDERS may not include %q outside development: members would never receive a code",
					ProviderSimulator)
			}
		case ProviderArkesel:
			if config.ArkeselAPIKey == "" {
				return fmt.Errorf("ARKESEL_API_KEY is required when OTP_PROVIDERS includes %q", ProviderArkesel)
			}
			if config.ArkeselSender == "" {
				return fmt.Errorf("ARKESEL_SENDER_ID is required when OTP_PROVIDERS includes %q", ProviderArkesel)
			}
		case ProviderTwilio:
			if config.TwilioAccountSID == "" {
				return fmt.Errorf("TWILIO_ACCOUNT_SID is required when OTP_PROVIDERS includes %q", ProviderTwilio)
			}
			if config.TwilioAuthToken == "" {
				return fmt.Errorf("TWILIO_AUTH_TOKEN is required when OTP_PROVIDERS includes %q", ProviderTwilio)
			}
			if config.TwilioFromNumber == "" && config.TwilioMessagingServiceSID == "" {
				return fmt.Errorf(
					"one of TWILIO_FROM_NUMBER or TWILIO_MESSAGING_SERVICE_SID is required when OTP_PROVIDERS includes %q",
					ProviderTwilio)
			}
		case ProviderWhatsApp:
			// The rung reuses the WhatsApp channel, so that channel must be
			// on a real provider too.
			if config.WhatsAppProvider == ProviderSimulator && !simulatorsAllowed {
				return fmt.Errorf(
					"OTP_PROVIDERS includes %q but WHATSAPP_PROVIDER is %q",
					ProviderWhatsApp, ProviderSimulator)
			}
		default:
			return fmt.Errorf(
				"OTP_PROVIDERS contains unknown provider %q (want %q, %q, %q or %q)",
				selected, ProviderArkesel, ProviderTwilio, ProviderWhatsApp, ProviderSimulator)
		}
	}

	switch config.WhatsAppProvider {
	case ProviderSimulator:
		if !simulatorsAllowed {
			return fmt.Errorf(
				"WHATSAPP_PROVIDER may not be %q outside development: Nnoboa consent invites would be reported as sent and reach nobody. Use %q to take the channel out of service explicitly",
				ProviderSimulator, ProviderDisabled)
		}
	case ProviderDisabled:
		// Deliberately out of service. Only an OTP ladder that depends on it
		// is a contradiction.
		if config.UsesOtpProvider(ProviderWhatsApp) {
			return fmt.Errorf(
				"OTP_PROVIDERS includes %q but WHATSAPP_PROVIDER is %q",
				ProviderWhatsApp, ProviderDisabled)
		}
	case ProviderMeta:
		if config.MetaPhoneNumberID == "" {
			return fmt.Errorf("META_WHATSAPP_PHONE_NUMBER_ID is required when WHATSAPP_PROVIDER is %q", ProviderMeta)
		}
		if config.MetaAccessToken == "" {
			return fmt.Errorf("META_WHATSAPP_ACCESS_TOKEN is required when WHATSAPP_PROVIDER is %q", ProviderMeta)
		}
	default:
		return fmt.Errorf("WHATSAPP_PROVIDER must be %q, %q or %q, got %q",
			ProviderMeta, ProviderDisabled, ProviderSimulator, config.WhatsAppProvider)
	}

	switch config.EmailProvider {
	case ProviderSimulator:
		if !simulatorsAllowed {
			return fmt.Errorf(
				"EMAIL_PROVIDER may not be %q outside development: admin sign-in codes would never arrive",
				ProviderSimulator)
		}
	case ProviderResend:
		if config.ResendAPIKey == "" {
			return fmt.Errorf("RESEND_API_KEY is required when EMAIL_PROVIDER is %q", ProviderResend)
		}
		if config.ResendFrom == "" {
			return fmt.Errorf("RESEND_FROM_ADDRESS is required when EMAIL_PROVIDER is %q", ProviderResend)
		}
	default:
		return fmt.Errorf("EMAIL_PROVIDER must be %q or %q, got %q",
			ProviderResend, ProviderSimulator, config.EmailProvider)
	}

	if config.OtpSmsTemplate != "" && strings.Count(config.OtpSmsTemplate, "%s") != 1 {
		return fmt.Errorf("OTP_SMS_TEMPLATE must contain exactly one %%s verb for the code")
	}
	return nil
}
