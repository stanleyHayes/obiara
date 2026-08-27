// Package arkesel is the production SMS adapter for OTP delivery, backed by
// the Arkesel SMS v2 API. Arkesel is the primary OTP transport for Obiara:
// it terminates directly on the Ghanaian networks members actually use,
// which WhatsApp did not do reliably.
//
// This adapter carries the code only in the request body. Errors report the
// provider status and code, never the code, the API key, or the full
// recipient number.
//
// Arkesel answers business failures — an unapproved sender id, an exhausted
// balance, an invalid recipient — with HTTP 200 and a non-success "status"
// field. Treating 2xx as delivery would silently swallow exactly the
// failures that leave a member without a code, so the body status is
// authoritative here and the HTTP code is only a first gate.
package arkesel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var (
	// ErrNotConfigured reports missing credentials at construction time.
	ErrNotConfigured = errors.New("arkesel sms sender is not configured")
	// ErrDeliveryFailed reports a provider-side rejection.
	ErrDeliveryFailed = errors.New("arkesel sms delivery failed")
)

const (
	defaultBaseURL   = "https://sms.arkesel.com"
	maxResponseBytes = 1 << 16
	defaultTimeout   = 10 * time.Second
	maxAttempts      = 2
	retryBackoff     = 250 * time.Millisecond
	// maxSenderLength is the GSM alphanumeric originator limit. Arkesel
	// rejects anything longer, so it is caught at startup instead.
	maxSenderLength = 11
)

// Config carries Arkesel credentials and the sending identity.
type Config struct {
	APIKey string
	// Sender is the pre-approved alphanumeric sender id shown on the
	// handset, at most 11 characters.
	Sender string
	// BaseURL overrides the API host in tests.
	BaseURL string
	// Template renders the message body around the code. It must contain
	// exactly one "%s" verb. Empty uses the default Obiara wording.
	Template string
}

// Sender delivers OTP codes over Arkesel SMS.
type Sender struct {
	client   *http.Client
	endpoint string
	apiKey   string
	sender   string
	template string
}

// NewSender validates configuration and builds the adapter. It fails closed:
// an incompletely configured provider is a startup error, never a silently
// dropped code at request time.
func NewSender(config Config, client *http.Client) (*Sender, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	senderID := strings.TrimSpace(config.Sender)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: api key is empty", ErrNotConfigured)
	}
	if senderID == "" {
		return nil, fmt.Errorf("%w: sender id is empty", ErrNotConfigured)
	}
	if len(senderID) > maxSenderLength {
		return nil, fmt.Errorf("%w: sender id %q exceeds %d characters", ErrNotConfigured, senderID, maxSenderLength)
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
		endpoint: base + "/api/v2/sms/send",
		apiKey:   apiKey,
		sender:   senderID,
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
		return fmt.Errorf("%s: unexpected contact channel %q", "arkesel", contact.Channel())
	}
	phone := contact.Value()

	payload, err := json.Marshal(map[string]any{
		"sender":     sender.sender,
		"message":    fmt.Sprintf(sender.template, code),
		"recipients": []string{recipient(phone)},
	})
	if err != nil {
		return fmt.Errorf("arkesel: encode request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		retryable, err := sender.post(ctx, payload)
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

// recipient renders an E.164 number the way Arkesel documents recipients:
// country code and subscriber digits with no leading "+".
func recipient(phone string) string {
	return strings.TrimPrefix(strings.TrimSpace(phone), "+")
}

func (sender *Sender) post(ctx context.Context, payload []byte) (retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("arkesel: build request: %w", err)
	}
	request.Header.Set("api-key", sender.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return true, fmt.Errorf("arkesel: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retryable = response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		return retryable, fmt.Errorf("arkesel: %w: status %d%s", ErrDeliveryFailed, response.StatusCode, providerCode(body, readErr))
	}
	if readErr != nil {
		return true, fmt.Errorf("arkesel: %w: receipt unreadable", ErrDeliveryFailed)
	}

	// A 2xx with a non-success status is the common Arkesel failure shape.
	var receipt struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	if err := json.Unmarshal(body, &receipt); err != nil {
		return false, fmt.Errorf("arkesel: %w: receipt was not json", ErrDeliveryFailed)
	}
	if !accepted(receipt.Status) {
		// Balance and sender-id faults do not clear on a retry; only an
		// explicit provider-side rate limit does.
		retryable = strings.EqualFold(receipt.Status, "rate_limit")
		return retryable, fmt.Errorf("arkesel: %w: provider status %q%s",
			ErrDeliveryFailed, receipt.Status, providerCode(body, nil))
	}
	return false, nil
}

// accepted reports the statuses Arkesel uses for an accepted submission
// across its v1 and v2 response shapes.
func accepted(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success", "ok":
		return true
	default:
		return false
	}
}

// providerCode extracts Arkesel's numeric/string code when present, for a
// diagnostic that never echoes provider message text.
func providerCode(body []byte, readErr error) string {
	if readErr != nil || len(body) == 0 {
		return ""
	}
	var payload struct {
		Code json.RawMessage `json:"code"`
	}
	if json.Unmarshal(body, &payload) != nil || len(payload.Code) == 0 {
		return ""
	}
	var asString string
	if json.Unmarshal(payload.Code, &asString) == nil && asString != "" {
		return ", provider code " + asString
	}
	var asNumber int
	if json.Unmarshal(payload.Code, &asNumber) == nil && asNumber != 0 {
		return ", provider code " + strconv.Itoa(asNumber)
	}
	return ""
}
