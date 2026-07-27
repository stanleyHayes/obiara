//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/application"
	"github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrencyIdempotencyPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13")
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, e := apimongo.Connect(ctx, uri)
	if e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("market_pack_test")
	repo := adapter.New(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 3, 0, 0, 0, time.UTC)
	master, _ := domain.NewMaster(domain.MasterSpec{ID: "ghana.master", Version: 1, Entries: []domain.MasterEntry{{Key: "hello.world", Text: "Hello {name}."}}})
	pack, _ := domain.Propose(key(1), "GH", "tw-GH", key(2), 1, master, []domain.Translation{{Key: "hello.world", Text: "Maakye {name}."}}, "propose-1", now)
	ch := make(chan error, 2)
	go func() { ch <- repo.Create(ctx, pack) }()
	go func() { ch <- repo.Create(ctx, pack) }()
	saved, replay := 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("%d %d", saved, replay)
	}
	review := domain.Review{Stage: domain.StageProfessional, ReviewerKey: key(3), Checks: []domain.Check{domain.CheckMeaning, domain.CheckVoice}, EvidenceRef: key(4), ReviewedAt: now}
	next, _ := pack.AddReview(review, "review-1")
	ch = make(chan error, 2)
	go func() { ch <- repo.Save(ctx, next, pack.Revision(), "review-1") }()
	go func() { ch <- repo.Save(ctx, next, pack.Revision(), "review-1") }()
	saved, replay = 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			saved++
		} else if errors.Is(e, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(e)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("save %d %d", saved, replay)
	}
	var raw bson.M
	if e = db.Collection("market_pack_proposals").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "memberid", "machine_translation", "prompt", "conversation", "deployment", "activation", "rank", "score"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
}
