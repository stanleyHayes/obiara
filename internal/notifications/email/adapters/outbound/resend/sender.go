// Package resend is the production transactional email adapter backed by
// the Resend API. It implements the email channel's application.Sender
// port.
//
// Rendering lives here rather than in the domain: the domain owns which
// templates exist and which parameters are permitted, while the wire
// subject and body are a provider concern. Every parameter is HTML-escaped
// before interpolation, so a parameter can never inject markup.
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

var (
	// ErrNotConfigured reports missing credentials at construction time.
	ErrNotConfigured = errors.New("resend email sender is not configured")
	// ErrDeliveryFailed reports a provider-side rejection, carrying no
	// message content.
	ErrDeliveryFailed = errors.New("resend email delivery failed")
	// ErrTemplateUnmapped reports a domain template with no render spec.
	ErrTemplateUnmapped = errors.New("email template has no render spec")
)

const (
	defaultBaseURL   = "https://api.resend.com"
	maxResponseBytes = 1 << 16
	defaultTimeout   = 10 * time.Second
	maxAttempts      = 2
	retryBackoff     = 250 * time.Millisecond
)

// renderSpec is the subject and body shape of one template. body is a Go
// text template written with %s verbs filled from params in the order named
// by fields.
type renderSpec struct {
	subject string
	fields  []string
	body    string
}

// specs covers every template the domain admits. A domain template with no
// entry here fails closed rather than sending a blank email.
var specs = map[domain.Template]renderSpec{
	domain.TemplateAdminNotice: {
		subject: "Your Obiara admin sign-in code",
		fields:  []string{"code"},
		body: `<p>Your Obiara admin sign-in code is:</p>` +
			`<p style="font-size:28px;font-weight:700;letter-spacing:4px;margin:16px 0">%s</p>` +
			`<p>It expires in 5 minutes and can be used once.</p>` +
			`<p>If you did not request it, someone has your email address but not your access. ` +
			`Do not share this code — Obiara staff will never ask you for it.</p>`,
	},
	domain.TemplateOpsAlert: {
		subject: "Obiara operations alert",
		fields:  []string{"summary"},
		body:    `<p>An operations alert was raised:</p><p>%s</p><p>Open the admin console to review it.</p>`,
	},
	domain.TemplateVerificationHelp: {
		subject: "Help with your Obiara verification",
		fields:  []string{"reference"},
		body: `<p>We need one more step to finish your verification.</p>` +
			`<p>Your reference is <strong>%s</strong>. Open Obiara and follow the verification prompt to continue.</p>`,
	},
}

// Config carries Resend credentials and the sending identity.
type Config struct {
	APIKey string
	// From is the verified sender, either "name@domain" or
	// "Name <name@domain>".
	From string
	// ReplyTo is optional.
	ReplyTo string
	// BaseURL overrides the API host in tests.
	BaseURL string
}

// Sender delivers transactional email over Resend.
type Sender struct {
	client   *http.Client
	endpoint string
	apiKey   string
	from     string
	replyTo  string
}

// NewSender validates configuration and builds the adapter, failing closed
// on incomplete credentials.
func NewSender(config Config, client *http.Client) (*Sender, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	from := strings.TrimSpace(config.From)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: api key is empty", ErrNotConfigured)
	}
	if from == "" {
		return nil, fmt.Errorf("%w: from address is empty", ErrNotConfigured)
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
		endpoint: base + "/emails",
		apiKey:   apiKey,
		from:     from,
		replyTo:  strings.TrimSpace(config.ReplyTo),
	}, nil
}

// Send transmits one message and returns Resend's id as the provider
// reference, which the delivery-status webhook later correlates against.
func (sender *Sender) Send(ctx context.Context, message domain.Message) (string, error) {
	payload, err := sender.encode(message)
	if err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		providerRef, retryable, err := sender.post(ctx, payload)
		if err == nil {
			return providerRef, nil
		}
		lastErr = err
		if !retryable || attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(retryBackoff):
		}
	}
	return "", lastErr
}

