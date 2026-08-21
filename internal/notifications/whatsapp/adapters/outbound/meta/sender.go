// Package meta is the production WhatsApp adapter backed by the WhatsApp
// Business Cloud API (graph.facebook.com). It implements the channel's
// application.Sender port.
//
// FR-701 is enforced upstream in the domain: every message is a
// provider-approved template with bounded parameters, so this adapter only
// has to map those parameters onto Meta's component wire format in a fixed,
// per-template order. Free text can never reach it.
package meta

import (
	"bytes"
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

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
)

var (
	// ErrNotConfigured reports missing credentials at construction time.
	ErrNotConfigured = errors.New("meta whatsapp sender is not configured")
	// ErrDeliveryFailed reports a provider-side rejection, carrying no
	// message content.
	ErrDeliveryFailed = errors.New("meta whatsapp delivery failed")
	// ErrTemplateUnmapped reports a domain template with no wire mapping.
	// It is a composition error, not a runtime condition.
	ErrTemplateUnmapped = errors.New("whatsapp template has no meta component mapping")
)

const (
	defaultBaseURL    = "https://graph.facebook.com"
	defaultAPIVersion = "v21.0"
	defaultLanguage   = "en"
	maxResponseBytes  = 1 << 16
	defaultTimeout    = 10 * time.Second
	maxAttempts       = 2
	retryBackoff      = 250 * time.Millisecond
)

// templateSpec fixes the wire shape of one approved template. bodyParams
// names the domain parameter keys in the exact positional order Meta
// expects. copyCodeParam, when set, names the parameter that also fills the
// template's one-tap copy-code button (Meta requires authentication
// templates to repeat the code in the button component).
type templateSpec struct {
	bodyParams    []string
	copyCodeParam string
	urlParams     []string
}

// specs is the whole mapping surface. Adding a template to the domain
// without adding it here fails closed at send time with ErrTemplateUnmapped.
var specs = map[domain.Template]templateSpec{
	domain.TemplateOtpCode: {
		bodyParams:    []string{"code"},
		copyCodeParam: "code",
	},
	domain.TemplatePodAlert: {
		bodyParams: []string{"ref"},
	},
	domain.TemplateNnoboaConsent: {
		bodyParams: []string{"kin_name"},
		urlParams:  []string{"consent_token"},
	},
}

// Config carries Cloud API credentials and routing.
type Config struct {
	// PhoneNumberID is the Cloud API sender identity (not a phone number).
	PhoneNumberID string
	AccessToken   string
	// APIVersion pins the Graph version, e.g. "v21.0". Empty uses the
	// package default so an unattended deploy never floats to a version
	// whose breaking changes have not been reviewed.
	APIVersion string
	// LanguageCode is the approved template locale, e.g. "en" or "en_US".
	LanguageCode string
	// TemplateNames maps a domain template onto the name it was approved
	// under in the WhatsApp Manager. Missing entries use the domain name.
	TemplateNames map[domain.Template]string
	// BaseURL overrides the Graph host in tests.
	BaseURL string
}

// Sender delivers approved templates over the Cloud API.
type Sender struct {
	client    *http.Client
	endpoint  string
	token     string
	language  string
	templates map[domain.Template]string
}

// NewSender validates configuration and builds the adapter, failing closed
// on incomplete credentials.
func NewSender(config Config, client *http.Client) (*Sender, error) {
	phoneNumberID := strings.TrimSpace(config.PhoneNumberID)
	token := strings.TrimSpace(config.AccessToken)
	if phoneNumberID == "" {
		return nil, fmt.Errorf("%w: phone number id is empty", ErrNotConfigured)
	}
	if token == "" {
		return nil, fmt.Errorf("%w: access token is empty", ErrNotConfigured)
	}

	version := strings.TrimSpace(config.APIVersion)
	if version == "" {
		version = defaultAPIVersion
	}
	if !strings.HasPrefix(version, "v") {
		return nil, fmt.Errorf("%w: api version %q must look like v21.0", ErrNotConfigured, version)
	}
	language := strings.TrimSpace(config.LanguageCode)
	if language == "" {
		language = defaultLanguage
	}
	base := strings.TrimSpace(config.BaseURL)
	if base == "" {
		base = defaultBaseURL
	}
	base = strings.TrimSuffix(base, "/")

	templates := make(map[domain.Template]string, len(specs))
	for template := range specs {
		name := strings.TrimSpace(config.TemplateNames[template])
		if name == "" {
			name = string(template)
		}
		templates[template] = name
	}

	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	return &Sender{
		client:    client,
		endpoint:  base + "/" + url.PathEscape(version) + "/" + url.PathEscape(phoneNumberID) + "/messages",
		token:     token,
		language:  language,
		templates: templates,
	}, nil
}

type textParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type component struct {
	Type       string          `json:"type"`
	SubType    string          `json:"sub_type,omitempty"`
	Index      string          `json:"index,omitempty"`
	Parameters []textParameter `json:"parameters"`
}

// Send transmits one approved template. It satisfies the channel's
// application.Sender port and returns Meta's message id as the provider
// reference so delivery webhooks can be correlated later.
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

	components := make([]component, 0, 3)
	if len(spec.bodyParams) > 0 {
		body := component{Type: "body", Parameters: make([]textParameter, 0, len(spec.bodyParams))}
		for _, key := range spec.bodyParams {
			value, ok := params[key]
			if !ok {
				return nil, fmt.Errorf("%w: %s is missing parameter %q", ErrTemplateUnmapped, message.Template(), key)
			}
			body.Parameters = append(body.Parameters, textParameter{Type: "text", Text: value})
		}
		components = append(components, body)
	}
	if spec.copyCodeParam != "" {
		value, ok := params[spec.copyCodeParam]
		if !ok {
			return nil, fmt.Errorf("%w: %s is missing parameter %q", ErrTemplateUnmapped, message.Template(), spec.copyCodeParam)
		}
		components = append(components, component{
			Type: "button", SubType: "copy_code", Index: "0",
			Parameters: []textParameter{{Type: "coupon_code", Text: value}},
		})
	}
	for offset, key := range spec.urlParams {
		value, ok := params[key]
		if !ok {
			return nil, fmt.Errorf("%w: %s is missing parameter %q", ErrTemplateUnmapped, message.Template(), key)
		}
		components = append(components, component{
			Type: "button", SubType: "url", Index: strconv.Itoa(offset),
			Parameters: []textParameter{{Type: "text", Text: value}},
		})
	}

	return json.Marshal(map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                message.To(),
		"type":              "template",
		"template": map[string]any{
			"name":       sender.templates[message.Template()],
			"language":   map[string]string{"code": sender.language},
			"components": components,
		},
	})
}

func (sender *Sender) post(ctx context.Context, payload []byte) (providerRef string, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", false, fmt.Errorf("meta whatsapp: build request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+sender.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return "", true, fmt.Errorf("meta whatsapp: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retryable = response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		return "", retryable, fmt.Errorf("meta whatsapp: %w: %s", ErrDeliveryFailed, describe(response.StatusCode, body, readErr))
	}
	if readErr != nil {
		// The send was accepted; only the receipt is unreadable. Report
		// success with an empty reference rather than resending and
		// double-charging the member a second message.
		return "", false, nil
	}

	var receipt struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &receipt) == nil && len(receipt.Messages) > 0 {
		return receipt.Messages[0].ID, false, nil
	}
	return "", false, nil
}

// describe renders a safe diagnostic: HTTP status plus Meta's numeric error
// code and subcode. The provider's human-readable message is not echoed
// because it can quote template parameters.
func describe(status int, body []byte, readErr error) string {
	rendered := "status " + strconv.Itoa(status)
	if readErr != nil || len(body) == 0 {
		return rendered
	}
	var payload struct {
		Error struct {
			Code      int `json:"code"`
			Subcode   int `json:"error_subcode"`
			ErrorData struct {
				Details string `json:"details"`
			} `json:"error_data"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) != nil || payload.Error.Code == 0 {
		return rendered
	}
	rendered += ", provider code " + strconv.Itoa(payload.Error.Code)
	if payload.Error.Subcode != 0 {
		rendered += "/" + strconv.Itoa(payload.Error.Subcode)
	}
	return rendered
}
