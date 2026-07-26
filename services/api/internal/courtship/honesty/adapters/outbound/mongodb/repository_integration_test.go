//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	honestymongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMutualVisibilityImmediateRevokeAndNoScore(t *testing.T) {
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
	db := client.Database("honesty_test")
	r := honestymongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	x, _ := domain.New(key(1), []string{key(2), key(3)})
	if e = r.Create(ctx, x); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	x, _ = x.Grant(domain.Command{"grant-command-01", key(2), 0, now})
	_ = r.Save(ctx, x, 0, "grant-command-01")
	x, _ = x.Grant(domain.Command{"grant-command-02", key(3), 1, now})
	_ = r.Save(ctx, x, 1, "grant-command-02")
	x, _ = x.Revoke(domain.Command{"revoke-command-01", key(2), 2, now})
	_ = r.Save(ctx, x, 2, "revoke-command-01")
	saved, e := r.Find(ctx, key(1))
	if e != nil || saved.Visible() {
		t.Fatalf("visible=%v err=%v", saved.Visible(), e)
	}
	var raw bson.M
	_ = db.Collection("courtship_honesty_ribbons").FindOne(ctx, bson.M{}).Decode(&raw)
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "score", "badge", "rank", "reputation", "public"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