func (sender *Sender) encode(message domain.Message) ([]byte, error) {
	spec, ok := specs[message.Template()]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTemplateUnmapped, message.Template())
	}
	params := message.Params()
	values := make([]any, 0, len(spec.fields))
	for _, field := range spec.fields {
		value, ok := params[field]
		if !ok {
			return nil, fmt.Errorf("%w: %s is missing parameter %q", ErrTemplateUnmapped, message.Template(), field)
		}
		values = append(values, html.EscapeString(value))
	}

	request := map[string]any{
		"from":    sender.from,
		"to":      []string{message.To()},
		"subject": spec.subject,
		"html":    fmt.Sprintf(spec.body, values...),
	}
	if sender.replyTo != "" {
		request["reply_to"] = sender.replyTo
	}
	return json.Marshal(request)
}

func (sender *Sender) post(ctx context.Context, payload []byte) (providerRef string, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", false, fmt.Errorf("resend: build request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+sender.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return "", true, fmt.Errorf("resend: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retryable = response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		return "", retryable, fmt.Errorf("resend: %w: %s", ErrDeliveryFailed, describe(response.StatusCode, body, readErr))
	}
	if readErr != nil {
		// Accepted but the receipt is unreadable; do not resend and
		// double-deliver.
		return "", false, nil
	}

	var receipt struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(body, &receipt) == nil {
		return receipt.ID, false, nil
	}
	return "", false, nil
}

// describe renders a safe diagnostic: HTTP status plus Resend's error name.
// The provider's message text is not echoed because it can quote the
// recipient or subject.
func describe(status int, body []byte, readErr error) string {
	rendered := "status " + strconv.Itoa(status)
	if readErr != nil || len(body) == 0 {
		return rendered
	}
	var payload struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Name != "" {
		return rendered + ", provider error " + payload.Name
	}
	return rendered
}

// DomainStatus is one sending domain as Resend sees it.
type DomainStatus struct {
	Name   string
	Status string
}

// Preflight validates the credentials without sending anything.
//
// It exists because a rejected key and an unverified domain both surface at
// send time as the same failed notification, and finding out which meant
// searching request logs for one correlation id. This asks Resend directly,
// at startup, using a read-only endpoint: no email is sent and no member is
// involved.
//
// The returned slice is every domain this key can send from, so a caller can
// say whether the configured sender is among them.
func (sender *Sender) Preflight(ctx context.Context) ([]DomainStatus, error) {
	endpoint := strings.TrimSuffix(sender.endpoint, "/emails") + "/domains"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("resend: build preflight request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+sender.apiKey)
	request.Header.Set("Accept", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("resend preflight: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("resend preflight: %w: %s",
			ErrNotConfigured, describe(response.StatusCode, body, readErr))
	}
	if readErr != nil {
		return nil, fmt.Errorf("resend preflight: %w: receipt unreadable", ErrDeliveryFailed)
	}

	var payload struct {
		Data []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("resend preflight: %w: receipt was not json", ErrDeliveryFailed)
	}
	domains := make([]DomainStatus, 0, len(payload.Data))
	for _, entry := range payload.Data {
		domains = append(domains, DomainStatus{Name: entry.Name, Status: entry.Status})
	}
	return domains, nil
}

// SenderDomain returns the domain part of the configured from address, for
// comparison against what Preflight reports.
func (sender *Sender) SenderDomain() string {
	address := sender.from
	// "Name <user@domain>" or bare "user@domain".
	if open := strings.LastIndex(address, "<"); open >= 0 {
		if close := strings.LastIndex(address, ">"); close > open {
			address = address[open+1 : close]
		}
	}
	if at := strings.LastIndex(address, "@"); at >= 0 {
		return strings.ToLower(strings.TrimSpace(address[at+1:]))
	}
	return ""
}
