// Package mtn asks MTN Mobile Money to collect from a member.
//
// It implements the one method the payment context needs — "prompt this
// number for this much" — and nothing else. It does not decide whether the
// money arrived: that is the provider's signed callback, which the payment
// service verifies itself. An adapter that reported success here would be
// reporting that a prompt was sent, and a member who never approves the
// prompt has not paid.
package mtn

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
)

var (
	// ErrUnavailable covers every way the provider fails to accept a request.
	// The payment context turns it into a refusal; it never turns it into a
	// payment that might have happened.
	ErrUnavailable = errors.New("mobile money provider unavailable")
	ErrConfigured  = errors.New("mobile money provider is not configured")
)

// Config is what the provider issued for this deployment.
type Config struct {
	BaseURL           string
	SubscriptionKey   string
	APIUser           string
	APIKey            string
	TargetEnvironment string
	CallbackURL       string
}

// Provider is the collection client.
type Provider struct {
	config Config
	client *http.Client
	now    func() time.Time

	// token is cached because MTN's access tokens are short-lived and minting
	// one per collection would double every payment's latency and its failure
	// surface. The mutex is here because a member paying is not a serialised
	// event.
	mu        sync.Mutex
	token     string
	tokenTill time.Time
}

func New(config Config, now func() time.Time) (*Provider, error) {
	if now == nil {
		now = time.Now
	}
	if config.BaseURL == "" || config.SubscriptionKey == "" ||
		config.APIUser == "" || config.APIKey == "" || config.TargetEnvironment == "" {
		return nil, ErrConfigured
	}
	return &Provider{
		config: config,
		client: &http.Client{Timeout: 20 * time.Second},
		now:    now,
	}, nil
}

// RequestCollection asks the provider to prompt the member's phone.
//
// It returns the reference it was given, which is what the payment context
// checks: a provider that answered with a different reference is not talking
// about this collection, and treating that as success would attach somebody
// else's payment to this intent.
func (provider *Provider) RequestCollection(
	ctx context.Context, request application.ProviderRequest,
) (string, error) {
	if provider == nil || provider.client == nil {
		return "", ErrUnavailable
	}
	if strings.TrimSpace(request.RequestRef) == "" ||
		strings.TrimSpace(request.PhoneRef) == "" ||
		request.AmountPesewas == 0 || request.Currency != "GHS" {
		return "", ErrUnavailable
	}
	token, err := provider.accessToken(ctx)
	if err != nil {
		return "", ErrUnavailable
	}
	// MTN takes major units as a decimal string; the rest of this codebase
	// counts pesewas, so the conversion happens here and only here.
	body, err := json.Marshal(map[string]any{
		"amount":       majorUnits(request.AmountPesewas),
		"currency":     request.Currency,
		"externalId":   request.RequestRef,
		"payer":        map[string]string{"partyIdType": "MSISDN", "partyId": request.PhoneRef},
		"payerMessage": "Obiara membership",
		"payeeNote":    "Obiara membership",
	})
	if err != nil {
		return "", ErrUnavailable
	}
	call, err := http.NewRequestWithContext(
		ctx, http.MethodPost, provider.config.BaseURL+"/collection/v1_0/requesttopay",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", ErrUnavailable
	}
	call.Header.Set("Authorization", "Bearer "+token)
	call.Header.Set("X-Reference-Id", request.RequestRef)
	call.Header.Set("X-Target-Environment", provider.config.TargetEnvironment)
	call.Header.Set("Ocp-Apim-Subscription-Key", provider.config.SubscriptionKey)
	call.Header.Set("Content-Type", "application/json")
	if provider.config.CallbackURL != "" {
		call.Header.Set("X-Callback-Url", provider.config.CallbackURL)
	}

	response, err := provider.client.Do(call)
	if err != nil {
		return "", ErrUnavailable
	}
	defer drain(response)
	// 202 Accepted is the only success: the prompt is on its way. Anything
	// else, including a 200, is not a prompt this adapter can vouch for.
	if response.StatusCode != http.StatusAccepted {
		return "", ErrUnavailable
	}
	return request.RequestRef, nil
}

// accessToken returns a live token, minting one when the cached one is spent.
//
// Refreshed a minute early on purpose: a token that expires between this
// check and the request it authorizes fails a member's payment for no reason
// they could understand.
func (provider *Provider) accessToken(ctx context.Context) (string, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if provider.token != "" && provider.now().UTC().Before(provider.tokenTill) {
		return provider.token, nil
	}
	call, err := http.NewRequestWithContext(
		ctx, http.MethodPost, provider.config.BaseURL+"/collection/token/", nil)
	if err != nil {
		return "", ErrUnavailable
	}
	credential := base64.StdEncoding.EncodeToString(
		[]byte(provider.config.APIUser + ":" + provider.config.APIKey))
	call.Header.Set("Authorization", "Basic "+credential)
	call.Header.Set("Ocp-Apim-Subscription-Key", provider.config.SubscriptionKey)

	response, err := provider.client.Do(call)
	if err != nil {
		return "", ErrUnavailable
	}
	defer drain(response)
	if response.StatusCode != http.StatusOK {
		return "", ErrUnavailable
	}
	var minted struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<16)).Decode(&minted); err != nil {
		return "", ErrUnavailable
	}
	if strings.TrimSpace(minted.AccessToken) == "" || minted.ExpiresIn <= 60 {
		return "", ErrUnavailable
	}
	provider.token = minted.AccessToken
	provider.tokenTill = provider.now().UTC().Add(time.Duration(minted.ExpiresIn-60) * time.Second)
	return provider.token, nil
}

// majorUnits renders pesewas as the decimal string MTN expects. Done with
// integers rather than floats because money and floating point do not mix.
func majorUnits(pesewas uint64) string {
	return fmt.Sprintf("%d.%02d", pesewas/100, pesewas%100)
}

func drain(response *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<16))
	_ = response.Body.Close()
}
