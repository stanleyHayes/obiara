//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
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
	db := client.Database("suban_explanation_test")
	r := adapter.New(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)
	appeal, _ := domain.File(key(1), key(2), "event-1", domain.ReasonEventInaccurate, "file-1", now)
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, appeal) }()
	go func() { ch <- r.Create(ctx, appeal) }()
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
	resolved, _ := loaded.Resolve(domain.StatusOverturned, key(3), key(4), "resolve-1", now)
	ch = make(chan error, 2)
	go func() { ch <- r.Save(ctx, resolved, loaded.Revision(), "resolve-1") }()
	go func() { ch <- r.Save(ctx, resolved, loaded.Revision(), "resolve-1") }()
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
	if e = db.Collection("suban_appeals").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-evidence", "content", "message", "phone", "email", "score", "rank", "matching", "visibility", "thirdparty"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
}
