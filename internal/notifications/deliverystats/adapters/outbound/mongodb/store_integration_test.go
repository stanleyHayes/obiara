//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/notifications/deliverystats/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/deliverystats/application"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

const integrationTimeout = 3 * time.Minute

func TestDeliveryStatsEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_deliverystats_test")
	now := time.Now()
	old := now.Add(-60 * 24 * time.Hour)

	// WhatsApp: 8 sent (7 recent, 1 outside window), 2 failed.
	for i := 0; i < 7; i++ {
		insertDelivery(t, ctx, database, "whatsapp_deliveries", "otp_code", "sent", now)
	}
	for i := 0; i < 2; i++ {
		insertDelivery(t, ctx, database, "whatsapp_deliveries", "otp_code", "failed", now)
	}
	insertDelivery(t, ctx, database, "whatsapp_deliveries", "otp_code", "sent", old)

	// Email: 5 sent, 3 delivered, 1 bounced, 1 complained.
	for i := 0; i < 5; i++ {
		insertDelivery(t, ctx, database, "email_deliveries", "ops_alert", "sent", now)
	}
	for i := 0; i < 3; i++ {
		insertDelivery(t, ctx, database, "email_deliveries", "ops_alert", "delivered", now)
	}
	insertDelivery(t, ctx, database, "email_deliveries", "ops_alert", "bounced", now)
	insertDelivery(t, ctx, database, "email_deliveries", "ops_alert", "complained", now)

	service := application.NewStatsService(mongodb.NewStore(database), time.Now)
	report, err := service.Stats(ctx, 30)
	if err != nil {
		t.Fatal(err)
	}

	whatsapp := report.Channels["whatsapp"]
	if whatsapp.Attempted != 9 || whatsapp.Failed != 2 || whatsapp.Sent != 7 {
		t.Fatalf("whatsapp = %#v", whatsapp)
	}
	email := report.Channels["email"]
	if email.Attempted != 10 || email.Failed != 2 || email.Delivered != 3 {
		t.Fatalf("email = %#v", email)
	}
	if email.SuccessRate != 0.8 {
		t.Fatalf("email rate = %v, want 0.8", email.SuccessRate)
	}
}

func insertDelivery(t *testing.T, ctx context.Context, database *mongo.Database, collection, template, status string, at time.Time) {
	t.Helper()
	if _, err := database.Collection(collection).InsertOne(ctx, bson.M{
		"to": "+233550000101", "template": template, "status": status, "at": at, "updatedAt": at,
	}); err != nil {
		t.Fatal(err)
	}
}
