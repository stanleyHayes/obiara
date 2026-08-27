package arkesel

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

const testCode = "482913"

func newTestSender(t *testing.T, handler http.HandlerFunc) (*Sender, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	sender, err := NewSender(Config{
		APIKey:  "synthetic-test-only",
		Sender:  "Obiara",
		BaseURL: server.URL,
	}, server.Client())
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	return sender, server
}

func TestSendAcceptsSuccessStatus(t *testing.T) {
	var captured struct {
		Sender     string   `json:"sender"`
		Message    string   `json:"message"`
		Recipients []string `json:"recipients"`
	}
	var apiKey string

	sender, _ := newTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		apiKey = r.Header.Get("api-key")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":[{"recipient":"233544919953","id":"abc"}]}`))
	})
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	if err := sender.Send(context.Background(), contact, testCode); err != nil {
		t.Fatalf("Send returned %v, want nil", err)
	}
	if apiKey != "synthetic-test-only" {
		t.Errorf("api-key header = %q", apiKey)
	}
	if captured.Sender != "Obiara" {
		t.Errorf("sender = %q, want Obiara", captured.Sender)
	}
	// Arkesel documents recipients without the leading "+".
	if len(captured.Recipients) != 1 || captured.Recipients[0] != "233544919953" {
		t.Errorf("recipients = %v, want [233544919953]", captured.Recipients)
	}
	if !strings.Contains(captured.Message, testCode) {
		t.Errorf("message %q does not carry the code", captured.Message)
	}
}

// TestSendRejectsSuccessStatusCodeWithFailureBody is the important one:
// Arkesel answers an unapproved sender id or an exhausted balance with HTTP
// 200 and a failure status. Treating 2xx as delivery would report success
// while the member never receives a code.
func TestSendRejectsSuccessStatusCodeWithFailureBody(t *testing.T) {
	for _, body := range []string{
		`{"status":"error","message":"Insufficient balance","code":"104"}`,
		`{"status":"failed","message":"Sender ID not approved"}`,
	} {
		t.Run(body, func(t *testing.T) {
			sender, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(body))
			})
			contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

			err := sender.Send(context.Background(), contact, testCode)
			if err == nil {
				t.Fatal("Send accepted a 200 response carrying a failure status")
			}
			if !errors.Is(err, ErrDeliveryFailed) {
				t.Errorf("error %v should wrap ErrDeliveryFailed", err)
			}
		})
	}
}

func TestSendNeverLeaksTheCodeOrKey(t *testing.T) {
	sender, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		// A provider echoing the body back must not become a leak path.
		_, _ = w.Write([]byte(`{"status":"error","message":"bad request for code ` + testCode +
			` with key synthetic-test-only"}`))
	})
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	err := sender.Send(context.Background(), contact, testCode)
	if err == nil {
		t.Fatal("Send accepted a 400 response")
	}
	if strings.Contains(err.Error(), testCode) {
		t.Errorf("error leaked the OTP code: %v", err)
	}
	if strings.Contains(err.Error(), "synthetic-test-only") {
		t.Errorf("error leaked the api key: %v", err)
	}
}

func TestSendRetriesServerFaultsOnce(t *testing.T) {
	var attempts atomic.Int32
	sender, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"status":"success"}`))
	})
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	if err := sender.Send(context.Background(), contact, testCode); err != nil {
		t.Fatalf("Send returned %v after a retryable fault, want nil", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
}

func TestSendDoesNotRetryBusinessFailures(t *testing.T) {
	var attempts atomic.Int32
	sender, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		_, _ = w.Write([]byte(`{"status":"error","message":"Insufficient balance"}`))
	})
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	if err := sender.Send(context.Background(), contact, testCode); err == nil {
		t.Fatal("Send accepted a business failure")
	}
	// Resending cannot fix an empty balance, and each attempt is billable.
	if got := attempts.Load(); got != 1 {
		t.Errorf("attempts = %d, want 1", got)
	}
}

func TestSendHonoursContextCancellation(t *testing.T) {
	sender, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success"}`))
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	if err := sender.Send(ctx, contact, testCode); err == nil {
		t.Fatal("Send ignored a cancelled context")
	}
}

func TestNewSenderFailsClosed(t *testing.T) {
	cases := map[string]Config{
		"no api key":      {Sender: "Obiara"},
		"no sender id":    {APIKey: "k"},
		"sender too long": {APIKey: "k", Sender: "ObiaraGhanaLimited"},
		"template without a verb": {
			APIKey: "k", Sender: "Obiara", Template: "your code",
		},
		"template with two verbs": {
			APIKey: "k", Sender: "Obiara", Template: "%s and %s",
		},
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSender(config, nil); !errors.Is(err, ErrNotConfigured) {
				t.Fatalf("NewSender error = %v, want ErrNotConfigured", err)
			}
		})
	}
}

func TestCustomTemplateIsUsed(t *testing.T) {
	var captured struct {
		Message string `json:"message"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	t.Cleanup(server.Close)

	sender, err := NewSender(Config{
		APIKey: "k", Sender: "Obiara", BaseURL: server.URL,
		Template: "Obiara: %s. Akwaaba.",
	}, server.Client())
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")
	if err := sender.Send(context.Background(), contact, testCode); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if captured.Message != "Obiara: "+testCode+". Akwaaba." {
		t.Errorf("message = %q", captured.Message)
	}
}
