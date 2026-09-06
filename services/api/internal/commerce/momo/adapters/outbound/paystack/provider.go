// Package paystack charges a member through Paystack.
//
// It does two separable things, and they are separate on purpose:
//
//   - RequestCollection asks Paystack to prompt a member's phone. It reports
//     only that a prompt was accepted. It never reports that money arrived,
//     because a member who does not approve the prompt has not paid.
//   - VerifyWebhook authenticates what Paystack sends back. That is the only
//     thing in this product that may say a payment happened.
package paystack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
)

var (
	// ErrUnavailable covers every way the charge fails to be accepted. The
	// payment context turns it into a refusal; it never turns it into a
	// payment that might have happened.
	ErrUnavailable = errors.New("paystack unavailable")
	ErrConfigured  = errors.New("paystack is not configured")
	// ErrBadSignature refuses a webhook whose HMAC does not match the bytes
	// received. It is the only thing standing between a stranger and a free
	// membership, so it is never softened.
	ErrBadSignature = errors.New("paystack webhook signature does not match")
	ErrBadEvent     = errors.New("paystack webhook could not be read")
)

// SignatureHeader is the header Paystack signs its webhooks with.
const SignatureHeader = "x-paystack-signature"

// maxWebhookBytes bounds what is read before verifying anything. A signature
// check on an unbounded body is a way to make the server read forever.
const maxWebhookBytes = 1 << 20

// Config is what Paystack issued for this deployment.
type Config struct {
	BaseURL string
	// SecretKey is both the API bearer token and the key Paystack signs
	// webhooks with.
	SecretKey   string
	CallbackURL string
}

type Provider struct {
	config Config
	client *http.Client
}

func New(config Config) (*Provider, error) {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if config.BaseURL == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, ErrConfigured
	}
	return &Provider{config: config, client: &http.Client{Timeout: 20 * time.Second}}, nil
}

// Networks are the mobile money providers Paystack accepts in Ghana.
//
// A closed list rather than passing through whatever a client sends: an
// unrecognised network is refused here instead of becoming a charge Paystack
// rejects after the member has already been told something is happening.
var Networks = map[string]string{
	"mtn":        "mtn",
	"vodafone":   "vod",
	"telecel":    "vod",
	"airteltigo": "atl",
	"at":         "atl",
}

// Network maps a network somebody named to the code Paystack expects.
func Network(name string) (string, bool) {
	code, ok := Networks[strings.ToLower(strings.TrimSpace(name))]
	return code, ok
}

type chargeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Reference   string `json:"reference"`
		Status      string `json:"status"`
		DisplayText string `json:"display_text"`
	} `json:"data"`
}

// RequestCollection asks Paystack to prompt the member's phone.
//
// The reference sent is the one the payment context gave, so the webhook that
// comes back names a collection this product already knows about. Paystack
// requires references to be unique and alphanumeric with hyphens, periods or
// equals signs, which an opaque 64-hex id satisfies.
func (provider *Provider) RequestCollection(
	ctx context.Context, request application.ProviderRequest,
) (string, error) {
	if provider == nil || provider.client == nil {
		return "", ErrUnavailable
	}
	reference := strings.TrimSpace(request.RequestRef)
	network, known := Network(request.Network)
	if reference == "" || strings.TrimSpace(request.Phone) == "" ||
		strings.TrimSpace(request.Email) == "" || !known ||
		request.AmountPesewas == 0 || request.Currency != "GHS" {
		return "", ErrUnavailable
	}
	// Amounts go in the subunit — pesewas for cedis — which is what the rest
	// of this codebase counts in, so there is nothing to convert and no
	// rounding to get wrong.
	body, err := json.Marshal(map[string]any{
		"amount":       request.AmountPesewas,
		"email":        request.Email,
		"currency":     request.Currency,
		"reference":    reference,
		"mobile_money": map[string]string{"phone": request.Phone, "provider": network},
	})
	if err != nil {
		return "", ErrUnavailable
	}
	call, err := http.NewRequestWithContext(
		ctx, http.MethodPost, provider.config.BaseURL+"/charge", bytes.NewReader(body))
	if err != nil {
		return "", ErrUnavailable
	}
	call.Header.Set("Authorization", "Bearer "+provider.config.SecretKey)
	call.Header.Set("Content-Type", "application/json")

	response, err := provider.client.Do(call)
	if err != nil {
		return "", ErrUnavailable
	}
	defer drain(response)
	if response.StatusCode != http.StatusOK {
		return "", ErrUnavailable
	}
	var charged chargeResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxWebhookBytes)).Decode(&charged); err != nil {
		return "", ErrUnavailable
	}
	if !charged.Status {
		return "", ErrUnavailable
	}
	// pay_offline is a prompt waiting on the member's phone. success is an
	// instant charge. Anything else — send_otp, send_pin — is a card flow
	// this product does not offer, and pretending it was a sent prompt would
	// leave a member waiting for something that will never arrive.
	if charged.Data.Status != "pay_offline" && charged.Data.Status != "success" {
		return "", ErrUnavailable
	}
	// The reference Paystack echoes, checked against the one sent. A provider
	// answering about a different collection is not an answer about this one.
	if charged.Data.Reference != reference {
		return "", ErrUnavailable
	}
	return reference, nil
}

// Event is what a verified webhook said.
type Event struct {
	Name string
	// Reference is the reference this product gave Paystack, so it names a
	// collection already on record.
	Reference     string
	Status        string
	AmountPesewas int64
	Currency      string
}

