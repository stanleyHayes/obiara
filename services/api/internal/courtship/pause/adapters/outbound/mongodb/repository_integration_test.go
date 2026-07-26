//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	pausemongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestImmediateSuspensionMutualResumeAndPrivacy(t *testing.T) {
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
	db := client.Database("pause_test")
	r := pausemongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	s, _ := domain.New(key(1), []string{key(2), key(3)})
	if e = r.Create(ctx, s); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	s, _ = s.Pause(domain.Command{ID: "pause-command-01", ActorKey: key(2), At: now})
	if e = r.Save(ctx, s, 0, "pause-command-01"); e != nil {
		t.Fatal(e)
	}
	saved, e := r.Find(ctx, key(1))
	if e != nil || saved.CanSend(key(3)) != domain.ErrSuspended {
		t.Fatalf("saved=%#v err=%v", saved, e)
	}
	saved, _ = saved.Acknowledge(domain.Command{ID: "ack-command-01", ActorKey: key(3), ExpectedRevision: 1, At: now})
	if e = r.Save(ctx, saved, 1, "ack-command-01"); e != nil {
		t.Fatal(e)
	}
	saved, _ = saved.Resume(domain.Command{ID: "resume-command-01", ActorKey: key(2), ExpectedRevision: 2, At: now})
	if e = r.Save(ctx, saved, 2, "resume-command-01"); e != nil {
		t.Fatal(e)
	}
	var raw bson.M
	if e = db.Collection("courtship_pause_stones").FindOne(ctx, bson.M{"_id": key(1)}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "message", "content", "publicActivity", "popularity"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
