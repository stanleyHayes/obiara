//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	closuremongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestTerminalTimeoutCASAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("closure_test")
	r := closuremongo.NewRepository(db)
	if err = r.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	c, _ := domain.New(key(9), []string{key(1), key(2)}, time.Unix(100, 0))
	if err = r.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	c, _ = c.CloseInactive(domain.Command{"timeout-command-01", "", 0, time.Unix(200, 0)}, time.Minute)
	if err = r.Save(ctx, c, 0, "timeout-command-01"); err != nil {
		t.Fatal(err)
	}
	saved, err := r.Find(ctx, key(9))
	if err != nil || saved.Status() != domain.StatusClosed {
		t.Fatalf("%s %v", saved.Status(), err)
	}
	var raw bson.M
	_ = db.Collection("courtship_closures").FindOne(ctx, bson.M{}).Decode(&raw)
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-room", "raw-member", "reason", "readReceipt", "accusation", "score", "public", "reputation"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
