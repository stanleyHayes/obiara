package expo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

func testCopy() domain.Copy {
	return domain.Copy{Title: "Obiara", Body: "A new day in your courtyard."}
}

func newTestSender(t *testing.T, handler http.HandlerFunc) *Sender {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	sender, err := NewSender(Config{BaseURL: server.URL, AccessToken: "synthetic-test-only"}, server.Client())
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	return sender
}

func TestSendDeliversToEveryToken(t *testing.T) {
	var received []message
	var auth string
	sender := newTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		_, _ = w.Write([]byte(`{"data":[{"status":"ok","id":"1"},{"status":"ok","id":"2"}]}`))
	})

	dead, err := sender.Send(context.Background(),
		[]string{"ExponentPushToken[a]", "ExponentPushToken[b]"}, testCopy(), "ref_1")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(dead) != 0 {
		t.Errorf("dead = %v, want none", dead)
	}
	if auth != "Bearer synthetic-test-only" {
		t.Errorf("Authorization = %q", auth)
	}
	if len(received) != 2 {
		t.Fatalf("sent %d messages, want 2", len(received))
	}
	if received[0].Title != "Obiara" || received[0].Body == "" {
		t.Errorf("message = %+v", received[0])
	}
	// The reference is opaque; no content may ride the notification.
	if received[0].Data["ref"] != "ref_1" {
		t.Errorf("data = %v", received[0].Data)
	}
}

// TestFailedTicketsAreNotTreatedAsDelivery is the important one: Expo answers
// a wholly failed batch with HTTP 200 and per-message statuses. Trusting the
// status code would repeat the bug that let OTP codes vanish.
func TestFailedTicketsAreNotTreatedAsDelivery(t *testing.T) {
	sender := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"status":"error","message":"bad","details":{"error":"MessageTooBig"}}]}`))
	})

	_, err := sender.Send(context.Background(), []string{"ExponentPushToken[a]"}, testCopy(), "ref")
	if err == nil {
		t.Fatal("Send accepted a 200 response whose every message failed")
	}
	if !errors.Is(err, ErrDeliveryFailed) {
		t.Errorf("error %v should wrap ErrDeliveryFailed", err)
	}
}

// TestDeviceNotRegisteredIsPruned keeps stale tokens from accumulating until
// they poison whole batches.
func TestDeviceNotRegisteredIsPruned(t *testing.T) {
	sender := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"status":"ok","id":"1"},
			{"status":"error","message":"gone","details":{"error":"DeviceNotRegistered"}}
		]}`))
	})

	dead, err := sender.Send(context.Background(),
		[]string{"ExponentPushToken[live]", "ExponentPushToken[gone]"}, testCopy(), "ref")
	if err != nil {
		t.Fatalf("Send = %v; one dead device must not fail a partly delivered batch", err)
	}
	if len(dead) != 1 || dead[0] != "ExponentPushToken[gone]" {
		t.Errorf("dead = %v, want the unregistered device", dead)
	}
}

func TestRequestLevelErrorsAreReported(t *testing.T) {
	sender := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"invalid credentials"}]}`))
	})
	if _, err := sender.Send(context.Background(), []string{"ExponentPushToken[a]"}, testCopy(), "ref"); err == nil {
		t.Fatal("Send accepted a request-level error")
	}
}

func TestBatchesAreChunked(t *testing.T) {
	var batches int
	sender := newTestSender(t, func(w http.ResponseWriter, r *http.Request) {
		batches++
		var received []message
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		if len(received) > batchSize {
			t.Errorf("batch of %d exceeds Expo's limit of %d", len(received), batchSize)
		}
		tickets := make([]string, 0, len(received))
		for range received {
			tickets = append(tickets, `{"status":"ok","id":"x"}`)
		}
		_, _ = w.Write([]byte(`{"data":[` + join(tickets) + `]}`))
	})

	tokens := make([]string, 250)
	for i := range tokens {
		tokens[i] = "ExponentPushToken[t]"
	}
	if _, err := sender.Send(context.Background(), tokens, testCopy(), "ref"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if batches != 3 {
		t.Errorf("batches = %d, want 3 for 250 tokens", batches)
	}
}

func join(values []string) string {
	out := ""
	for index, value := range values {
		if index > 0 {
			out += ","
		}
		out += value
	}
	return out
}

func TestServerFaultsAreReported(t *testing.T) {
	sender := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	if _, err := sender.Send(context.Background(), []string{"ExponentPushToken[a]"}, testCopy(), "ref"); err == nil {
		t.Fatal("Send accepted a 502")
	}
}
