package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
)

type webhookStub struct {
	verifyErr error
	applied   map[string]string
}

func (stub *webhookStub) VerifySignature(_, _, _ string, _ []byte) error {
	return stub.verifyErr
}

func (stub *webhookStub) ApplyStatus(_ context.Context, providerRef, status string) error {
	if stub.applied == nil {
		stub.applied = map[string]string{}
	}
	stub.applied[providerRef] = status
	return nil
}

func postWebhook(t *testing.T, handler http.Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/webhooks/resend", strings.NewReader(body))
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	handler := resendWebhookHandler(&webhookStub{verifyErr: application.ErrSignatureInvalid}, nil)
	response := postWebhook(t, handler, `{"type":"email.delivered","data":{"email_id":"ref-1"}}`, map[string]string{
		"svix-id": "msg_1", "svix-timestamp": "1", "svix-signature": "v1,bad",
	})
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestWebhookAppliesKnownEvent(t *testing.T) {
	stub := &webhookStub{}
	handler := resendWebhookHandler(stub, nil)
	response := postWebhook(t, handler, `{"type":"email.delivered","data":{"email_id":"ref-1"}}`, map[string]string{
		"svix-id": "msg_1", "svix-timestamp": "1", "svix-signature": "v1,ok",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if stub.applied["ref-1"] != "delivered" {
		t.Fatalf("applied = %#v", stub.applied)
	}
}

func TestWebhookIgnoresUnknownEvent(t *testing.T) {
	stub := &webhookStub{}
	handler := resendWebhookHandler(stub, nil)
	response := postWebhook(t, handler, `{"type":"email.opened","data":{"email_id":"ref-1"}}`, map[string]string{
		"svix-id": "msg_1", "svix-timestamp": "1", "svix-signature": "v1,ok",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if len(stub.applied) != 0 {
		t.Fatal("unknown event must not apply")
	}
	if !strings.Contains(response.Body.String(), "ignored") {
		t.Fatalf("body = %s, want ignored", response.Body.String())
	}
}

func TestWebhookRejectsMalformedPayload(t *testing.T) {
	stub := &webhookStub{}
	handler := resendWebhookHandler(stub, nil)
	response := postWebhook(t, handler, `{"type":`, map[string]string{
		"svix-id": "msg_1", "svix-timestamp": "1", "svix-signature": "v1,ok",
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
