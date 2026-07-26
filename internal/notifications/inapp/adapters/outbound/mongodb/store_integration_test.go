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
	inappmongodb "github.com/stanleyHayes/obiara/internal/notifications/inapp/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/adapters/outbound/inappsender"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

const integrationTimeout = 3 * time.Minute

func TestRoutingEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_routing_test")
	inappStore := inappmongodb.NewStore(database)
	if err := inappStore.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	preferencesRepository := notificationmongodb.NewRepository(database)
	decisionClock := func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	}
	decider := notificationapplication.NewNotificationService(preferencesRepository, preferencesRepository, decisionClock)

	push := simulator.NewPush()
	push.FailNext = true // force the fallback
	sms := simulator.NewSMS()
	router := application.NewRouter(
		[]application.ChannelSender{push, sms, inappsender.New(inappStore, time.Now)},
		decider,
		time.Now,
	)

	// Pods: push fails, no whatsapp registered → in-app wins.
	result, err := router.Deliver(ctx, application.Outbound{
		MemberID: "m-1", Phone: "+233550000101", Purpose: domain.PurposePods,
		Template: "pod_alert", Reference: "pod_1",
	})
	if err != nil || result.Channel != domain.ChannelInApp {
		t.Fatalf("result = %#v, %v, want in-app fallback", result, err)
	}

	// The inbox holds the entry; OTP (SMS-primary) delivers via SMS.
	feed, err := inappStore.ListForMember(ctx, "m-1", 10)
	if err != nil || len(feed) != 1 || feed[0].Reference() != "pod_1" {
		t.Fatalf("inbox = %#v, %v", feed, err)
	}
	result, err = router.Deliver(ctx, application.Outbound{
		MemberID: "m-1", Phone: "+233550000101", Purpose: domain.PurposeOtp,
		Template: "otp_code", Reference: "123456",
	})
	if err != nil || result.Channel != domain.ChannelSMS {
		t.Fatalf("otp result = %#v, %v, want SMS primary", result, err)
	}
	if len(sms.Sent()) != 1 {
		t.Fatalf("sms sends = %d", len(sms.Sent()))
	}

	// Mark read is idempotent.
	if err := inappStore.MarkRead(ctx, feed[0].ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := inappStore.MarkRead(ctx, feed[0].ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	feed, _ = inappStore.ListForMember(ctx, "m-1", 10)
	if feed[0].ReadAt() == nil {
		t.Fatal("notification must be marked read")
	}
}
