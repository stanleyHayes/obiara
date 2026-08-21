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
	"time"

	emailresend "github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/resend"
	emailsimulator "github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/simulator"
	emailapp "github.com/stanleyHayes/obiara/internal/notifications/email/application"
	whatsappdisabled "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/disabled"
	whatsappmeta "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/meta"
	whatsappsimulator "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/simulator"
	whatsappapp "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/arkesel"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/failover"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/twilio"
	identityapp "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
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

// OtpSender builds the OTP delivery ladder named by OTP_PROVIDERS.
// whatsappChannel supplies the "whatsapp" rung and may be zero when that
// rung is not selected.
func OtpSender(
	cfg config.NotificationsConfig,
	whatsappChannel whatsappapp.ChannelService,
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

	sender, err := failover.NewSender(ladderObserver(names, logger), rungs...)
	if err != nil {
		return nil, fmt.Errorf("build otp ladder: %w", err)
	}
	return sender, nil
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

func (sender whatsappOtpSender) Send(ctx context.Context, phone, code string) error {
	_, err := sender.channel.SendOtp(ctx, phone, code)
	return err
}
