//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
)

const doorwayIntegrationTimeout = 3 * time.Minute

func TestDoorwayAndVaultEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), doorwayIntegrationTimeout)
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

	database := client.Database("obiara_profile_doorway_test")
	vaultRepository := mongodb.NewVaultRepository(database)
	if err := vaultRepository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure vault indexes: %v", err)
	}
	doorway := application.NewDoorwayService(mongodb.NewDoorwayRepository(database), time.Now)
	ids := func() func() string {
		var counter atomic.Int64
		return func() string { return fmt.Sprintf("vi_%d", counter.Add(1)) }
	}()
	vault := application.NewVaultService(vaultRepository, time.Now, ids)

	// Missing question reports not-found.
	if _, err := doorway.Get(ctx, "m-1"); err != application.ErrDoorwayQuestionMissing {
		t.Fatalf("get missing = %v, want missing", err)
	}

	// Set then update (upsert keeps one question per member).
	if _, err := doorway.Set(ctx, "m-1", "What does home mean to you?", true); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := doorway.Set(ctx, "m-1", "What does family mean to you?", false); err != nil {
		t.Fatalf("update: %v", err)
	}
	question, err := doorway.Get(ctx, "m-1")
	if err != nil || question.Text() != "What does family mean to you?" || question.Custom() {
		t.Fatalf("question = %#v, %v", question, err)
	}

	// Vault: add two items, position conflict rejected, veil applied.
	if _, err := vault.Add(ctx, "m-1", "asset-1", 0); err != nil {
		t.Fatalf("add 0: %v", err)
	}
	if _, err := vault.Add(ctx, "m-1", "asset-2", 0); err != application.ErrVaultItemConflict {
		t.Fatalf("position conflict = %v, want conflict", err)
	}
	if _, err := vault.Add(ctx, "m-1", "asset-2", 1); err != nil {
		t.Fatalf("add 1: %v", err)
	}

	stranger, err := vault.ViewFor(ctx, "m-1", "m-2")
	if err != nil || len(stranger) != 2 {
		t.Fatalf("stranger view = %#v, %v", stranger, err)
	}
	for _, view := range stranger {
		if !view.Veiled {
			t.Fatal("stranger must see everything veiled")
		}
	}
	owner, _ := vault.ViewFor(ctx, "m-1", "m-1")
	for _, view := range owner {
		if view.Veiled {
			t.Fatal("owner must see clear items")
		}
	}
}
