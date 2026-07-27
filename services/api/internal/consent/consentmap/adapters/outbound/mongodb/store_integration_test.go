//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/application"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
)

const integrationTimeout = 3 * time.Minute

func TestConsentMapEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_consentmap_test")
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "cons_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewConsentMapService(store, store, time.Now, ids)

	// Defaults resolve without any record.
	on, err := service.StateFor(ctx, "m-1", domain.PurposeScamArc)
	if err != nil || !on {
		t.Fatalf("scam-arc default = %v, %v", on, err)
	}

	// Opt-in: matching turns on, with a receipt.
	if _, err := service.Set(ctx, "m-1", domain.PurposeMatching, true); err != nil {
		t.Fatal(err)
	}
	receiptCount, err := database.Collection("consent_receipts").CountDocuments(ctx, bson.M{"memberId": "m-1"})
	if err != nil || receiptCount != 1 {
		t.Fatalf("receipts = %d, want 1", receiptCount)
	}

	// Opt-out: scam-arc turns off.
	if _, err := service.Set(ctx, "m-1", domain.PurposeScamArc, false); err != nil {
		t.Fatal(err)
	}
	on, _ = service.StateFor(ctx, "m-1", domain.PurposeScamArc)
	if on {
		t.Fatal("explicit opt-out must win")
	}

	// Locked and wrong-direction changes are rejected without receipts.
	if _, err := service.Set(ctx, "m-1", domain.PurposeIdentitySafety, false); err != domain.ErrPurposeLocked {
		t.Fatalf("locked = %v", err)
	}
	if _, err := service.Set(ctx, "m-1", domain.PurposeMatching, false); err != domain.ErrWrongDirection {
		t.Fatalf("wrong direction = %v", err)
	}
	receiptCount, _ = database.Collection("consent_receipts").CountDocuments(ctx, bson.M{"memberId": "m-1"})
	if receiptCount != 2 {
		t.Fatalf("receipts = %d, want exactly 2", receiptCount)
	}

	// Switchboard merges choices with defaults.
	board, err := service.Switchboard(ctx, "m-1")
	if err != nil {
		t.Fatal(err)
	}
	if !board[domain.PurposeIdentitySafety] || !board[domain.PurposeMatching] ||
		board[domain.PurposeScamArc] || board[domain.PurposePlayPortraits] || !board[domain.PurposeProductAnalytics] {
		t.Fatalf("board = %#v", board)
	}
}
