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
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/application"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
)

const integrationTimeout = 3 * time.Minute

func TestMarketPackGovernanceEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_marketpack_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "pack_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewMarketPackService(repository, time.Now, ids)

	// Draft a Twi pack; audited.
	pack, err := service.Draft(ctx, domain.MarketGhanaTwi, "term:gh-tw:1", map[string]bool{"fires": true, "circles": true}, "proposer-1")
	if err != nil {
		t.Fatalf("draft: %v", err)
	}
	auditCount, err := database.Collection("configuration_changes").CountDocuments(ctx, bson.M{"packId": pack.ID()})
	if err != nil || auditCount != 1 {
		t.Fatalf("audits = %d, want 1", auditCount)
	}

	// Self-approval rejected (four eyes).
	if _, err := service.Publish(ctx, pack.ID(), "proposer-1"); err != domain.ErrSelfApproval {
		t.Fatalf("self-approval = %v, want rejected", err)
	}

	// Second approver publishes; listed as published.
	if _, err := service.Publish(ctx, pack.ID(), "approver-1"); err != nil {
		t.Fatal(err)
	}
	published, err := service.Published(ctx)
	if err != nil || len(published) != 1 || published[0].ID() != pack.ID() {
		t.Fatalf("published = %#v, %v", published, err)
	}

	// Audit trail: draft + publish.
	auditCount, _ = database.Collection("configuration_changes").CountDocuments(ctx, bson.M{"packId": pack.ID()})
	if auditCount != 2 {
		t.Fatalf("audits = %d, want 2", auditCount)
	}

	// Retire removes it from the published list.
	if _, err := service.Retire(ctx, pack.ID(), "approver-1"); err != nil {
		t.Fatal(err)
	}
	published, _ = service.Published(ctx)
	if len(published) != 0 {
		t.Fatalf("after retire = %#v", published)
	}
}
