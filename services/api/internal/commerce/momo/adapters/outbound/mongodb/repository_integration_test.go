//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	mongoadapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

func TestMongoIdempotencyConcurrencyAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("momo_test")
	repo := mongoadapter.New(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	intent, _ := domain.Create(key(1), key(2), key(3), 1200, "create-1", now)
	results := make(chan error, 2)
	go func() { results <- repo.Create(ctx, intent) }()
	go func() { results <- repo.Create(ctx, intent) }()
	saved, replay := 0, 0
	for range 2 {
		err = <-results
		if err == nil {
			saved++
		} else if errors.Is(err, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(err)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("saved=%d replay=%d", saved, replay)
	}
	loaded, err := repo.Find(ctx, key(1))
	if err != nil {
		t.Fatal(err)
	}
	confirmed, _ := loaded.Confirm("confirm-1", now.Add(time.Second))
	if err = repo.Save(ctx, confirmed, loaded.Revision(), "confirm-1"); err != nil {
		t.Fatal(err)
	}
	if err = repo.Save(ctx, confirmed, loaded.Revision(), "confirm-1"); !errors.Is(err, application.ErrApplied) {
		t.Fatalf("replay=%v", err)
	}
	var raw bson.M
	if err = db.Collection("momo_intents").FindOne(ctx, bson.M{}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"+233201234567", "phoneNumber", "rawPhone", "msisdn", "sku", "seed", "visibility"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
}
