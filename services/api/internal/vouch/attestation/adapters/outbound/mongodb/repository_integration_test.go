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
	attestationmongo "github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/domain"
)

type allowAll struct{}

func (allowAll) Require(context.Context, string, string, string) error { return nil }

type fixedPolicy struct{}

func (fixedPolicy) Validate(_ context.Context, scope string, units uint8) (string, error) {
	if scope != "circle" || units == 0 || units > 40 {
		return "", application.ErrPolicyDenied
	}
	return "reputation-policy-v1", nil
}

type fixedID struct{}

func (fixedID) NewID() string { return "attestation-1" }

func TestHistoryIsAppendOnlyConcurrentAndPrivacyMinimal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
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
	database := client.Database("obiara_vouch_attestation_test")
	repository := attestationmongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("v", 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.NewService(repository, allowAll{}, keyer, fixedPolicy{}, fixedID{}, func() time.Time { return now })
	rawSubject, rawVoucher, rawScope := "subject:private", "voucher:private", "circle:private"
	proposed, err := service.Propose(ctx, application.Command{
		ID: "propose-1", ActorID: rawVoucher, ReasonCode: "voucher_proposed",
	}, application.Proposal{
		SubjectID: rawSubject, VoucherID: rawVoucher, ScopeKind: "circle",
		ScopeID: rawScope, StakeUnits: 30, TTL: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	consented, err := service.Consent(ctx, application.Command{
		ID: "consent-1", AttestationID: proposed.Attestation.ID(), ActorID: rawVoucher,
		ExpectedRevision: 1, ReasonCode: "voucher_consented",
	})
	if err != nil || consented.Attestation.Status() != domain.StatusActive {
		t.Fatalf("consent=%+v err=%v", consented, err)
	}
	var before bson.M
	if err := database.Collection("vouch_attestations").FindOne(ctx, bson.M{"_id": "attestation-1"}).Decode(&before); err != nil {
		t.Fatal(err)
	}
	beforeEvents, _ := bson.MarshalExtJSON(before["events"], false, false)
	beforeCommands, _ := bson.MarshalExtJSON(before["commands"], false, false)

	operatorA, operatorB := "operator:a:private", "operator:b:private"
	keyA, _ := keyer.Key("vouch-attestation:actor", operatorA)
	keyB, _ := keyer.Key("vouch-attestation:actor", operatorB)
	first, err := consented.Attestation.Revoke(domain.Command{
		ID: "revoke-a", ActorKey: keyA, ExpectedRevision: 2, ReasonCode: "policy_revoked", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := consented.Attestation.Revoke(domain.Command{
		ID: "revoke-b", ActorKey: keyB, ExpectedRevision: 2, ReasonCode: "voucher_revoked", At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	go func() { results <- repository.Append(ctx, first, 2, "revoke-a") }()
	go func() { results <- repository.Append(ctx, second, 2, "revoke-b") }()
	var saved, conflicted int
	for range 2 {
		switch saveErr := <-results; {
		case saveErr == nil:
			saved++
		case errors.Is(saveErr, application.ErrOptimisticConflict):
			conflicted++
		default:
			t.Fatalf("append race: %v", saveErr)
		}
	}
	if saved != 1 || conflicted != 1 {
		t.Fatalf("saved=%d conflicted=%d", saved, conflicted)
	}
	final, err := repository.Find(ctx, "attestation-1")
	if err != nil || final.Status() != domain.StatusRevoked || final.Revision() != 3 ||
		final.Provenance() == nil || final.Provenance().PolicyVersion != "reputation-policy-v1" {
		t.Fatalf("final=%+v err=%v", final, err)
	}
	var after bson.M
	if err := database.Collection("vouch_attestations").FindOne(ctx, bson.M{"_id": "attestation-1"}).Decode(&after); err != nil {
		t.Fatal(err)
	}
	afterEvents := after["events"].(bson.A)
	afterCommands := after["commands"].(bson.A)
	prefixEvents, _ := bson.MarshalExtJSON(afterEvents[:2], false, false)
	prefixCommands, _ := bson.MarshalExtJSON(afterCommands[:2], false, false)
	if string(prefixEvents) != string(beforeEvents) || string(prefixCommands) != string(beforeCommands) {
		t.Fatal("revocation rewrote immutable history")
	}
	encoded, _ := bson.MarshalExtJSON(after, false, false)
	for _, forbidden := range []string{
		rawSubject, rawVoucher, rawScope, operatorA, operatorB,
		`"money"`, `"payment"`, `"escrow"`, `"token"`, `"balance"`, `"graph"`,
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("forbidden %q in %s", forbidden, encoded)
		}
	}
}
