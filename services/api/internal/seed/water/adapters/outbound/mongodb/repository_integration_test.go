//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	watermongo "github.com/stanleyHayes/obiara/services/api/internal/seed/water/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMutualitySingleRoomAndPrivacy(t *testing.T) {
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
	client, e := apimongo.Connect(ctx, uri+sep+"directConnection=true")
	if e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("seed_water_test")
	r := watermongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, e := domain.Start("water-1", []string{key(1), key(2)}, domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Create(ctx, w); e != nil {
		t.Fatal(e)
	}
	var raw bson.M
	if e = db.Collection("seed_mutual_water").FindOne(ctx, bson.M{"_id": "water-1"}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	if _, exists := raw["roomKey"]; exists {
		t.Fatal("room exists before mutuality")
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "publicActivity", "reverseLookup"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatal("privacy leak")
		}
	}
	a, _ := w.Water(domain.Command{ID: "second-a", ActorKey: key(2), ReasonCode: "member_watered", ExpectedRevision: 1, At: now}, key(9))
	b, _ := w.Water(domain.Command{ID: "second-b", ActorKey: key(2), ReasonCode: "member_watered", ExpectedRevision: 1, At: now}, key(10))
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, a, 1, "second-a") }()
	go func() { ch <- r.Append(ctx, b, 1, "second-b") }()
	saved, conflict := 0, 0
	for range 2 {
		if x := <-ch; x == nil {
			saved++
		} else if errors.Is(x, application.ErrOptimisticConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if saved != 1 || conflict != 1 {
		t.Fatalf("%d %d", saved, conflict)
	}
	final, e := r.Find(ctx, "water-1")
	if e != nil || final.Status() != domain.StatusRoomCreated || final.RoomKey() == "" || final.Revision() != 2 {
		t.Fatalf("%+v %v", final, e)
	}
	count, e := db.Collection("seed_mutual_water").CountDocuments(ctx, bson.M{"roomKey": bson.M{"$exists": true, "$ne": ""}})
	if e != nil || count != 1 {
		t.Fatalf("rooms=%d err=%v", count, e)
	}
}
