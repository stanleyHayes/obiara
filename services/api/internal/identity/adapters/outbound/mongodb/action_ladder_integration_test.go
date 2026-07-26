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
	safetymongodb "github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	safetyapplication "github.com/stanleyHayes/obiara/internal/safety/application"
	safetydomain "github.com/stanleyHayes/obiara/internal/safety/domain"
	identitymongodb "github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

const actionLadderTimeout = 3 * time.Minute

type sessionRevokerStub struct{ revoked []string }

func (stub *sessionRevokerStub) RevokeMemberSessions(_ context.Context, memberID string) error {
	stub.revoked = append(stub.revoked, memberID)
	return nil
}

func TestActionLadderPropagationEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), actionLadderTimeout)
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

	database := client.Database("obiara_actions_test")
	caseRepository := safetymongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureCaseIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	accountRepository := identitymongodb.NewAccountRepository(database)
	if err := accountRepository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	actionLog := safetymongodb.NewActionLogStore(database)
	enforcement := identityapplication.NewEnforcementService(accountRepository, time.Now)
	revoker := &sessionRevokerStub{}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "act_" + strings.Repeat("z", counter)
		}
	}()

	actionService := safetyapplication.NewActionService(caseRepository, actionLog, enforcement, revoker, actionLog, time.Now, ids)
	caseService := safetyapplication.NewCaseService(caseRepository, time.Now, ids)

	account, err := identitydomain.NewAccount("m-2", "+233550000112", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := accountRepository.Create(ctx, account); err != nil {
		t.Fatal(err)
	}

	// Tier-B case into review, then a 14-day suspension.
	report := safetydomain.ReconstituteReport("rep_1", "m-1", "m-2", safetydomain.CategoryHarassment, safetydomain.TierB, safetydomain.SurfaceRoom, "", "", safetydomain.StatusReceived, 1, time.Now())
	safetyCase, err := caseService.Open(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := caseService.Assign(ctx, safetyCase.ID(), "agent-1"); err != nil {
		t.Fatal(err)
	}
	if err := actionService.Apply(ctx, safetyCase.ID(), safetydomain.ActionSuspend14d, "agent-1"); err != nil {
		t.Fatalf("apply suspension: %v", err)
	}

	account, err = accountRepository.FindByID(ctx, "m-2")
	if err != nil {
		t.Fatal(err)
	}
	if account.Status() != identitydomain.AccountSuspended || account.SuspendedUntil() == nil {
		t.Fatalf("account = %#v", account)
	}
	if until := account.SuspendedUntil().Sub(time.Now()); until < 13*24*time.Hour || until > 15*24*time.Hour {
		t.Fatalf("suspension length = %v, want ~14 days", until)
	}
	if len(revoker.revoked) != 1 {
		t.Fatalf("revocations = %v", revoker.revoked)
	}
	priors, err := actionLog.CountForSubject(ctx, "m-2")
	if err != nil || priors != 1 {
		t.Fatalf("action log priors = %d, want 1", priors)
	}
	resolved, err := caseRepository.FindByID(ctx, safetyCase.ID())
	if err != nil || resolved.Status() != safetydomain.CaseResolved {
		t.Fatalf("case = %#v, %v", resolved, err)
	}
	if err := account.Usable(); err != identitydomain.ErrAccountNotUsable {
		t.Fatalf("suspended account usable = %v", err)
	}

	// Repeat Tier-B must ban, with the device blocklist entry.
	report2 := safetydomain.ReconstituteReport("rep_2", "m-3", "m-2", safetydomain.CategoryHarassment, safetydomain.TierB, safetydomain.SurfaceRoom, "", "", safetydomain.StatusReceived, 1, time.Now())
	case2, err := caseService.Open(ctx, report2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := caseService.Assign(ctx, case2.ID(), "agent-1"); err != nil {
		t.Fatal(err)
	}
	if err := actionService.Apply(ctx, case2.ID(), safetydomain.ActionBan, "agent-1"); err != nil {
		t.Fatalf("apply ban: %v", err)
	}
	account, _ = accountRepository.FindByID(ctx, "m-2")
	if account.Status() != identitydomain.AccountBlocked {
		t.Fatalf("status = %q, want blocked", account.Status())
	}
	var risk struct {
		Blocked bool `bson:"blocked"`
	}
	if err := database.Collection("device_risk").FindOne(ctx, bson.M{"_id": "m-2"}).Decode(&risk); err != nil || !risk.Blocked {
		t.Fatalf("device_risk = %#v, %v", risk, err)
	}

	// The ladder itself blocks off-ladder actions before any propagation.
	report3 := safetydomain.ReconstituteReport("rep_3", "m-4", "m-5", safetydomain.CategorySpam, safetydomain.TierC, safetydomain.SurfaceCircle, "", "", safetydomain.StatusReceived, 1, time.Now())
	case3, _ := caseService.Open(ctx, report3)
	if _, err := caseService.Assign(ctx, case3.ID(), "agent-2"); err != nil {
		t.Fatal(err)
	}
	if err := actionService.Apply(ctx, case3.ID(), safetydomain.ActionBan, "agent-2"); err == nil {
		t.Fatal("first tier-C ban must be rejected by the ladder")
	}
}
