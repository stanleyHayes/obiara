// Package delivery turns validated notification configuration into live
// outbound adapters. It is the one place that knows which provider packages
// exist, so the composition root stays a wiring list and the modules stay
// free of provider imports.
//
// Every constructor here fails closed. Configuration has already been
// validated, so a failure at this layer is a credential the provider itself
// rejected — which must stop the deploy, not degrade quietly into a channel
// that accepts messages and drops them.
package delivery

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	emailresend "github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/resend"
	emailsimulator "github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/simulator"
	emailapp "github.com/stanleyHayes/obiara/internal/notifications/email/application"
	pushdisabled "github.com/stanleyHayes/obiara/internal/notifications/push/adapters/outbound/disabled"
	pushexpo "github.com/stanleyHayes/obiara/internal/notifications/push/adapters/outbound/expo"
	pushapp "github.com/stanleyHayes/obiara/internal/notifications/push/application"
	whatsappdisabled "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/disabled"
	whatsappmeta "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/meta"
	whatsappsimulator "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/simulator"
	whatsappapp "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/arkesel"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/channel"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/emailotp"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/failover"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/twilio"
	identityapp "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// providerTimeout bounds a single outbound provider call. It is deliberately
// shorter than a typical gateway timeout so a stalled provider surfaces as a
// failed rung the ladder can fall through, not as a hung sign-up request.
const providerTimeout = 10 * time.Second

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: providerTimeout}
}

// WhatsAppSender builds the WhatsApp provider adapter.
func WhatsAppSender(cfg config.NotificationsConfig) (whatsappapp.Sender, error) {
	switch cfg.WhatsAppProvider {
	case config.ProviderMeta:
		names := map[whatsappdomain.Template]string{
			whatsappdomain.TemplateOtpCode:       cfg.MetaOtpTemplate,
			whatsappdomain.TemplatePodAlert:      cfg.MetaPodTemplate,
			whatsappdomain.TemplateNnoboaConsent: cfg.MetaNnoboaTemplate,
		}
		sender, err := whatsappmeta.NewSender(whatsappmeta.Config{
			PhoneNumberID: cfg.MetaPhoneNumberID,
			AccessToken:   cfg.MetaAccessToken,
			APIVersion:    cfg.MetaAPIVersion,
			LanguageCode:  cfg.MetaLanguageCode,
			TemplateNames: names,
			BaseURL:       cfg.MetaBaseURL,
		}, newHTTPClient())
		if err != nil {
			return nil, fmt.Errorf("build whatsapp sender: %w", err)
		}
		return sender, nil
	case config.ProviderDisabled:
		return whatsappdisabled.NewSender(), nil
	case config.ProviderSimulator:
		return whatsappsimulator.NewSender(), nil
	default:
		return nil, fmt.Errorf("unknown whatsapp provider %q", cfg.WhatsAppProvider)
	}
}

// PushSender builds the device push provider adapter.
func PushSender(cfg config.NotificationsConfig) (pushapp.Sender, error) {
	switch cfg.PushProvider {
	case config.ProviderExpo:
		sender, err := pushexpo.NewSender(pushexpo.Config{
			AccessToken: cfg.ExpoAccessToken,
			BaseURL:     cfg.ExpoBaseURL,
		}, newHTTPClient())
		if err != nil {
			return nil, fmt.Errorf("build push sender: %w", err)
		}
		return sender, nil
	case config.ProviderDisabled:
		return pushdisabled.NewSender(), nil
	case config.ProviderSimulator:
		return pushdisabled.NewSender(), nil
	default:
		return nil, fmt.Errorf("unknown push provider %q", cfg.PushProvider)
	}
}

// EmailSender builds the transactional email provider adapter.
func EmailSender(cfg config.NotificationsConfig) (emailapp.Sender, error) {
	switch cfg.EmailProvider {
	case config.ProviderResend:
		sender, err := emailresend.NewSender(emailresend.Config{
			APIKey:  cfg.ResendAPIKey,
			From:    cfg.ResendFrom,
			ReplyTo: cfg.ResendReplyTo,
			BaseURL: cfg.ResendBaseURL,
		}, newHTTPClient())
		if err != nil {
			return nil, fmt.Errorf("build email sender: %w", err)
		}
		return sender, nil
	case config.ProviderSimulator:
		return emailsimulator.NewSender(), nil
	default:
		return nil, fmt.Errorf("unknown email provider %q", cfg.EmailProvider)
	}
}

// PreflightEmail checks the email credentials at startup without sending
// anything.
//
// A rejected key and an unverified sending domain both surface at send time
// as the same failed notification, which previously meant hunting one
// correlation id through request logs to tell them apart. This asks the
// provider directly, on boot, so the answer is at the top of the deploy log
// instead.
//
// It never fails startup. Email being misconfigured must not take the API
// down for everyone; it has to be loud, not fatal.
func PreflightEmail(ctx context.Context, sender emailapp.Sender, logger *slog.Logger) {
	if logger == nil {
		return
	}
	preflight, ok := sender.(interface {
		Preflight(context.Context) ([]emailresend.DomainStatus, error)
		SenderDomain() string
	})
	if !ok {
		// The simulator and any future provider without a preflight.
		return
	}

	domains, err := preflight.Preflight(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "email provider preflight failed",
			slog.String("provider", "resend"),
			slog.String("fault", err.Error()),
			slog.String("hint", "the API key was rejected; check RESEND_API_KEY"))
		return
	}

	configured := preflight.SenderDomain()
	verified := make([]string, 0, len(domains))
	matched := false
	for _, entry := range domains {
		verified = append(verified, entry.Name+"="+entry.Status)
		if strings.EqualFold(entry.Name, configured) && strings.EqualFold(entry.Status, "verified") {
			matched = true
		}
	}

	if matched {
		logger.InfoContext(ctx, "email provider ready",
			slog.String("provider", "resend"), slog.String("senderDomain", configured))
		return
	}
	// The key works, so the sender domain is the problem. Naming both sides
	// turns this from "delivery failed" into one obvious mismatch.
	logger.ErrorContext(ctx, "email sender domain is not verified for this key",
		slog.String("provider", "resend"),
		slog.String("senderDomain", configured),
		slog.String("verifiedDomains", strings.Join(verified, ", ")),
		slog.String("hint", "RESEND_FROM_ADDRESS must use a domain listed as verified above"))
}

