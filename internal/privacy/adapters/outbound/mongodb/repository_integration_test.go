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
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/privacy/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/privacy/application"
	"github.com/stanleyHayes/obiara/internal/privacy/domain"
)

const integrationTimeout = 3 * time.Minute

type stubAssembler struct{ called []string }

func (stub *stubAssembler) Assemble(_ context.Context, _, accountID string) (string, error) {
	stub.called = append(stub.called, accountID)
	return "archive://" + accountID, nil
}

type stubErasure struct{ called []string }

func (stub *stubErasure) Erase(_ context.Context, _, accountID string) error {
	stub.called = append(stub.called, accountID)
	return nil
}

type stubRevoker struct{ called []string }

func (stub *stubRevoker) RevokeMemberSessions(_ context.Context, memberID string) error {
	stub.called = append(stub.called, memberID)
	return nil
}

func TestPrivacyRequestsAndHoldsEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_privacy_test")
	repository := mongodb.NewRequestRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	ids := func() func() string {
		var counter atomic.Int64
		return func() string { return fmt.Sprintf("pr_%d", counter.Add(1)) }
	}()
	service := application.NewPrivacyService(repository, repository, time.Now, ids)

	// Export opens; a second open export is rejected.
	exportRequest, err := service.RequestExport(ctx, "id_1")
	if err != nil {
		t.Fatalf("request export: %v", err)
	}
	if _, err := service.RequestExport(ctx, "id_1"); err != application.ErrOpenRequestExists {
		t.Fatalf("duplicate export = %v, want exists", err)
	}

	// Deletion opens, then a legal hold blocks it and new deletions.
	deletionRequest, err := service.RequestDeletion(ctx, "id_1")
	if err != nil {
		t.Fatalf("request deletion: %v", err)
	}
	if err := service.PlaceHold(ctx, "id_1", "court order", "agent-1"); err != nil {
		t.Fatalf("place hold: %v", err)
	}
	if _, err := service.RequestDeletion(ctx, "id_1"); err != domain.ErrLegalHoldActive {
		t.Fatalf("deletion under hold = %v, want blocked", err)
	}
	held, err := service.Status(ctx, deletionRequest.ID())
	if err != nil || held.Status() != domain.StatusBlocked {
		t.Fatalf("held request status = %q, %v", held.Status(), err)
	}

	// Processor must skip the held deletion and complete the export.
	assembler := &stubAssembler{}
	erasure := &stubErasure{}
	revoker := &stubRevoker{}
	processor := application.NewProcessor(repository, assembler, erasure, revoker, time.Now)
	if err := processor.RunBatch(ctx, 10); err != nil {
		t.Fatalf("run batch: %v", err)
	}
	if len(assembler.called) != 1 || assembler.called[0] != "id_1" {
		t.Fatalf("assembler calls = %v", assembler.called)
	}
	if len(erasure.called) != 0 {
		t.Fatalf("held deletion must not be erased: %v", erasure.called)
	}

	// Lift the hold: the deletion unblocks and the next batch erases and
	// revokes sessions.
	if err := service.LiftHold(ctx, "id_1"); err != nil {
		t.Fatalf("lift hold: %v", err)
	}
	if err := processor.RunBatch(ctx, 10); err != nil {
		t.Fatalf("second batch: %v", err)
	}
	if len(erasure.called) != 1 || erasure.called[0] != "id_1" {
		t.Fatalf("erasure calls = %v", erasure.called)
	}
	if len(revoker.called) != 1 || revoker.called[0] != "id_1" {
		t.Fatalf("revoker calls = %v", revoker.called)
	}
	completed, err := service.Status(ctx, deletionRequest.ID())
	if err != nil || completed.Status() != domain.StatusCompleted {
		t.Fatalf("deletion status = %q, %v", completed.Status(), err)
	}
	if exportDone, _ := service.Status(ctx, exportRequest.ID()); exportDone.Status() != domain.StatusCompleted {
		t.Fatalf("export status = %q", exportDone.Status())
	}

	t.Run("production cross-context archive and erasure are replay safe", func(t *testing.T) {
		accountID := "id_cross_context"
		for _, seed := range []struct {
			collection string
			document   bson.M
		}{
			{"members", bson.M{"_id": accountID, "email": "member@example.test"}},
			{"profiles", bson.M{"_id": accountID, "displayName": "Archive Subject"}},
			{"photo_vault", bson.M{"_id": "photo-1", "memberId": accountID, "assetId": "opaque-asset"}},
			{"media_assets", bson.M{"_id": "media-1", "ownerId": accountID, "ciphertext": "encrypted"}},
		} {
			if _, err := database.Collection(seed.collection).InsertOne(ctx, seed.document); err != nil {
				t.Fatal(err)
			}
		}
		assembler := mongodb.NewArchiveAssembler(database, func() time.Time { return time.Now().UTC() })
		if err := assembler.EnsureIndexes(ctx); err != nil {
			t.Fatal(err)
		}
		export, err := service.RequestExport(ctx, accountID)
		if err != nil {
			t.Fatal(err)
		}
		processor := application.NewProcessor(repository, assembler, mongodb.NewErasureRunner(database, time.Now), &stubRevoker{}, time.Now)
		if err := processor.RunBatch(ctx, 10); err != nil {
			t.Fatal(err)
		}
		var archive struct {
			ArchiveRef string `bson:"archiveRef"`
			Payload    []byte `bson:"payload"`
		}
		if err := database.Collection("privacy_export_archives").FindOne(ctx, bson.M{"_id": export.ID()}).Decode(&archive); err != nil {
			t.Fatal(err)
		}
		if archive.ArchiveRef != "privacy-export:"+export.ID() || !strings.Contains(string(archive.Payload), "opaque-asset") {
			t.Fatalf("archive ref=%q payload=%s", archive.ArchiveRef, archive.Payload)
		}
		if replayed, err := assembler.Assemble(ctx, export.ID(), accountID); err != nil || replayed != archive.ArchiveRef {
			t.Fatalf("archive replay ref=%q error=%v", replayed, err)
		}
		if _, err := database.Collection("legal_holds").InsertOne(ctx, bson.M{
			"_id": "held-subject", "reason": "court-order", "placedBy": "operator", "placedAt": time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
		if err := mongodb.NewErasureRunner(database, time.Now).Erase(ctx, "pr-held", "held-subject"); err != domain.ErrLegalHoldActive {
			t.Fatalf("held erasure = %v, want legal hold", err)
		}

		deletion, err := service.RequestDeletion(ctx, accountID)
		if err != nil {
			t.Fatal(err)
		}
		revoker := &stubRevoker{}
		processor = application.NewProcessor(repository, assembler, mongodb.NewErasureRunner(database, time.Now), revoker, time.Now)
		if err := processor.RunBatch(ctx, 10); err != nil {
			t.Fatal(err)
		}
		for _, collection := range []string{"members", "profiles", "photo_vault", "media_assets"} {
			count, err := database.Collection(collection).CountDocuments(ctx, bson.M{"$or": bson.A{
				bson.M{"_id": accountID}, bson.M{"memberId": accountID}, bson.M{"ownerId": accountID},
			}})
			if err != nil || count != 0 {
				t.Fatalf("%s retained %d documents: %v", collection, count, err)
			}
		}
		var audit bson.M
		if err := database.Collection("privacy_erasure_audit").FindOne(ctx, bson.M{"_id": deletion.ID()}).Decode(&audit); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(fmt.Sprint(audit), accountID) || len(revoker.called) != 1 {
			t.Fatalf("unsafe audit=%v revocations=%v", audit, revoker.called)
		}
		if err := mongodb.NewErasureRunner(database, time.Now).Erase(ctx, deletion.ID(), accountID); err != nil {
			t.Fatalf("erasure replay: %v", err)
		}
	})
}
