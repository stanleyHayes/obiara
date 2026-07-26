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
	collisionmongo "github.com/stanleyHayes/obiara/services/api/internal/identity/collision/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/domain"
)

func TestCollisionReviewUsesOnlyPseudonymousKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_collision_test")
	repository := collisionmongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("c", 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.NewService(repository, keyer, func() time.Time { return now })
	rawDevice := "device:fingerprint:private"
	firstAccount := "account:private:first"
	secondAccount := "account:private:second"

	first, err := service.Assess(ctx, application.Assessment{
		Kind: domain.KindSharedDevice, Signal: rawDevice, SubjectID: firstAccount,
	})
	if err != nil || !first.Allowed || first.Collision {
		t.Fatalf("first observation = %+v, %v", first, err)
	}
	second, err := service.Assess(ctx, application.Assessment{
		Kind: domain.KindSharedDevice, Signal: rawDevice, SubjectID: secondAccount,
	})
	if err != nil || second.Allowed || !second.ReviewRequired {
		t.Fatalf("collision = %+v, %v", second, err)
	}
	resolved, err := service.Resolve(ctx, second.Case.ID(), domain.ResolutionApprove, "household_confirmed", "operator:private")
	if err != nil || !resolved.Allowed {
		t.Fatalf("resolution = %+v, %v", resolved, err)
	}

	var stored bson.M
	if err := database.Collection("identity_collision_cases").FindOne(ctx, bson.M{"_id": second.Case.ID()}).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(stored, false, false)
	for _, forbidden := range []string{rawDevice, firstAccount, secondAccount, "operator:private"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("stored collision proof leaked %q: %s", forbidden, encoded)
		}
	}
	audit, ok := stored["audit"].(bson.A)
	if !ok || len(audit) != 2 {
		t.Fatalf("audit events = %#v, want create+resolve", stored["audit"])
	}

	stale, _, err := domain.NewCase(second.Case.ID(), domain.KindSharedDevice, second.Case.SignalKey(), second.Case.SubjectKey(), now)
	if err != nil {
		t.Fatal(err)
	}
	staleAudit, err := stale.Resolve(domain.ResolutionDeny, "duplicate_account", "actor-key", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Resolve(ctx, stale, staleAudit, 1); err != application.ErrStaleCase {
		t.Fatalf("stale resolution = %v, want %v", err, application.ErrStaleCase)
	}
}
