//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/application"
)

const testTimeout = 3 * time.Minute

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "restriction:integration" }

type failOncePurger struct {
	delegate application.ArtifactPurger
	failed   bool
}

func (purger *failOncePurger) Purge(ctx context.Context, subjectID, sourceRef string) error {
	if !purger.failed {
		purger.failed = true
		return errors.New("simulated object-store outage")
	}
	return purger.delegate.Purge(ctx, subjectID, sourceRef)
}

func TestUnder18HardBlockPurgesPersonalDataWithin24Hours(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
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

	database := client.Database("obiara_safeguarding_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure safeguarding indexes: %v", err)
	}
	subjectID := "account:minor"
	sourceRef := "verification:minor"
	cardNumber := "GHA-000000000-0"
	dateOfBirth := time.Date(2012, 5, 4, 0, 0, 0, 0, time.UTC)
	fixtures := []struct {
		collection string
		document   bson.M
	}{
		{"verification_cases", bson.M{"_id": sourceRef, "accountId": subjectID, "cardNumber": cardNumber, "dateOfBirth": dateOfBirth}},
		{"accounts", bson.M{"_id": subjectID, "phone": "+233500000000"}},
		{"sessions", bson.M{"_id": "session:minor", "memberId": subjectID}},
		{"consent_records", bson.M{"_id": "consent:minor", "subjectId": subjectID}},
		{"media_assets", bson.M{"_id": "asset:minor", "ownerId": subjectID}},
	}
	for _, fixture := range fixtures {
		if _, err := database.Collection(fixture.collection).InsertOne(ctx, fixture.document); err != nil {
			t.Fatalf("seed %s: %v", fixture.collection, err)
		}
	}

	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("s", 32)))
	if err != nil {
		t.Fatal(err)
	}
	serverTime := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	purger := &failOncePurger{delegate: mongodb.NewArtifactPurger(database)}
	service := application.NewService(repository, purger, keyer, fixedIDs{}, func() time.Time {
		return serverTime
	})

	decision, err := service.Assess(ctx, application.Assessment{
		CommandID: "command:minor", SubjectID: subjectID, SourceRef: sourceRef,
		DateOfBirth: dateOfBirth,
	})
	if decision.Allowed || !errors.Is(err, application.ErrUnder18) ||
		!errors.Is(err, application.ErrPurgePending) {
		t.Fatalf("initial decision = %+v, %v; want blocked with pending purge", decision, err)
	}
	if count, _ := database.Collection("verification_cases").CountDocuments(ctx, bson.M{"_id": sourceRef}); count != 1 {
		t.Fatal("outage fixture should remain queued for retry")
	}

	// A worker running one minute before the deadline includes the next
	// interval and completes every Mongo deletion before 24 hours elapse.
	serverTime = decision.Restriction.PurgeDueAt().Add(-time.Minute)
	completed, err := service.PurgePending(ctx, decision.Restriction.PurgeDueAt(), 10)
	if err != nil || completed != 1 {
		t.Fatalf("purge retry completed=%d, err=%v", completed, err)
	}
	for _, fixture := range fixtures {
		count, countErr := database.Collection(fixture.collection).CountDocuments(ctx, bson.M{})
		if countErr != nil || count != 0 {
			t.Fatalf("%s retained %d personal records: %v", fixture.collection, count, countErr)
		}
	}
	if count, _ := database.Collection("safeguarding_purge_jobs").CountDocuments(ctx, bson.M{}); count != 0 {
		t.Fatal("temporary raw purge locator survived completion")
	}
	proof, err := repository.FindByID(ctx, decision.Restriction.ID())
	if err != nil || !proof.Blocked() || !proof.PurgedWithinSLA() {
		t.Fatalf("retained proof = %+v, %v", proof, err)
	}
	if count, _ := database.Collection("safeguarding_events").CountDocuments(ctx, bson.M{
		"restrictionId": proof.ID(),
	}); count != 2 {
		t.Fatalf("audit event count=%d, want blocked+purged", count)
	}

	// A later payload claiming an adult birth date cannot override the retained
	// subject-key block after all raw purge locators are gone.
	replayed, replayErr := service.Assess(ctx, application.Assessment{
		CommandID: "command:adult-retry", SubjectID: subjectID,
		SourceRef: "verification:new", DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if replayed.Allowed || !replayed.Replayed || !errors.Is(replayErr, application.ErrUnder18) {
		t.Fatalf("retained hard block bypassed: %+v, %v", replayed, replayErr)
	}

	var retained bson.M
	if err := database.Collection("safeguarding_restrictions").FindOne(ctx, bson.M{"_id": proof.ID()}).Decode(&retained); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(retained, false, false)
	for _, forbidden := range []string{subjectID, sourceRef, cardNumber, dateOfBirth.Format("2006-01-02"), "+233500000000"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("retained proof leaked personal data %q: %s", forbidden, encoded)
		}
	}
}