// OtpSender builds the per-channel OTP delivery routes.
//
// SMS gets the failover ladder named by OTP_PROVIDERS. Email gets a single
// route through the email context, present whenever an email service was
// configured — a member who verified an address must receive their code
// there, so there is nothing to fall back to and nothing to choose.
//
// whatsappChannel supplies the "whatsapp" rung and may be zero when that
// rung is not selected. emails may be nil, in which case the deployment
// simply cannot offer email sign-in and says so at boot.
func OtpSender(
	cfg config.NotificationsConfig,
	whatsappChannel whatsappapp.ChannelService,
	emails emailotp.EmailService,
	logger *slog.Logger,
) (identityapp.OtpSender, error) {
	rungs := make([]identityapp.OtpSender, 0, len(cfg.OtpProviders))
	names := make([]string, 0, len(cfg.OtpProviders))

	for _, selected := range cfg.OtpProviders {
		switch selected {
		case config.ProviderArkesel:
			sender, err := arkesel.NewSender(arkesel.Config{
				APIKey:   cfg.ArkeselAPIKey,
				Sender:   cfg.ArkeselSender,
				BaseURL:  cfg.ArkeselBaseURL,
				Template: cfg.OtpSmsTemplate,
			}, newHTTPClient())
			if err != nil {
				return nil, fmt.Errorf("build arkesel otp sender: %w", err)
			}
			rungs = append(rungs, sender)
		case config.ProviderTwilio:
			sender, err := twilio.NewSender(twilio.Config{
				AccountSID:          cfg.TwilioAccountSID,
				AuthToken:           cfg.TwilioAuthToken,
				FromNumber:          cfg.TwilioFromNumber,
				MessagingServiceSID: cfg.TwilioMessagingServiceSID,
				BaseURL:             cfg.TwilioBaseURL,
				Template:            cfg.OtpSmsTemplate,
			}, newHTTPClient())
			if err != nil {
				return nil, fmt.Errorf("build twilio otp sender: %w", err)
			}
			rungs = append(rungs, sender)
		case config.ProviderWhatsApp:
			rungs = append(rungs, whatsappOtpSender{channel: whatsappChannel})
		case config.ProviderSimulator:
			rungs = append(rungs, simulator.NewSender())
		default:
			return nil, fmt.Errorf("unknown otp provider %q", selected)
		}
		names = append(names, string(selected))
	}

	routes := make(map[identitydomain.Channel]identityapp.OtpSender, 2)
	if len(rungs) > 0 {
		ladder, err := failover.NewSender(ladderObserver(names, logger), rungs...)
		if err != nil {
			return nil, fmt.Errorf("build otp ladder: %w", err)
		}
		routes[identitydomain.ChannelSMS] = ladder
	}
	if emails != nil {
		emailSender, err := emailotp.NewSender(emails)
		if err != nil {
			return nil, fmt.Errorf("build email otp sender: %w", err)
		}
		routes[identitydomain.ChannelEmail] = emailSender
	}

	router, err := channel.NewRouter(routes)
	if err != nil {
		return nil, fmt.Errorf("build otp channel router: %w", err)
	}
	if logger != nil {
		available := make([]string, 0, 2)
		for _, name := range router.Channels() {
			available = append(available, string(name))
		}
		// Boot states plainly which sign-in channels members can actually
		// use, so a missing provider is visible here rather than as a member
		// staring at a code that never arrives.
		logger.Info("otp channels ready", "channels", strings.Join(available, ","))
	}
	return router, nil
}

// ladderObserver reports rung outcomes. A rung failure is warned rather than
// errored because a lower rung may still deliver; the caller reports the
// terminal failure when every rung is exhausted. Neither the code nor the
// recipient is logged.
func ladderObserver(names []string, logger *slog.Logger) failover.Observer {
	if logger == nil {
		return nil
	}
	return func(ctx context.Context, index int, err error) {
		name := "unknown"
		if index >= 0 && index < len(names) {
			name = names[index]
		}
		if err != nil {
			logger.WarnContext(ctx, "otp delivery rung failed",
				slog.String("provider", name), slog.Int("rung", index), slog.String("error", err.Error()))
			return
		}
		logger.InfoContext(ctx, "otp delivered",
			slog.String("provider", name), slog.Int("rung", index))
	}
}

// whatsappOtpSender adapts the WhatsApp channel to the identity context's
// OtpSender port.
type whatsappOtpSender struct {
	channel whatsappapp.ChannelService
}

func (sender whatsappOtpSender) Send(ctx context.Context, contact domain.Contact, code string) error {
	// The router only sends SMS contacts here; a different channel means a
	// composition mistake, and delivering a sign-in code to the wrong
	// transport is exactly what must not happen quietly.
	if contact.Channel() != domain.ChannelSMS {
		return fmt.Errorf("%s: unexpected contact channel %q", "whatsapp", contact.Channel())
	}
	phone := contact.Value()

	_, err := sender.channel.SendOtp(ctx, phone, code)
	return err
}
