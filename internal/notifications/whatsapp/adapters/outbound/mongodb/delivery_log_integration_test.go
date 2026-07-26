//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	notificationmongodb "github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	notificationapplication "github.com/stanleyHayes/obiara/internal/notifications/application"
	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

const integrationTimeout = 3 * time.Minute

func TestWhatsAppChannelEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_whatsapp_test")
	deliveryLog := mongodb.NewDeliveryLog(database)
	if err := deliveryLog.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	// Decider clock pinned outside quiet hours for determinism.
	decisionClock := func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	}
	preferencesRepository := notificationmongodb.NewRepository(database)
	decider := notificationapplication.NewNotificationService(preferencesRepository, preferencesRepository, decisionClock)

	sender := simulator.NewSender()
	channel := application.NewChannelService(sender, deliveryLog, decider, time.Now)

	// OTP: no preference gate, logged as sent.
	if _, err := channel.SendOtp(ctx, "+233550000101", "123456"); err != nil {
		t.Fatalf("send otp: %v", err)
	}

	// Pod alert for an allowed member: sent and logged.
	if _, err := channel.SendPodAlert(ctx, "m-1", "+233550000101", "pod_1"); err != nil {
		t.Fatalf("pod alert: %v", err)
	}

	// Muted member: nothing sent or logged.
	if _, err := decider.Configure(ctx, "m-2", map[notificationdomain.Category]bool{notificationdomain.CategoryPods: true}, 21*60, 7*60, "Africa/Accra"); err != nil {
		t.Fatal(err)
	}
	if _, err := channel.SendPodAlert(ctx, "m-2", "+233550000102", "pod_2"); err != nil {
		t.Fatal(err)
	}

	sent := sender.Sent()
	if len(sent) != 2 {
		t.Fatalf("simulator sends = %d, want 2 (otp + allowed pod alert)", len(sent))
	}
	sentCount, err := deliveryLog.CountByStatus(ctx, "sent")
	if err != nil || sentCount != 2 {
		t.Fatalf("delivery log sent = %d, want 2", sentCount)
	}
}
