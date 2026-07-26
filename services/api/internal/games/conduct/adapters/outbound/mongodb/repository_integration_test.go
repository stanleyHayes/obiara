//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	conductmongo "github.com/stanleyHayes/obiara/services/api/internal/games/conduct/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestEventUniquenessAppealConcurrencyAndPrivacy(t *testing.T) {
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
	db := client.Database("game_conduct_test")
	r := conductmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	a, _ := domain.Record("signal-a", key(1), key(2), key(3), domain.EventInactivity, now, domain.Command{ID: "record-a", At: now})
	b, _ := domain.Record("signal-b", key(1), key(2), key(3), domain.EventInactivity, now, domain.Command{ID: "record-b", At: now})
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, a) }()
	go func() { ch <- r.Create(ctx, b) }()
	ok, conflict := 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			ok++
		} else if errors.Is(x, application.ErrConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("%d %d", ok, conflict)
	}
	count, _ := db.Collection("game_conduct_signals").CountDocuments(ctx, bson.M{})
	if count != 1 {
		t.Fatalf("signals=%d", count)
	}
	var signal domain.Signal
	if x, e := r.Find(ctx, "signal-a"); e == nil {
		signal = x
	} else {
		signal, e = r.Find(ctx, "signal-b")
		if e != nil {
			t.Fatal(e)
		}
	}
	appealed, e := signal.Appeal(now.Add(time.Minute), domain.Command{ID: "appeal", ExpectedRevision: 1, At: now.Add(time.Minute)})
	if e != nil || r.Append(ctx, appealed, 1, "appeal") != nil {
		t.Fatal(e)
	}
	upheld, _ := appealed.Resolve(domain.AppealUpheld, now.Add(2*time.Minute), domain.Command{ID: "upheld", ExpectedRevision: 2, At: now.Add(2 * time.Minute)})
	overturned, _ := appealed.Resolve(domain.AppealOverturned, now.Add(2*time.Minute), domain.Command{ID: "overturned", ExpectedRevision: 2, At: now.Add(2 * time.Minute)})
	ch = make(chan error, 2)
	go func() { ch <- r.Append(ctx, upheld, 2, "upheld") }()
	go func() { ch <- r.Append(ctx, overturned, 2, "overturned") }()
	ok, conflict = 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			ok++
		} else if errors.Is(x, application.ErrConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("appeal %d %d", ok, conflict)
	}
	var raw bson.M
	if e = db.Collection("game_conduct_signals").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"game-private", "subject-private", "event-private", "skill", "winner", "loser", "winLoss", "rating", "popular", "matching", "trust", "visibility", "public", "listing", "freeText"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
