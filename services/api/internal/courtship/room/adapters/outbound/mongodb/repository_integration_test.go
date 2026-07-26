//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	roommongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/room/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestAppendProjectionConcurrencyAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	sep := "?"
	if strings.Contains(uri, "?") {
		sep = "&"
	}
	cl, e := apimongo.Connect(ctx, uri+sep+"directConnection=true")
	if e != nil {
		t.Fatal(e)
	}
	defer cl.Disconnect(context.Background())
	db := cl.Database("courtship_room_test")
	repo := roommongo.NewRepository(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	r, e := domain.Open("room-1", []string{key(1), key(2)}, domain.Command{ID: "open", ActorKey: key(1), ReasonCode: "member_action", At: now})
	if e != nil {
		t.Fatal(e)
	}
	if e = repo.Create(ctx, r); e != nil {
		t.Fatal(e)
	}
	var raw bson.M
	db.Collection("courtship_rooms").FindOne(ctx, bson.M{"_id": "room-1"}).Decode(&raw)
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "publicActivity", "popularity", "reverseLookup"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatal("privacy leak")
		}
	}
	a, _ := r.Message(key(9), domain.Command{ID: "a", ActorKey: key(1), ReasonCode: "member_action", ExpectedRevision: 1, At: now})
	b, _ := r.Message(key(10), domain.Command{ID: "b", ActorKey: key(2), ReasonCode: "member_action", ExpectedRevision: 1, At: now})
	ch := make(chan error, 2)
	go func() { ch <- repo.Append(ctx, a, 1, "a") }()
	go func() { ch <- repo.Append(ctx, b, 1, "b") }()
	ok, conflict := 0, 0
	for range 2 {
		if x := <-ch; x == nil {
			ok++
		} else if errors.Is(x, application.ErrOptimisticConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("%d %d", ok, conflict)
	}
	final, e := repo.Find(ctx, "room-1")
	if e != nil || final.Projection().MessageCount != 1 || final.Revision() != 2 {
		t.Fatalf("%+v %v", final.Projection(), e)
	}
}
