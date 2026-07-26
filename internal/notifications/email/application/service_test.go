package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

var emailNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestSendLogsOutcome(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)

	sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return("ref-1", nil)
	log.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, entry DeliveryEntry) error {
			if entry.Template != domain.TemplateOpsAlert || entry.Status != "sent" || entry.ProviderRef != "ref-1" {
				t.Fatalf("entry = %#v", entry)
			}
			return nil
		})

	service := NewEmailService(sender, log, func() time.Time { return emailNow })
	ref, err := service.Send(context.Background(), "ops@example.test", domain.TemplateOpsAlert, map[string]string{"queue": "verification"})
	if err != nil || ref != "ref-1" {
		t.Fatalf("Send = %q, %v", ref, err)
	}
}

func TestSendRejectsInvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)
	// No Send/Record expectation.
	service := NewEmailService(sender, log, func() time.Time { return emailNow })
	if _, err := service.Send(context.Background(), "not-an-email", domain.TemplateOpsAlert, nil); err == nil {
		t.Fatal("invalid recipient accepted")
	}
}

func sign(t *testing.T, secret, svixID, timestamp string, body []byte) string {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(fmt.Sprintf("%s.%s.%s", svixID, timestamp, string(body))))
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

const testSecret = "dGVzdC1zZWNyZXQtdGVzdC1zZWNyZXQtdGVzdA=="

func webhookService(log DeliveryLog) WebhookService {
	return NewWebhookService(log, "whsec_"+testSecret, func() time.Time { return emailNow })
}

func TestVerifySignature(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := NewMockDeliveryLog(ctrl)
	service := webhookService(log)
	body := []byte(`{"type":"email.delivered","data":{"email_id":"ref-1"}}`)
	timestamp := fmt.Sprintf("%d", emailNow.Unix())

	signature := sign(t, testSecret, "msg_1", timestamp, body)
	if err := service.VerifySignature("msg_1", timestamp, signature, body); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if err := service.VerifySignature("msg_1", timestamp, "v1,AAAAAAAA", body); err != ErrSignatureInvalid {
		t.Fatalf("wrong signature = %v", err)
	}
	if err := service.VerifySignature("msg_1", timestamp, sign(t, testSecret, "msg_2", timestamp, body), body); err != ErrSignatureInvalid {
		t.Fatalf("wrong id = %v", err)
	}
	stale := fmt.Sprintf("%d", emailNow.Add(-10*time.Minute).Unix())
	if err := service.VerifySignature("msg_1", stale, sign(t, testSecret, "msg_1", stale, body), body); err != ErrTimestampStale {
		t.Fatalf("stale timestamp = %v", err)
	}
	// Multiple candidate signatures: any valid v1 matches.
	if err := service.VerifySignature("msg_1", timestamp, "v1,AAAA "+signature, body); err != nil {
		t.Fatalf("rotation signature rejected: %v", err)
	}
}

func TestVerifySignatureMissingSecret(t *testing.T) {
	service := NewWebhookService(nil, "", func() time.Time { return emailNow })
	if err := service.VerifySignature("msg_1", "1", "v1,x", nil); err != ErrSecretMissing {
		t.Fatalf("missing secret = %v", err)
	}
}

func TestApplyStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := NewMockDeliveryLog(ctrl)
	log.EXPECT().UpdateStatus(gomock.Any(), "ref-1", "delivered", gomock.Any()).Return(nil)
	service := webhookService(log)
	if err := service.ApplyStatus(context.Background(), "ref-1", "delivered"); err != nil {
		t.Fatal(err)
	}
}