// Succeeded reports whether this event says money arrived.
//
// Both the event name and the transaction status, because Paystack sends
// charge.success for the event and carries the outcome in data.status; taking
// either alone would accept an event that is not the one it looks like.
func (event Event) Succeeded() bool {
	return event.Name == "charge.success" && event.Status == "success"
}

type webhookBody struct {
	Event string `json:"event"`
	Data  struct {
		Reference string `json:"reference"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
	} `json:"data"`
}

// VerifyWebhook authenticates a webhook and reads what it says.
//
// The signature is an HMAC-SHA512 of the exact bytes received, keyed with the
// secret key, hex encoded, in the x-paystack-signature header. So the bytes
// have to be verified before they are parsed — re-serialising the decoded
// body and hashing that would hash something Paystack never sent, and would
// break the moment Paystack added a field or changed key order.
//
// Unknown fields are tolerated for the same reason: a webhook that refused a
// body carrying a field this build has not heard of would stop settling
// payments the first time Paystack extended its event.
func VerifyWebhook(secret string, body []byte, signature string) (Event, error) {
	if strings.TrimSpace(secret) == "" {
		return Event{}, ErrConfigured
	}
	expected := hmac.New(sha512.New, []byte(secret))
	expected.Write(body)
	// Constant time, so a wrong signature cannot be found a byte at a time.
	if !hmac.Equal(
		[]byte(strings.ToLower(strings.TrimSpace(signature))),
		[]byte(hex.EncodeToString(expected.Sum(nil))),
	) {
		return Event{}, ErrBadSignature
	}
	var read webhookBody
	if err := json.Unmarshal(body, &read); err != nil {
		return Event{}, ErrBadEvent
	}
	if strings.TrimSpace(read.Event) == "" || strings.TrimSpace(read.Data.Reference) == "" {
		return Event{}, ErrBadEvent
	}
	return Event{
		Name:          read.Event,
		Reference:     strings.TrimSpace(read.Data.Reference),
		Status:        read.Data.Status,
		AmountPesewas: read.Data.Amount,
		Currency:      read.Data.Currency,
	}, nil
}

func drain(response *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<16))
	_ = response.Body.Close()
}

// Transfers is money going out.
//
// Separated from collection because it is a different act with a different
// risk: a mistake here sends real money to somebody, and there is no member
// standing in front of it to notice. Everything about it is deliberately
// two-step — a recipient is created, then a transfer is initiated against it —
// so a wrong number fails at the recipient stage rather than after the money
// has gone.

type recipientResponse struct {
	Status bool `json:"status"`
	Data   struct {
		RecipientCode string `json:"recipient_code"`
	} `json:"data"`
}

// CreateRecipient registers who is being paid and returns Paystack's code for
// them.
//
// The telco code goes in as bank_code, which is how Paystack models a mobile
// money destination in Ghana.
func (provider *Provider) CreateRecipient(
	ctx context.Context, name, phone, network string,
) (string, error) {
	if provider == nil || provider.client == nil {
		return "", ErrUnavailable
	}
	code, known := Network(network)
	if strings.TrimSpace(name) == "" || strings.TrimSpace(phone) == "" || !known {
		return "", ErrUnavailable
	}
	body, err := json.Marshal(map[string]any{
		"type": "mobile_money", "name": name, "account_number": phone,
		"bank_code": strings.ToUpper(code), "currency": "GHS",
	})
	if err != nil {
		return "", ErrUnavailable
	}
	var created recipientResponse
	if err := provider.post(ctx, "/transferrecipient", body, &created); err != nil {
		return "", err
	}
	if !created.Status || strings.TrimSpace(created.Data.RecipientCode) == "" {
		return "", ErrUnavailable
	}
	return created.Data.RecipientCode, nil
}

type transferResponse struct {
	Status bool `json:"status"`
	Data   struct {
		Reference    string `json:"reference"`
		TransferCode string `json:"transfer_code"`
		Status       string `json:"status"`
	} `json:"data"`
}

// Transfer sends money to a registered recipient.
//
// The reference is supplied so a retry cannot send twice: Paystack refuses a
// duplicate reference, which is the only thing between a retried payout and
// paying somebody the same amount again.
func (provider *Provider) Transfer(
	ctx context.Context, recipientCode, reference, reason string, amountPesewas int64,
) (string, error) {
	if provider == nil || provider.client == nil {
		return "", ErrUnavailable
	}
	if strings.TrimSpace(recipientCode) == "" || strings.TrimSpace(reference) == "" ||
		amountPesewas <= 0 {
		return "", ErrUnavailable
	}
	body, err := json.Marshal(map[string]any{
		"source": "balance", "amount": amountPesewas, "recipient": recipientCode,
		"reason": reason, "reference": reference, "currency": "GHS",
	})
	if err != nil {
		return "", ErrUnavailable
	}
	var sent transferResponse
	if err := provider.post(ctx, "/transfer", body, &sent); err != nil {
		return "", err
	}
	if !sent.Status || strings.TrimSpace(sent.Data.TransferCode) == "" {
		return "", ErrUnavailable
	}
	return sent.Data.TransferCode, nil
}

func (provider *Provider) post(ctx context.Context, path string, body []byte, into any) error {
	call, err := http.NewRequestWithContext(
		ctx, http.MethodPost, provider.config.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return ErrUnavailable
	}
	call.Header.Set("Authorization", "Bearer "+provider.config.SecretKey)
	call.Header.Set("Content-Type", "application/json")
	response, err := provider.client.Do(call)
	if err != nil {
		return ErrUnavailable
	}
	defer drain(response)
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return ErrUnavailable
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxWebhookBytes)).Decode(into); err != nil {
		return ErrUnavailable
	}
	return nil
}
