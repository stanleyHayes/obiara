//go:build integration

package mongodb_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	storymongo "github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func command(id string, rev uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ExpectedRevision: rev, At: at}
}
func TestAlternationConsentCASRedactionAndPrivacy(t *testing.T) {
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
	db := client.Database("anansesem_test")
	r := storymongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	story, e := domain.Create("story-1", key(3), "spider-path", []string{key(1), key(2)}, command("create", 0, now))
	if e != nil || r.Create(ctx, story) != nil {
		t.Fatal(e)
	}
	story, _ = story.Add("passage-1", key(1), "Once.", now, command("add-1", 1, now))
	if e = r.Append(ctx, story, 1, "add-1"); e != nil {
		t.Fatal(e)
	}
	story, _ = story.Add("passage-2", key(2), "Then.", now, command("add-2", 2, now))
	if e = r.Append(ctx, story, 2, "add-2"); e != nil {
		t.Fatal(e)
	}
	story, _ = story.Grant(key(1), command("grant-1", 3, now))
	if e = r.Append(ctx, story, 3, "grant-1"); e != nil {
		t.Fatal(e)
	}
	base := story
	a, _ := base.Grant(key(2), command("grant-a", 4, now))
	b, _ := base.Grant(key(2), command("grant-b", 4, now))
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, a, 4, "grant-a") }()
	go func() { ch <- r.Append(ctx, b, 4, "grant-b") }()
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
	story, e = r.Find(ctx, "story-1")
	if e != nil || len(story.Grants()) != 2 {
		t.Fatal(e)
	}
	story, e = story.Edit("passage-1", key(1), "Once again.", now, command("edit", 5, now))
	if e != nil || len(story.Grants()) != 0 || r.Append(ctx, story, 5, "edit") != nil {
		t.Fatal("edit did not invalidate")
	}
	story, _ = story.Grant(key(1), command("fresh-1", 6, now))
	_ = r.Append(ctx, story, 6, "fresh-1")
	story, _ = story.Grant(key(2), command("fresh-2", 7, now))
	_ = r.Append(ctx, story, 7, "fresh-2")
	story, edition, e := story.Publish(now, command("publish", 8, now))
	if e != nil || r.Append(ctx, story, 8, "publish") != nil {
		t.Fatal(e)
	}
	redacted, _ := json.Marshal(edition)
	for _, bad := range []string{key(1), key(2), key(3), "author", "room", "contact"} {
		if strings.Contains(strings.ToLower(string(redacted)), strings.ToLower(bad)) {
			t.Fatalf("edition leak %q: %s", bad, redacted)
		}
	}
	var raw bson.M
	if e = db.Collection("anansesem_stories").FindOne(ctx, bson.M{"_id": "story-1"}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"room-private", "author-private", "email", "phone", "contact"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("storage leak %q: %s", bad, encoded)
		}
	}
}
