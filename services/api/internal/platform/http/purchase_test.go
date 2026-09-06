package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	momoapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
)

type purchaseStub struct {
	started   purchase.Started
	startErr  error
	settleErr error
	command   purchase.StartCommand
	callback  momoapplication.Callback
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

func (s *purchaseStub) Settle(_ context.Context, callback momoapplication.Callback) error {
	s.settles++
	s.callback = callback
	return s.settleErr
}

func purchaseCall(t *testing.T, stub *purchaseStub, path, body, key string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterPurchaseRoutes(mux, stub, sessionStub{memberID: "member_1"})
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
		`{"skuId":"sku_membership","skuVersion":1,"phone":"0200000000"}`, "cmd-1")

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
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000"}`, "")

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
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000"}`, "cmd-1")

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
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000"}`, "cmd-1")

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "payment_unavailable") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestACallbackCarriesItsSignatureThrough(t *testing.T) {
	// The callback has no session: the HMAC over the payload is the whole
	// authentication, so every field it is computed over has to arrive.
	stub := &purchaseStub{}
	response := purchaseCall(t, stub, "/v1/payments/momo/callback",
		`{"callbackId":"cb_1","intentId":"intent_1","providerRef":"ref-1",`+
			`"success":true,"occurredAt":1757160000,"signature":"abc"}`, "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", response.Code, response.Body.String())
	}
	if stub.callback.Signature != "abc" || stub.callback.OccurredUnix != 1757160000 {
		t.Fatalf("callback = %#v", stub.callback)
	}
	if !stub.callback.Success || stub.callback.CallbackID != "cb_1" {
		t.Fatalf("callback = %#v", stub.callback)
	}
}

func TestARepeatedCallbackIsAnsweredAsDone(t *testing.T) {
	// Providers retry. The payment context is idempotent by callback id, so
	// an error on the second one would only make them retry harder.
	stub := &purchaseStub{settleErr: purchase.ErrAlreadySettled}
	response := purchaseCall(t, stub, "/v1/payments/momo/callback",
		`{"callbackId":"cb_1","intentId":"intent_1","providerRef":"r","success":true,`+
			`"occurredAt":1,"signature":"abc"}`, "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", response.Code, response.Body.String())
	}
}

func TestARejectedCallbackSaysNothingAboutWhy(t *testing.T) {
	// A bad signature and an unknown intent answer identically, or this would
	// be a way to probe which intents exist.
	stub := &purchaseStub{settleErr: errors.New("bad signature")}
	response := purchaseCall(t, stub, "/v1/payments/momo/callback",
		`{"callbackId":"cb_1","intentId":"intent_1","providerRef":"r","success":true,`+
			`"occurredAt":1,"signature":"wrong"}`, "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", response.Code, response.Body.String())
	}
	body := strings.ToLower(response.Body.String())
	for _, leak := range []string{"signature", "intent", "not found", "unknown"} {
		if strings.Contains(body, leak) {
			t.Fatalf("the refusal said why (%q): %s", leak, response.Body.String())
		}
	}
}

func TestTheCallbackTakesNoSession(t *testing.T) {
	// The caller is a provider. Requiring a member session would make the
	// route impossible to call, and accepting one would suggest the session
	// mattered when the HMAC is what is checked.
	stub := &purchaseStub{}
	mux := http.NewServeMux()
	RegisterPurchaseRoutes(mux, stub, sessionStub{})
	request := httptest.NewRequest(http.MethodPost, "/v1/payments/momo/callback",
		strings.NewReader(`{"callbackId":"cb_1","intentId":"i","providerRef":"r",`+
			`"success":true,"occurredAt":1,"signature":"abc"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", response.Code, response.Body.String())
	}
	if stub.settles != 1 {
		t.Fatal("an unauthenticated provider callback was refused")
	}
}
