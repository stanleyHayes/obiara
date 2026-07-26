//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	assistedmongo "github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/application"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/domain"
)

type allowManual struct{}

func (allowManual) Require(context.Context, string, string, string) error { return nil }

type sequenceIDs struct{ value atomic.Uint64 }

func (ids *sequenceIDs) NewID() string {
	return "assisted-vouch-" + string(rune('a'+ids.value.Add(1)))
}

func TestManualVouchIsConcurrentSafeAndContainsNoRawIdentity(t *testing.T) {
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

	database := client.Database("obiara_assisted_vouch_test")
	repository := assistedmongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("a", 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.NewService(repository, allowManual{}, keyer, &sequenceIDs{}, func() time.Time { return now })
	rawSubject := "member:subject:private"
	rawRequester := "member:requester:private"
	rawVoucher := "member:voucher:private"
	rawOperatorA := "operator:a:private"
	rawOperatorB := "operator:b:private"
	requested, err := service.Request(ctx, application.Command{
		ID: "request-1", ActorID: rawRequester, ReasonCode: "assisted_request",
	}, application.CreateInput{
		SubjectID: rawSubject, RequesterID: rawRequester, VoucherID: rawVoucher, TTL: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	consented, err := service.Consent(ctx, application.Command{
		ID: "consent-1", RequestID: requested.Request.ID(), ActorID: rawVoucher,
		ExpectedRevision: 1, ReasonCode: "voucher_consented",
	})
	if err != nil || consented.Request.Status() != domain.StatusConsented {
		t.Fatalf("consent=%+v err=%v", consented, err)
	}

	operatorKeyA, _ := keyer.Key("vouch:actor", rawOperatorA)
	operatorKeyB, _ := keyer.Key("vouch:actor", rawOperatorB)
	approved, err := consented.Request.Decide(domain.DecisionApprove, domain.Command{
		ID: "decision-a", ActorKey: operatorKeyA, ExpectedRevision: 2,
		ReasonCode: "identity_confirmed", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	declined, err := consented.Request.Decide(domain.DecisionDecline, domain.Command{
		ID: "decision-b", ActorKey: operatorKeyB, ExpectedRevision: 2,
		ReasonCode: "identity_unconfirmed", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	saveErrors := make(chan error, 2)
	go func() { saveErrors <- repository.Save(ctx, approved, 2, "decision-a") }()
	go func() { saveErrors <- repository.Save(ctx, declined, 2, "decision-b") }()
	var saved, conflicted int
	for range 2 {
		switch saveErr := <-saveErrors; {
		case saveErr == nil:
			saved++
		case errors.Is(saveErr, application.ErrOptimisticConflict):
			conflicted++
		default:
			t.Fatalf("decision race error: %v", saveErr)
		}
	}
	if saved != 1 || conflicted != 1 {
		t.Fatalf("saved=%d conflicted=%d", saved, conflicted)
	}
	final, err := repository.Find(ctx, requested.Request.ID())
	if err != nil || final.Revision() != 3 || final.Outcome() == nil ||
		final.Outcome().Provenance != "manual_assisted" || len(final.Events()) != 3 {
		t.Fatalf("final status=%s revision=%d outcome=%+v events=%d err=%v",
			final.Status(), final.Revision(), final.Outcome(), len(final.Events()), err)
	}

	last := final.Events()[2]
	replayOperator := rawOperatorA
	replayDecision := domain.DecisionApprove
	if last.CommandID == "decision-b" {
		replayOperator, replayDecision = rawOperatorB, domain.DecisionDecline
	}
	replayed, err := service.Decide(ctx, application.Command{
		ID: last.CommandID, RequestID: final.ID(), ActorID: replayOperator,
		ExpectedRevision: 2, ReasonCode: last.ReasonCode,
	}, replayDecision)
	if err != nil || !replayed.Replayed || replayed.Request.Revision() != 3 {
		t.Fatalf("replay=%+v err=%v", replayed, err)
	}

	var stored bson.M
	if err := database.Collection("assisted_vouch_requests").FindOne(ctx, bson.M{"_id": final.ID()}).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(stored, false, false)
	for _, forbidden := range []string{
		rawSubject, rawRequester, rawVoucher, rawOperatorA, rawOperatorB,
		`"score"`, `"stake"`, `"money"`, `"payment"`, `"graph"`,
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("stored assisted vouch leaked or productized %q: %s", forbidden, encoded)
		}
	}
}
