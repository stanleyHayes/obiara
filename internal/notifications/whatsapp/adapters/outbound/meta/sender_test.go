package meta

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
)

type wireMessage struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Template         struct {
		Name     string `json:"name"`
		Language struct {
			Code string `json:"code"`
		} `json:"language"`
		Components []struct {
			Type       string `json:"type"`
			SubType    string `json:"sub_type"`
			Index      string `json:"index"`
			Parameters []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"parameters"`
		} `json:"components"`
	} `json:"template"`
}

func newTestSender(t *testing.T, config Config, respond http.HandlerFunc) (*Sender, *wireMessage, *string) {
	t.Helper()
	captured := &wireMessage{}
	path := new(string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, captured)
		respond(w, r)
	}))
	t.Cleanup(server.Close)

	config.BaseURL = server.URL
	if config.PhoneNumberID == "" {
		config.PhoneNumberID = "123456789"
	}
	if config.AccessToken == "" {
		config.AccessToken = "synthetic-test-only"
	}
	sender, err := NewSender(config, server.Client())
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	return sender, captured, path
}

func accepted(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.TEST"}]}`))
}

func TestOtpTemplateCarriesBodyAndCopyCodeButton(t *testing.T) {
	sender, captured, path := newTestSender(t, Config{APIVersion: "v21.0"}, accepted)

	message, err := domain.NewOtpMessage("+233544919953", "482913")
	if err != nil {
		t.Fatalf("NewOtpMessage: %v", err)
	}
	providerRef, err := sender.Send(context.Background(), message)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if providerRef != "wamid.TEST" {
		t.Errorf("providerRef = %q", providerRef)
	}
	if *path != "/v21.0/123456789/messages" {
		t.Errorf("path = %q", *path)
	}
	if captured.MessagingProduct != "whatsapp" || captured.Type != "template" {
		t.Errorf("wire shape = %+v", captured)
	}
	if captured.Template.Name != "otp_code" {
		t.Errorf("template name = %q", captured.Template.Name)
	}
	if len(captured.Template.Components) != 2 {
		t.Fatalf("components = %d, want body + copy-code button", len(captured.Template.Components))
	}
	body := captured.Template.Components[0]
	if body.Type != "body" || len(body.Parameters) != 1 || body.Parameters[0].Text != "482913" {
		t.Errorf("body component = %+v", body)
	}
	// Meta rejects authentication templates that omit the button copy.
	button := captured.Template.Components[1]
	if button.Type != "button" || button.SubType != "copy_code" || button.Parameters[0].Text != "482913" {
		t.Errorf("button component = %+v", button)
	}
}

func TestTemplateNamesCanBeRemapped(t *testing.T) {
	sender, captured, _ := newTestSender(t, Config{
		TemplateNames: map[domain.Template]string{domain.TemplateOtpCode: "obiara_otp_v2"},
		LanguageCode:  "en_US",
	}, accepted)

	message, _ := domain.NewOtpMessage("+233544919953", "482913")
	if _, err := sender.Send(context.Background(), message); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if captured.Template.Name != "obiara_otp_v2" {
		t.Errorf("template name = %q, want the WhatsApp Manager approval name", captured.Template.Name)
	}
	if captured.Template.Language.Code != "en_US" {
		t.Errorf("language = %q", captured.Template.Language.Code)
	}
}

func TestEveryDomainTemplateIsMapped(t *testing.T) {
	// A domain template with no spec here fails at send time, so the two
	// must stay in step.
	for _, template := range []domain.Template{
		domain.TemplateOtpCode, domain.TemplatePodAlert, domain.TemplateNnoboaConsent,
	} {
		if _, ok := specs[template]; !ok {
			t.Errorf("domain template %q has no meta component mapping", template)
		}
	}
	if len(specs) != 3 {
		t.Errorf("specs has %d entries, want 3", len(specs))
	}
}

func TestSendReportsProviderRejectionWithoutLeakingTheCode(t *testing.T) {
	sender, _, _ := newTestSender(t, Config{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"template param 482913 invalid","code":132000,"error_subcode":2494008}}`))
	})

	message, _ := domain.NewOtpMessage("+233544919953", "482913")
	_, err := sender.Send(context.Background(), message)
	if err == nil {
		t.Fatal("Send accepted a 400 response")
	}
	if !errors.Is(err, ErrDeliveryFailed) {
		t.Errorf("error %v should wrap ErrDeliveryFailed", err)
	}
	if strings.Contains(err.Error(), "482913") {
		t.Errorf("error leaked the OTP code: %v", err)
	}
	if !strings.Contains(err.Error(), "132000") {
		t.Errorf("error should carry the provider code for triage: %v", err)
	}
}

func TestNewSenderFailsClosed(t *testing.T) {
	cases := map[string]Config{
		"no phone number id": {AccessToken: "t"},
		"no access token":    {PhoneNumberID: "1"},
		"bad api version":    {PhoneNumberID: "1", AccessToken: "t", APIVersion: "21.0"},
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSender(config, nil); !errors.Is(err, ErrNotConfigured) {
				t.Fatalf("NewSender error = %v, want ErrNotConfigured", err)
			}
		})
	}
}
