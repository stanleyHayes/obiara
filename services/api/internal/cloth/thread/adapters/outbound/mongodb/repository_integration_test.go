//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	threadmongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestOneTimeConcurrentIssuanceAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
	db := client.Database("thread_test")
	r := threadmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	base, _ := domain.New(key(9), []string{key(1), key(2)})
	if e = r.Create(ctx, base); e != nil {
		t.Fatal(e)
	}
	a, _ := base.Issue(domain.Command{ID: "thread-command-01", ActorKey: key(1), RevealRef: "ref_revealabcdefghijklmnop", RecipeRef: "ref_recipeabcdefghijklmnop", BandVersion: 1, At: time.Unix(100, 0)})
	b, _ := base.Issue(domain.Command{ID: "thread-command-02", ActorKey: key(2), RevealRef: "ref_revealabcdefghijklmnop", RecipeRef: "ref_recipeabcdefghijklmnop", BandVersion: 1, At: time.Unix(100, 0)})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, candidate := range []domain.Thread{a, b} {
		wg.Add(1)
		go func(v domain.Thread) { defer wg.Done(); errs <- r.Save(ctx, v, 0, v.Commands()[0].ID) }(candidate)
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successes=%d", success)
	}
	saved, e := r.Find(ctx, key(9))
	if e != nil || saved.Revision() != 1 {
		t.Fatalf("%d %v", saved.Revision(), e)
	}
	var raw bson.M
	_ = db.Collection("cloth_threads").FindOne(ctx, bson.M{}).Decode(&raw)
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "raw response", "relationship", "purchase", "bypass", "public", "reverse"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
