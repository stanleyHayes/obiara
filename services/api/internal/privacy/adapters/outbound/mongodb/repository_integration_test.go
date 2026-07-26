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
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/application"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/domain"
)

const integrationTimeout = 3 * time.Minute

type stubAssembler struct{ called []string }

func (stub *stubAssembler) Assemble(_ context.Context, accountID string) (string, error) {
	stub.called = append(stub.called, accountID)
	return "archive://" + accountID, nil
}

type stubErasure struct{ called []string }

func (stub *stubErasure) Erase(_ context.Context, accountID string) error {
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
}
