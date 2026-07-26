//go:build integration

package mongodb_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/email/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

const integrationTimeout = 3 * time.Minute

func TestEmailChannelEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_email_test")
	deliveryLog := mongodb.NewDeliveryLog(database)
	if err := deliveryLog.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	sender := simulator.NewSender()
	service := application.NewEmailService(sender, deliveryLog, time.Now)
	webhook := application.NewWebhookService(deliveryLog, "whsec_dGVzdC1zZWNyZXQtdGVzdC1zZWNyZXQtdGVzdA==", time.Now)

	// Send: logged as sent.
	ref, err := service.Send(ctx, "ops@example.test", domain.TemplateOpsAlert, map[string]string{"queue": "verification"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	status, err := deliveryLog.StatusOf(ctx, ref)
	if err != nil || status != "sent" {
		t.Fatalf("status = %q, %v", status, err)
	}

	// Webhook: delivered status applies by provider reference.
	if err := webhook.ApplyStatus(ctx, ref, "delivered"); err != nil {
		t.Fatal(err)
	}
	status, _ = deliveryLog.StatusOf(ctx, ref)
	if status != "delivered" {
		t.Fatalf("status = %q, want delivered", status)
	}

	// Unknown references fail loudly (dead-letter input for the relay).
	if err := webhook.ApplyStatus(ctx, "no-such-ref", "delivered"); err != application.ErrDeliveryNotFound {
		t.Fatalf("unknown ref = %v, want not found", err)
	}

	// Signature round-trip against the real service: a properly signed
	// payload passes; a tampered body fails.
	secret := "dGVzdC1zZWNyZXQtdGVzdC1zZWNyZXQtdGVzdA=="
	body := []byte(`{"type":"email.delivered","data":{"email_id":"x"}}`)
	timestamp := time.Now().Unix()
	key, _ := base64.StdEncoding.DecodeString(secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(fmt.Sprintf("%s.%d.%s", "msg_1", timestamp, string(body))))
	signature := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if err := webhook.VerifySignature("msg_1", fmt.Sprintf("%d", timestamp), signature, body); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if err := webhook.VerifySignature("msg_1", fmt.Sprintf("%d", timestamp), signature, []byte(`{"tampered":true}`)); err != application.ErrSignatureInvalid {
		t.Fatalf("tampered body = %v, want invalid", err)
	}
}
