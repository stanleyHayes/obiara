// Package expo is the production push adapter, backed by the Expo push
// service (exp.host). It implements the push channel's application.Sender
// port.
//
// Expo answers a batch with HTTP 200 and a per-message status array, so a
// request that "succeeded" can still contain wholly failed messages. Treating
// 2xx as delivery would repeat the class of bug that let OTP codes vanish, so
// the per-ticket statuses are authoritative here.
package expo

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

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

var (
	// ErrNotConfigured reports an unusable configuration.
	ErrNotConfigured = errors.New("expo push sender is not configured")
	// ErrDeliveryFailed reports a provider-side rejection.
	ErrDeliveryFailed = errors.New("expo push delivery failed")
)

const (
	defaultBaseURL   = "https://exp.host"
	maxResponseBytes = 1 << 20
	defaultTimeout   = 10 * time.Second
	maxAttempts      = 2
	retryBackoff     = 250 * time.Millisecond
	// batchSize is Expo's documented maximum messages per request.
	batchSize = 100
)

// Config carries Expo credentials and routing.
type Config struct {
	// AccessToken is optional. It is required only when the Expo project has
	// enhanced security enabled, and is sent as a bearer token when set.
	AccessToken string
	// BaseURL overrides the host in tests.
	BaseURL string
}

// Sender delivers push notifications over Expo.
type Sender struct {
	client   *http.Client
	endpoint string
	token    string
}

func NewSender(config Config, client *http.Client) (*Sender, error) {
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
		endpoint: base + "/--/api/v2/push/send",
		token:    strings.TrimSpace(config.AccessToken),
	}, nil
}

type message struct {
	To       string            `json:"to"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Sound    string            `json:"sound,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Priority string            `json:"priority,omitempty"`
}

type ticket struct {
	Status  string `json:"status"`
	ID      string `json:"id"`
	Message string `json:"message"`
	Details struct {
		Error string `json:"error"`
	} `json:"details"`
}

// Send delivers copy to every token, in batches, and reports the tokens Expo
// considers permanently dead.
func (sender *Sender) Send(ctx context.Context, tokens []string, copy domain.Copy, reference string) ([]string, error) {
	var dead []string
	var failures []error

	for start := 0; start < len(tokens); start += batchSize {
		end := start + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		chunk := tokens[start:end]

		batchDead, err := sender.sendBatch(ctx, chunk, copy, reference)
		dead = append(dead, batchDead...)
		if err != nil {
			failures = append(failures, err)
		}
	}

	if len(failures) > 0 {
		return dead, errors.Join(failures...)
	}
	return dead, nil
}

func (sender *Sender) sendBatch(ctx context.Context, tokens []string, copy domain.Copy, reference string) ([]string, error) {
	messages := make([]message, 0, len(tokens))
	for _, token := range tokens {
		messages = append(messages, message{
			To: token, Title: copy.Title, Body: copy.Body,
			Sound: "default", Priority: "high",
			// The reference is opaque; the app resolves it after unlock. No
			// content rides the notification itself.
			Data: map[string]string{"ref": reference},
		})
	}
	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("expo: encode request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dead, retryable, err := sender.post(ctx, payload, tokens)
		if err == nil {
			return dead, nil
		}
		lastErr = err
		if !retryable || attempt == maxAttempts {
			return dead, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryBackoff):
		}
	}
	return nil, lastErr
}

func (sender *Sender) post(ctx context.Context, payload []byte, tokens []string) (dead []string, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("expo: build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if sender.token != "" {
		request.Header.Set("Authorization", "Bearer "+sender.token)
	}

	response, err := sender.client.Do(request)
	if err != nil {
		return nil, true, fmt.Errorf("expo: %w: transport", ErrDeliveryFailed)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		_ = response.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retryable = response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		return nil, retryable, fmt.Errorf("expo: %w: status %d", ErrDeliveryFailed, response.StatusCode)
	}
	if readErr != nil {
		// Accepted but the receipt is unreadable; resending would double
		// notify, so report success with nothing to prune.
		return nil, false, nil
	}

	var receipt struct {
		Data   []ticket `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &receipt); err != nil {
		return nil, false, fmt.Errorf("expo: %w: receipt was not json", ErrDeliveryFailed)
	}
	if len(receipt.Errors) > 0 {
		return nil, false, fmt.Errorf("expo: %w: request rejected", ErrDeliveryFailed)
	}

	// A 200 can still carry per-message failures. DeviceNotRegistered is the
	// one that must prune, or the dead token is retried forever and
	// eventually poisons whole batches.
	var failed int
	for index, entry := range receipt.Data {
		if strings.EqualFold(entry.Status, "ok") {
			continue
		}
		failed++
		if entry.Details.Error == "DeviceNotRegistered" && index < len(tokens) {
			dead = append(dead, tokens[index])
		}
	}
	if failed == len(receipt.Data) && failed > 0 {
		return dead, false, fmt.Errorf("expo: %w: all %s messages rejected",
			ErrDeliveryFailed, strconv.Itoa(failed))
	}
	return dead, false, nil
}
