package resend

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

type capturedRequest struct {
	Authorization string
	From          string   `json:"from"`
	To            []string `json:"to"`
	Subject       string   `json:"subject"`
	HTML          string   `json:"html"`
	ReplyTo       string   `json:"reply_to"`
}

func newTestSender(t *testing.T, config Config, respond http.HandlerFunc) (*Sender, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Authorization = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, captured)
		respond(w, r)
	}))
	t.Cleanup(server.Close)

	config.BaseURL = server.URL
	if config.APIKey == "" {
		config.APIKey = "re_synthetic_test_only"
	}
	if config.From == "" {
		config.From = "Obiara <no-reply@obiara.test>"
	}
	sender, err := NewSender(config, server.Client())
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	return sender, captured
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"id":"6229f547-f3f7-4f3a-9b1a-2b0f9d4b0d2f"}`))
}

// TestSendAdminNoticeCarriesTheCode covers admin sign-in, which is the only
// way into the console: if this template stops carrying the code, every
// operator is locked out.
func TestSendAdminNoticeCarriesTheCode(t *testing.T) {
	sender, captured := newTestSender(t, Config{ReplyTo: "support@obiara.test"}, ok)

	message, err := domain.NewMessage("operator@obiara.test", domain.TemplateAdminNotice,
		map[string]string{"code": "550142"})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	providerRef, err := sender.Send(context.Background(), message)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if providerRef != "6229f547-f3f7-4f3a-9b1a-2b0f9d4b0d2f" {
		t.Errorf("providerRef = %q; the delivery webhook correlates on it", providerRef)
	}
	if captured.Authorization != "Bearer re_synthetic_test_only" {
		t.Errorf("Authorization = %q", captured.Authorization)
	}
	if len(captured.To) != 1 || captured.To[0] != "operator@obiara.test" {
		t.Errorf("to = %v", captured.To)
	}
	if captured.From != "Obiara <no-reply@obiara.test>" {
		t.Errorf("from = %q", captured.From)
	}
	if captured.ReplyTo != "support@obiara.test" {
		t.Errorf("reply_to = %q", captured.ReplyTo)
	}
	if !strings.Contains(captured.HTML, "550142") {
		t.Errorf("body %q does not carry the code", captured.HTML)
	}
	if captured.Subject == "" {
		t.Error("subject is empty")
	}
}

// TestParametersAreHTMLEscaped stops a template parameter from injecting
// markup into the rendered email.
func TestParametersAreHTMLEscaped(t *testing.T) {
	sender, captured := newTestSender(t, Config{}, ok)

	message, err := domain.NewMessage("ops@obiara.test", domain.TemplateOpsAlert,
		map[string]string{"summary": `<script>alert(1)</script>`})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if _, err := sender.Send(context.Background(), message); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if strings.Contains(captured.HTML, "<script>") {
		t.Errorf("parameter markup survived into the body: %q", captured.HTML)
	}
	if !strings.Contains(captured.HTML, "&lt;script&gt;") {
		t.Errorf("parameter was not escaped: %q", captured.HTML)
	}
}

func TestEveryDomainTemplateRenders(t *testing.T) {
	cases := map[domain.Template]map[string]string{
		domain.TemplateAdminNotice:      {"code": "112233"},
		domain.TemplateOpsAlert:         {"summary": "queue depth exceeded"},
		domain.TemplateVerificationHelp: {"reference": "vrf_123"},
	}
	// A template the domain admits but this adapter cannot render would send
	// a blank email, so the mapping must stay exhaustive.
	if len(cases) != len(specs) {
		t.Fatalf("specs covers %d templates, test covers %d", len(specs), len(cases))
	}

	for template, params := range cases {
		t.Run(string(template), func(t *testing.T) {
			sender, captured := newTestSender(t, Config{}, ok)
			message, err := domain.NewMessage("someone@obiara.test", template, params)
			if err != nil {
				t.Fatalf("NewMessage: %v", err)
			}
			if _, err := sender.Send(context.Background(), message); err != nil {
				t.Fatalf("Send: %v", err)
			}
			if captured.Subject == "" || captured.HTML == "" {
				t.Errorf("template rendered empty: subject=%q html=%q", captured.Subject, captured.HTML)
			}
		})
	}
}

func TestSendReportsProviderRejection(t *testing.T) {
	sender, _ := newTestSender(t, Config{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"name":"validation_error","message":"code 550142 rejected"}`))
	})

	message, _ := domain.NewMessage("operator@obiara.test", domain.TemplateAdminNotice,
		map[string]string{"code": "550142"})
	_, err := sender.Send(context.Background(), message)
	if err == nil {
		t.Fatal("Send accepted a 422 response")
	}
	if !errors.Is(err, ErrDeliveryFailed) {
		t.Errorf("error %v should wrap ErrDeliveryFailed", err)
	}
	if strings.Contains(err.Error(), "550142") {
		t.Errorf("error leaked the MFA code: %v", err)
	}
}

func TestNewSenderFailsClosed(t *testing.T) {
	cases := map[string]Config{
		"no api key": {From: "a@b.test"},
		"no from":    {APIKey: "re_x"},
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSender(config, nil); !errors.Is(err, ErrNotConfigured) {
				t.Fatalf("NewSender error = %v, want ErrNotConfigured", err)
			}
		})
	}
}
