//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	safetymongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestBlockReportCASAndPrivateStorage(t *testing.T) {
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
	db := client.Database("safety_test")
	r := safetymongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	s, _ := domain.New(key(9), []string{key(1), key(2)})
	if e = r.Create(ctx, s); e != nil {
		t.Fatal(e)
	}
	s, _ = s.Report(domain.Command{ID: "report-command-01", ActorKey: key(1), Category: domain.CategoryThreat, EvidenceRef: "enc_abcdefghijklmnopqrstuvwxyz", At: time.Unix(100, 0)})
	if e = r.Save(ctx, s, 0, "report-command-01"); e != nil {
		t.Fatal(e)
	}
	s, _ = s.Block(domain.Command{ID: "block-command-01", ActorKey: key(2), ExpectedRevision: 1, At: time.Unix(101, 0)})
	if e = r.Save(ctx, s, 1, "block-command-01"); e != nil {
		t.Fatal(e)
	}
	saved, e := r.Find(ctx, key(9))
	if e != nil || !saved.Blocked() || len(saved.Reviews()) != 1 {
		t.Fatalf("%v %v %v", saved.Blocked(), saved.Reviews(), e)
	}
	var raw bson.M
	_ = db.Collection("courtship_safety").FindOne(ctx, bson.M{}).Decode(&raw)
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "free-text", "reason", "accusation", "score", "public", "reverse"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
