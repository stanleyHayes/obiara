//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
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
	db := client.Database("matchmaker_test")
	r := adapter.New(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	license := domain.License{ID: "license.gh", MatchmakerKey: key(2), Jurisdiction: "ghana", Version: 1, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), MinimumFeePesewas: 1000, MaximumFeePesewas: 5000}
	terms := domain.Terms{ID: "terms.1", Version: 1, TotalFeePesewas: 1000, Milestones: []domain.Milestone{{ID: "consult", FeePesewas: 1000}}}
	x, _ := domain.Book(key(1), key(3), license, terms, "book-1", now)
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, x) }()
	go func() { ch <- r.Create(ctx, x) }()
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
	loaded, e := r.Find(ctx, key(1))
	if e != nil {
		t.Fatal(e)
	}
	curated, _ := loaded.Curate("curate-1", key(4), now)
	ch = make(chan error, 2)
	go func() { ch <- r.Save(ctx, curated, loaded.Revision(), "curate-1") }()
	go func() { ch <- r.Save(ctx, curated, loaded.Revision(), "curate-1") }()
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
	if e = db.Collection("matchmaker_engagements").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "address", "contact", "rank", "visibility", "seed", "globalbrowse", "rating"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
