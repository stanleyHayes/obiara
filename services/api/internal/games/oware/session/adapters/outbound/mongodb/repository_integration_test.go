//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	sessionmongo "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	session "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMovesCASExpiryReplayAndPrivacy(t *testing.T) {
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
	db := client.Database("oware_session_test")
	r := sessionmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	game, e := session.Create("session-1", key(3), []string{key(1), key(2)}, time.Hour, now, session.Command{ID: "create", At: now})
	if e != nil || r.Create(ctx, game) != nil {
		t.Fatal(e)
	}
	a, _ := game.Move(key(1), 0, now.Add(time.Minute), session.Command{ID: "move-a", ExpectedRevision: 1, At: now.Add(time.Minute)})
	b, _ := game.Move(key(1), 1, now.Add(time.Minute), session.Command{ID: "move-b", ExpectedRevision: 1, At: now.Add(time.Minute)})
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, a, 1, "move-a") }()
	go func() { ch <- r.Append(ctx, b, 1, "move-b") }()
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
	saved, e := r.Find(ctx, "session-1")
	if e != nil || saved.Revision() != 2 || saved.Board().TotalSeeds() != 48 {
		t.Fatalf("%+v %v", saved, e)
	}
	winner := saved.Commands()[1].ID
	if e = r.Append(ctx, saved, 1, winner); !errors.Is(e, application.ErrApplied) {
		t.Fatalf("replay=%v", e)
	}
	expiring, e := session.Create("session-2", key(4), []string{key(5), key(6)}, time.Minute, now, session.Command{ID: "create-expiry", At: now})
	if e != nil || r.Create(ctx, expiring) != nil {
		t.Fatal(e)
	}
	before := expiring.Board().Houses()
	expired, e := expiring.Expire(now.Add(time.Minute), session.Command{ID: "expire", ExpectedRevision: 1, At: now.Add(time.Minute)})
	if e != nil || r.Append(ctx, expired, 1, "expire") != nil || expired.Board().Houses() != before {
		t.Fatal("expiry invented a move")
	}
	var raw bson.M
	if e = db.Collection("oware_sessions").FindOne(ctx, bson.M{"_id": "session-1"}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"room-private", "player-private", "matching", "trust", "visibility", "rating", "public", "listing"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
