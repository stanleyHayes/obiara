//go:build integration

package privacy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	privacymongodb "github.com/stanleyHayes/obiara/internal/privacy/adapters/outbound/mongodb"
	privacyapplication "github.com/stanleyHayes/obiara/internal/privacy/application"
	"github.com/stanleyHayes/obiara/internal/privacy/domain"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/adapters/outbound/mongodb"
	privacyjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/privacy"
)

const integrationTimeout = 3 * time.Minute

func TestPrivacyProcessorJobEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_worker_privacy_test")

	// Seed member data across contexts, including token hashes that must
	// never reach an export archive.
	if _, err := database.Collection("accounts").InsertOne(ctx, bson.M{
		"_id": "id_1", "phone": "+233550000101", "status": "active", "createdAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Collection("sessions").InsertOne(ctx, bson.M{
		"_id": "sess_1", "memberId": "id_1", "deviceId": "device-1", "status": "active",
		"accessTokenHash": "deadbeef", "refreshTokenHash": "cafef00d", "version": 1, "createdAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Collection("doorway_questions").InsertOne(ctx, bson.M{
		"_id": "id_1", "text": "What does home mean to you?", "custom": true, "updatedAt": time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	requests := privacymongodb.NewRequestRepository(database)
	if err := requests.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	service := privacyapplication.NewPrivacyService(requests, requests, time.Now,
		func() func() string {
			var counter int
			return func() string {
				counter++
				return "pr_worker_" + string(rune('a'+counter))
			}
		}())
	if _, err := service.RequestExport(ctx, "id_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequestDeletion(ctx, "id_1"); err != nil {
		t.Fatal(err)
	}

	processor := privacyapplication.NewProcessor(
		requests,
		privacymongodb.NewArchiveAssembler(database, time.Now),
		privacymongodb.NewErasureRunner(database, time.Now),
		mongodb.NewSessionRevoker(database, time.Now),
		time.Now,
	)
	job := privacyjob.NewProcessorJob(processor, 25, time.Minute)
	if err := job.Run(ctx); err != nil {
		t.Fatalf("processor job: %v", err)
	}

	// Export archive exists with hashes stripped.
	var archive struct {
		Payload []byte `bson:"payload"`
	}
	if err := database.Collection("privacy_export_archives").FindOne(ctx, bson.M{"accountId": "id_1"}).Decode(&archive); err != nil {
		t.Fatalf("archive missing: %v", err)
	}
	var sections map[string][]bson.M
	if err := json.Unmarshal(archive.Payload, &sections); err != nil {
		t.Fatalf("archive payload does not decode: %v", err)
	}
	if len(sections["account"]) != 1 || len(sections["doorwayQuestion"]) != 1 {
		t.Fatalf("archive sections = %#v", sections)
	}
	for _, session := range sections["sessions"] {
		if _, ok := session["accessTokenHash"]; ok {
			t.Fatal("access token hash leaked into export archive")
		}
		if _, ok := session["refreshTokenHash"]; ok {
			t.Fatal("refresh token hash leaked into export archive")
		}
	}

	// Deletion erased and tombstoned.
	var account struct {
		Status string `bson:"status"`
	}
	if err := database.Collection("accounts").FindOne(ctx, bson.M{"_id": "id_1"}).Decode(&account); err != nil {
		t.Fatal(err)
	}
	if account.Status != "deleted" {
		t.Fatalf("account status = %q, want deleted", account.Status)
	}
	for collection, field := range map[string]string{"doorway_questions": "_id", "sessions": "memberId"} {
		count, err := database.Collection(collection).CountDocuments(ctx, bson.M{field: "id_1"})
		if err != nil || count != 0 {
			t.Fatalf("%s still holds %d documents", collection, count)
		}
	}
	// Sessions were deleted by the erasure runner.
	if count, _ := database.Collection("sessions").CountDocuments(ctx, bson.M{"memberId": "id_1"}); count != 0 {
		t.Fatal("sessions not erased")
	}

	// Both requests completed.
	for _, kind := range []domain.Kind{domain.KindExport, domain.KindDeletion} {
		if _, err := requests.FindOpenByAccountAndKind(ctx, "id_1", kind); err != privacyapplication.ErrRequestNotFound {
			t.Fatalf("open %s request remains: %v", kind, err)
		}
	}
}
