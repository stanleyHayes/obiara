//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	workflowmongo "github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/adapters/outbound/token"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/domain"
)

type allowAll struct{}

func (allowAll) Require(context.Context, string, string, string, string) error { return nil }

type sequenceIDs struct{ value atomic.Uint64 }

func (ids *sequenceIDs) NewID() string {
	return "workflow-" + string(rune('a'+ids.value.Add(1)))
}

func TestInviteRedemptionAndModerationArePrivateAndConcurrentSafe(t *testing.T) {
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

	database := client.Database("obiara_circle_workflow_test")
	repository := workflowmongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.NewService(repository, allowAll{}, token.NewIssuer(), &sequenceIDs{}, func() time.Time { return now })
	invited, err := service.CreateInvite(ctx, application.Command{
		ID: "invite-command", CircleID: "circle-1", ActorID: "host-1", ReasonCode: "host_invite",
	}, time.Hour)
	if err != nil || invited.Token == "" {
		t.Fatalf("create invite=%+v err=%v", invited, err)
	}
	var storedInvite bson.M
	if err := database.Collection("circle_workflow_invites").FindOne(ctx, bson.M{"_id": invited.Invite.ID()}).Decode(&storedInvite); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(storedInvite, false, false)
	if strings.Contains(string(encoded), invited.Token) {
		t.Fatalf("raw invite token persisted: %s", encoded)
	}

	type redemption struct {
		result application.RequestResult
		err    error
	}
	start := make(chan struct{})
	outcomes := make(chan redemption, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			result, redeemErr := service.RedeemInvite(ctx, application.Command{
				ID: "redeem-" + string(rune('a'+index)), ActorID: "member-" + string(rune('a'+index)),
				ExpectedRevision: 1, ReasonCode: "invite_redeem",
			}, invited.Token)
			outcomes <- redemption{result: result, err: redeemErr}
		}(index)
	}
	close(start)
	group.Wait()
	close(outcomes)
	successes := 0
	var winner domain.Request
	for outcome := range outcomes {
		if outcome.err == nil {
			successes++
			winner = outcome.result.Request
		} else if !errors.Is(outcome.err, domain.ErrStaleRevision) {
			t.Fatalf("unexpected competing redemption error: %v", outcome.err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful redemptions=%d, want 1", successes)
	}

	approved, err := service.Approve(ctx, application.Command{
		ID: "approve-1", ActorID: "host-1", RequestID: winner.ID(),
		ExpectedRevision: 1, ReasonCode: "request_approved",
	})
	if err != nil || approved.Request.Status() != domain.RequestApproved {
		t.Fatalf("approve=%+v err=%v", approved, err)
	}
	replayed, err := service.Approve(ctx, application.Command{
		ID: "approve-1", ActorID: "host-1", RequestID: winner.ID(),
		ExpectedRevision: 1, ReasonCode: "request_approved",
	})
	if err != nil || !replayed.Replayed || replayed.Request.Revision() != 2 {
		t.Fatalf("approval replay=%+v err=%v", replayed, err)
	}

	current, err := repository.FindRequest(ctx, winner.ID())
	if err != nil {
		t.Fatal(err)
	}
	first, err := current.Expel(domain.Command{
		ID: "expel-a", ActorID: "host-1", ExpectedRevision: 2,
		ReasonCode: "policy_violation", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := current.Expel(domain.Command{
		ID: "expel-b", ActorID: "host-2", ExpectedRevision: 2,
		ReasonCode: "safety_violation", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	saveErrors := make(chan error, 2)
	go func() { saveErrors <- repository.SaveRequest(ctx, first, 2, "expel-a") }()
	go func() { saveErrors <- repository.SaveRequest(ctx, second, 2, "expel-b") }()
	var saved, conflicted int
	for range 2 {
		saveErr := <-saveErrors
		switch {
		case saveErr == nil:
			saved++
		case errors.Is(saveErr, application.ErrOptimisticConflict):
			conflicted++
		default:
			t.Fatalf("concurrent expulsion error: %v", saveErr)
		}
	}
	if saved != 1 || conflicted != 1 {
		t.Fatalf("saved=%d conflicted=%d", saved, conflicted)
	}
	final, err := repository.FindRequest(ctx, winner.ID())
	if err != nil || final.Status() != domain.RequestExpelled || final.Revision() != 3 || len(final.Events()) != 3 {
		t.Fatalf("final request status=%s revision=%d events=%d err=%v", final.Status(), final.Revision(), len(final.Events()), err)
	}
}
