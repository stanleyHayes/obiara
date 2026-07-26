//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	identitymongodb "github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	verificationmongodb "github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

const integrationTimeout = 3 * time.Minute

type tierBridge struct {
	tiers identityapplication.TierService
}

func (bridge tierBridge) Transition(ctx context.Context, accountID string, target int, reason, actorID string) error {
	_, err := bridge.tiers.Transition(ctx, accountID, identitydomain.Tier(target), reason, actorID)
	return err
}

func TestVerificationFlowEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_verification_test")
	caseRepository := verificationmongodb.NewCaseRepository(database)
	accountRepository := identitymongodb.NewAccountRepository(database)
	if err := caseRepository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("case indexes: %v", err)
	}
	if err := accountRepository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("account indexes: %v", err)
	}

	tierService := identityapplication.NewTierService(accountRepository, time.Now)
	ids := func() string { return "vc_" + strings.ToLower(identitydomain.HashToken(time.Now().String())[:16]) }
	service := application.NewVerificationService(caseRepository, simulator.NewProvider(), tierBridge{tiers: tierService}, time.Now, ids)

	dob := time.Date(1998, time.April, 12, 0, 0, 0, 0, time.UTC)
	account, err := identitydomain.NewAccount("id_member_1", "+233550000111", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := accountRepository.Create(ctx, account); err != nil {
		t.Fatal(err)
	}

	// Match: approved and the account reaches Tier 1.
	approved, err := service.SubmitGhanaCard(ctx, account.ID(), "GHA-000000000-1", dob)
	if err != nil {
		t.Fatalf("submit match: %v", err)
	}
	if approved.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", approved.Status())
	}
	account, err = accountRepository.FindByID(ctx, account.ID())
	if err != nil {
		t.Fatal(err)
	}
	if account.Tier() != identitydomain.TierVerified {
		t.Fatalf("account tier = %d, want Tier 1", account.Tier())
	}

	// Outage: routed to the manual queue, listed oldest-first, and a human
	// approval promotes a second account.
	account2, _ := identitydomain.NewAccount("id_member_2", "+233550000112", time.Now())
	if err := accountRepository.Create(ctx, account2); err != nil {
		t.Fatal(err)
	}
	queued, err := service.SubmitGhanaCard(ctx, account2.ID(), "GHA-00000000-U", dob)
	if err != nil {
		t.Fatalf("submit outage: %v", err)
	}
	if queued.Status() != domain.StatusQueuedManual {
		t.Fatalf("status = %q, want queued_manual", queued.Status())
	}
	queue, err := service.ManualQueue(ctx, 10)
	if err != nil || len(queue) != 1 || queue[0].ID() != queued.ID() {
		t.Fatalf("queue = %#v, %v", queue, err)
	}
	decided, err := service.DecideManual(ctx, queued.ID(), true, "verified in person", "agent-1")
	if err != nil {
		t.Fatalf("manual decide: %v", err)
	}
	if decided.Status() != domain.StatusApproved {
		t.Fatalf("decided status = %q", decided.Status())
	}
	account2, _ = accountRepository.FindByID(ctx, account2.ID())
	if account2.Tier() != identitydomain.TierVerified {
		t.Fatalf("account2 tier = %d, want Tier 1 after manual approval", account2.Tier())
	}
	if queue, _ := service.ManualQueue(ctx, 10); len(queue) != 0 {
		t.Fatal("decided case still queued")
	}
}
