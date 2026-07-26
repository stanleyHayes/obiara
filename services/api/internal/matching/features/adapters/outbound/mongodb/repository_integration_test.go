//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	featuremongo "github.com/stanleyHayes/obiara/services/api/internal/matching/features/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/application"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type authority struct{}

func (authority) RequireMember(context.Context, string, string) error       { return nil }
func (authority) RequirePair(context.Context, string, string, string) error { return nil }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "decision:" + string(rune('0'+i.n)) }

func TestPairIntersectionSnapshotAndImmediateWithdrawal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, err := box.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("matching_features_test")
	repo := featuremongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	def, _ := domain.NewDefinition("shared.rituals", 1, "matching.compatibility", now.Add(-time.Hour))
	if err = repo.Put(ctx, def); err != nil {
		t.Fatal(err)
	}
	keyer, _ := privacy.NewKeyer([]byte("01234567890123456789012345678901"))
	idSource := &ids{}
	service := application.NewService(repo, repo, repo, authority{}, keyer, idSource, func() time.Time { return now })
	for i, member := range []string{"alice@example.invalid", "bob@example.invalid"} {
		_, err = service.Grant(ctx, application.GrantCommand{Actor: member, Member: member, Feature: def.Key, Purpose: def.Purpose, FeatureVersion: 1, GrantVersion: uint64(i + 1), CommandID: "grant:" + string(rune('a'+i))})
		if err != nil {
			t.Fatal(err)
		}
	}
	decision, err := service.Decide(ctx, application.PairRequest{Actor: "pair-owner", First: "alice@example.invalid", Second: "bob@example.invalid"})
	if err != nil || len(decision.Features) != 1 {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	ok, err := service.Revalidate(ctx, decision.ID)
	if err != nil || !ok {
		t.Fatalf("initial revalidation=%v err=%v", ok, err)
	}
	_, err = service.Withdraw(ctx, application.WithdrawCommand{Actor: "alice@example.invalid", Member: "alice@example.invalid", Feature: def.Key, CommandID: "withdraw:a", ExpectedRevision: 1})
	if err != nil {
		t.Fatal(err)
	}
	ok, err = service.Revalidate(ctx, decision.ID)
	if err != nil || ok {
		t.Fatalf("withdrawn revalidation=%v err=%v", ok, err)
	}

	for _, collection := range []string{"matching_feature_grants", "matching_feature_decisions"} {
		cur, e := db.Collection(collection).Find(ctx, bson.M{})
		if e != nil {
			t.Fatal(e)
		}
		var docs []bson.M
		if e = cur.All(ctx, &docs); e != nil {
			t.Fatal(e)
		}
		raw, _ := bson.MarshalExtJSON(docs, false, false)
		value := strings.ToLower(string(raw))
		for _, forbidden := range []string{"alice@example.invalid", "bob@example.invalid", "rawcontent", "sensitiveinference", "model", "vendor", "embedding", "retroactive"} {
			if strings.Contains(value, forbidden) {
				t.Fatalf("%s leaked %q: %s", collection, forbidden, value)
			}
		}
	}
}
