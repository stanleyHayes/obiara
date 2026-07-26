//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	gardenmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func gardenKey(n int) string { return fmt.Sprintf("%064x", n) }

func TestOwnerIsolationExpiryAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("seed_garden_test")
	repository := gardenmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 6, 0, 0, 0, time.UTC)
	due, _ := domain.New(gardenKey(1), gardenKey(10), now.Add(time.Minute), now)
	other, _ := domain.New(gardenKey(2), gardenKey(20), now.Add(time.Minute), now)
	if err = repository.Create(ctx, due); err != nil {
		t.Fatal(err)
	}
	if err = repository.Create(ctx, other); err != nil {
		t.Fatal(err)
	}
	changed, err := repository.ExpireDue(ctx, gardenKey(10), now.Add(2*time.Minute), 100)
	if err != nil || changed != 1 {
		t.Fatalf("changed=%d err=%v", changed, err)
	}
	ownerItems, err := repository.ListOwner(ctx, gardenKey(10))
	if err != nil || len(ownerItems) != 1 || ownerItems[0].State != domain.StateExpired {
		t.Fatalf("owner items=%#v err=%v", ownerItems, err)
	}
	otherItems, err := repository.ListOwner(ctx, gardenKey(20))
	if err != nil || len(otherItems) != 1 || otherItems[0].State != domain.StateQueued {
		t.Fatalf("other items=%#v err=%v", otherItems, err)
	}
	var raw bson.M
	if err = database.Collection("seed_garden").FindOne(ctx, bson.M{"ownerKey": gardenKey(10)}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{"raw-member", "raw-seed", "readReceipt", "declineReason", "streak", "publicActivity"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("stored privacy leak %q", forbidden)
		}
	}
}
