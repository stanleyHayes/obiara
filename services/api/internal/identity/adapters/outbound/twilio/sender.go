// Package twilio is the production SMS adapter for OTP delivery
// (agent_plan.md §11: SMS primary, WhatsApp fallback). It implements the
// identity context's OtpSender port over the Twilio Messages REST API.
//
// OTP codes are secrets: they are written to the request body and nowhere
// else. Errors returned from this package carry the provider status and
// error code only — never the code, never the auth token, and never the
// full recipient number.
package twilio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// ErrNotConfigured reports missing credentials at construction time.
var ErrNotConfigured = errors.New("twilio sms sender is not configured")

// ErrDeliveryFailed reports a provider-side rejection. It deliberately
// carries no message content.
var ErrDeliveryFailed = errors.New("twilio sms delivery failed")

const (
	defaultBaseURL = "https://api.twilio.com"
	// maxResponseBytes bounds the provider response we will parse. Twilio
	// error/success payloads are well under 8 KiB; anything larger is a
	// misconfigured endpoint, not a message receipt.
	maxResponseBytes = 1 << 16
	defaultTimeout   = 10 * time.Second
	// maxAttempts covers one retry for transient provider faults. The OTP
	// request is user-facing and synchronous, so the ladder stays short.
	maxAttempts  = 2
	retryBackoff = 250 * time.Millisecond
)

// Config carries the Twilio credentials and routing. Either FromNumber or
// MessagingServiceSID must be set; MessagingServiceSID wins when both are
// present because it lets Twilio pick the best sender per destination.
type Config struct {
	AccountSID          string
	AuthToken           string
	FromNumber          string
	MessagingServiceSID string
	// BaseURL overrides the API host in tests. Empty uses the public API.
	BaseURL string
	// Template renders the message body around the code. It must contain
	// exactly one "%s" verb. Empty uses the default Obiara wording.
	Template string
}

// Sender delivers OTP codes over Twilio SMS.
type Sender struct {
	client   *http.Client
	endpoint string
	config   Config
	template string
}

// NewSender validates the configuration and builds the adapter. It fails
// closed: an incompletely configured provider is a startup error, never a
// silently dropped message at request time.
func NewSender(config Config, client *http.Client) (*Sender, error) {
	config.AccountSID = strings.TrimSpace(config.AccountSID)
	config.AuthToken = strings.TrimSpace(config.AuthToken)
	config.FromNumber = strings.TrimSpace(config.FromNumber)
	config.MessagingServiceSID = strings.TrimSpace(config.MessagingServiceSID)

	if config.AccountSID == "" {
		return nil, fmt.Errorf("%w: account sid is empty", ErrNotConfigured)
	}
	if config.AuthToken == "" {
		return nil, fmt.Errorf("%w: auth token is empty", ErrNotConfigured)
	}
	if config.FromNumber == "" && config.MessagingServiceSID == "" {
		return nil, fmt.Errorf("%w: one of from number or messaging service sid is required", ErrNotConfigured)
	}

	template := strings.TrimSpace(config.Template)
	if template == "" {
		template = "%s is your Obiara code. It expires in 10 minutes. Obiara will never ask you for this code."
	}
	if strings.Count(template, "%s") != 1 {
		return nil, fmt.Errorf("%w: template must contain exactly one %%s verb", ErrNotConfigured)
	}

	base := strings.TrimSpace(config.BaseURL)
	if base == "" {
		base = defaultBaseURL
	}
	base = strings.TrimSuffix(base, "/")

	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	return &Sender{
		client:   client,
		endpoint: base + "/2010-04-01/Accounts/" + url.PathEscape(config.AccountSID) + "/Messages.json",
		config:   config,
		template: template,
	}, nil
}

// Send delivers the code to phone (E.164). It satisfies the identity
// context's application.OtpSender port.
func (sender *Sender) Send(ctx context.Context, contact domain.Contact, code string) error {
	// The router only sends SMS contacts here; a different channel means a
	// composition mistake, and delivering a sign-in code to the wrong
	// transport is exactly what must not happen quietly.
	if contact.Channel() != domain.ChannelSMS {
		return fmt.Errorf("%s: unexpected contact channel %q", "twilio", contact.Channel())
	}
	phone := contact.Value()

	form := url.Values{}
	form.Set("To", phone)
	if sender.config.MessagingServiceSID != "" {
		form.Set("MessagingServiceSid", sender.config.MessagingServiceSID)
	} else {
		form.Set("From", sender.config.FromNumber)
	}
	form.Set("Body", fmt.Sprintf(sender.template, code))
	encoded := form.Encode()

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		retryable, err := sender.post(ctx, encoded)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryable || attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryBackoff):
		}
	}
	return lastErr
}

// post performs one delivery attempt. The bool reports whether the failure
// is worth retrying (network faults, 429 and 5xx).
func (sender *Sender) post(ctx context.Context, encoded string) (retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, strings.NewReader(encoded))
	if err != nil {
		return false, fmt.Errorf("twilio: build request: %w", err)
	}
	request.SetBasicAuth(sender.config.AccountSID, sender.config.AuthToken)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		// Transport errors can embed the request URL but never the body,
		// so the code cannot leak through this path.
		return true, fmt.Errorf("twilio: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return false, nil
	}

	retryable = response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
	return retryable, fmt.Errorf("twilio: %w: %s", ErrDeliveryFailed, describe(response))
}

// describe renders a safe diagnostic from a failed response: the HTTP
// status plus Twilio's numeric error code when present. Provider message
// text is not echoed, because it can quote the message body.
func describe(response *http.Response) string {
	status := strconv.Itoa(response.StatusCode)
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil || len(body) == 0 {
		return "status " + status
	}
	var payload struct {
		Code int `json:"code"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Code != 0 {
		return "status " + status + ", provider code " + strconv.Itoa(payload.Code)
	}
	return "status " + status
}
