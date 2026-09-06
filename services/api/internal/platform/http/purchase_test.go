package apihttp

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
)

type purchaseStub struct {
	started   purchase.Started
	startErr  error
	settleErr error
	command   purchase.StartCommand
	outcome   purchase.Outcome
	starts    int
	settles   int
}

func (s *purchaseStub) Start(
	_ context.Context, command purchase.StartCommand,
) (purchase.Started, error) {
	s.starts++
	s.command = command
	return s.started, s.startErr
}

func (s *purchaseStub) Settle(_ context.Context, outcome purchase.Outcome) error {
	s.settles++
	s.outcome = outcome
	return s.settleErr
}

func purchaseCall(t *testing.T, stub *purchaseStub, path, body, key string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterPurchaseRoutes(mux, stub, sessionStub{memberID: "member_1"}, webhookSecret)
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

func TestBuyingAMembershipIsTheSessionsOwnPurchase(t *testing.T) {
	// The pass is granted to whoever paid. Nobody buys one for somebody else
	// here, so the buyer comes from the session and not from the body.
	stub := &purchaseStub{started: purchase.Started{
		IntentID: "intent_1", Status: "awaiting_member_confirmation", AmountPesewas: 5000,
	}}
	response := purchaseCall(t, stub, "/v1/membership/purchases",
		`{"skuId":"sku_membership","skuVersion":1,"phone":"0200000000","network":"mtn"}`, "cmd-1")

	// 202: the prompt is on its way and nothing has been paid.
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", response.Code, response.Body.String())
	}
	if stub.command.MemberID != "member_1" {
		t.Fatalf("buyer = %q", stub.command.MemberID)
	}
	if stub.command.CommandID != "cmd-1" {
		t.Fatalf("command id = %q, want the idempotency key", stub.command.CommandID)
	}
	// The amount is echoed so a client shows the real number rather than one
	// it assumed.
	if !strings.Contains(response.Body.String(), "5000") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAPurchaseWithoutARetryKeyNeverStarts(t *testing.T) {
	// Without one a double submission opens two collections, and a member is
	// prompted twice for the same membership.
	stub := &purchaseStub{}
	response := purchaseCall(t, stub, "/v1/membership/purchases",
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000","network":"mtn"}`, "")

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	if stub.starts != 0 {
		t.Fatal("a purchase with no retry key reached the service")
	}
}

func TestSomethingNobodyCanBuyIsRefusedPlainly(t *testing.T) {
	stub := &purchaseStub{startErr: purchase.ErrNotPurchasable}
	response := purchaseCall(t, stub, "/v1/membership/purchases",
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000","network":"mtn"}`, "cmd-1")

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "not_purchasable") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAnUnreachableProviderSaysToTryAgainRatherThanFailing(t *testing.T) {
	// A member whose payment could not be started has not lost anything, and
	// the message should say so rather than read as a fault in their account.
	stub := &purchaseStub{startErr: purchase.ErrUnavailable}
	response := purchaseCall(t, stub, "/v1/membership/purchases",
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000","network":"mtn"}`, "cmd-1")

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "payment_unavailable") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

const webhookSecret = "sk_test_0123456789abcdef0123456789abcdef"

func paystackSigned(body string) string {
	mac := hmac.New(sha512.New, []byte(webhookSecret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

func webhookCall(t *testing.T, stub *purchaseStub, body, signature string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterPurchaseRoutes(mux, stub, sessionStub{}, webhookSecret)
	request := httptest.NewRequest(
		http.MethodPost, "/v1/payments/paystack/webhook", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if signature != "" {
		request.Header.Set("x-paystack-signature", signature)
	}
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

const chargeSuccess = `{"event":"charge.success","data":{"reference":"intent-1",` +
	`"status":"success","amount":5000,"currency":"GHS"}}`

func TestAVerifiedWebhookSettlesThePayment(t *testing.T) {
	// The webhook carries no session: the HMAC over the exact bytes is the
	// whole authentication, because the caller is a processor and not a
	// member.
	stub := &purchaseStub{}
	response := webhookCall(t, stub, chargeSuccess, paystackSigned(chargeSuccess))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if stub.settles != 1 {
		t.Fatal("a verified webhook did not settle")
	}
	if !stub.outcome.Success || stub.outcome.Reference != "intent-1" {
		t.Fatalf("outcome = %#v", stub.outcome)
	}
	// The amount and currency travel, because settlement checks them against
	// what the collection was opened for.
	if stub.outcome.AmountPesewas != 5000 || stub.outcome.Currency != "GHS" {
		t.Fatalf("outcome = %#v", stub.outcome)
	}
}

func TestAWebhookIsVerifiedAgainstTheBytesThatArrived(t *testing.T) {
	// Not against a re-serialisation of the decoded body. A single altered
	// byte must not verify, and this is the only thing between a stranger and
	// a free membership.
	tampered := strings.Replace(chargeSuccess, `"amount":5000`, `"amount":1`, 1)
	stub := &purchaseStub{}
	response := webhookCall(t, stub, tampered, paystackSigned(chargeSuccess))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if stub.settles != 0 {
		t.Fatal("a tampered webhook settled a payment")
	}
}

func TestAnUnsignedWebhookSettlesNothing(t *testing.T) {
	for name, signature := range map[string]string{
		"missing": "", "wrong": strings.Repeat("0", 128), "not hex": "nonsense",
	} {
		t.Run(name, func(t *testing.T) {
			stub := &purchaseStub{}
			response := webhookCall(t, stub, chargeSuccess, signature)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", response.Code)
			}
			if stub.settles != 0 {
				t.Fatal("an unsigned webhook settled a payment")
			}
		})
	}
}

func TestARejectedWebhookSaysNothingAboutWhy(t *testing.T) {
	// A bad signature and an unreadable body answer identically, or this
	// tells somebody probing which of the two they got wrong.
	stub := &purchaseStub{}
	response := webhookCall(t, stub, chargeSuccess, strings.Repeat("0", 128))
	body := strings.ToLower(response.Body.String())
	for _, leak := range []string{"signature", "hmac", "secret", "intent"} {
		if strings.Contains(body, leak) {
			t.Fatalf("the refusal said why (%q): %s", leak, response.Body.String())
		}
	}
}

func TestAnEventThisEndpointDoesNotOwnIsAcknowledged(t *testing.T) {
	// Paystack sends transfers, invoices and subscriptions to the same URL.
	// Answering them 200 without acting is correct: they are not this
	// endpoint's business, and an error would make Paystack retry forever.
	other := `{"event":"transfer.success","data":{"reference":"x","status":"success",` +
		`"amount":1,"currency":"GHS"}}`
	stub := &purchaseStub{}
	response := webhookCall(t, stub, other, paystackSigned(other))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if stub.settles != 0 {
		t.Fatal("a transfer was settled as a member's payment")
	}
}

func TestARepeatedWebhookIsAnsweredAsDone(t *testing.T) {
	// Processors retry. The payment context is idempotent by callback id, so
	// an error on the second one would only make them retry harder.
	stub := &purchaseStub{settleErr: purchase.ErrAlreadySettled}
	response := webhookCall(t, stub, chargeSuccess, paystackSigned(chargeSuccess))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
}

func TestTheSameOutcomeAlwaysCarriesTheSameCallbackID(t *testing.T) {
	// The id is what makes a retry settle once, so it has to be derived from
	// the outcome rather than from anything that changes between deliveries.
	first, second := &purchaseStub{}, &purchaseStub{}
	webhookCall(t, first, chargeSuccess, paystackSigned(chargeSuccess))
	webhookCall(t, second, chargeSuccess, paystackSigned(chargeSuccess))
	if first.outcome.CallbackID == "" || first.outcome.CallbackID != second.outcome.CallbackID {
		t.Fatalf("%q then %q", first.outcome.CallbackID, second.outcome.CallbackID)
	}
}

func TestAnAmountMismatchStopsPaystackRetryingSomethingUnfixable(t *testing.T) {
	// Authentic and wrong. Retrying will never fix it, so it is acknowledged
	// and logged rather than answered with an error.
	stub := &purchaseStub{settleErr: purchase.ErrWrongAmount}
	response := webhookCall(t, stub, chargeSuccess, paystackSigned(chargeSuccess))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}

func TestAFailureOnThisSideAsksPaystackToTryAgain(t *testing.T) {
	stub := &purchaseStub{settleErr: errors.New("mongo down")}
	response := webhookCall(t, stub, chargeSuccess, paystackSigned(chargeSuccess))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", response.Code, response.Body.String())
	}
}
