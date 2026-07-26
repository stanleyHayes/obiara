//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	lifemongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestArchiveDeletePrivacy(t *testing.T) {
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
	db := client.Database("life")
	r := lifemongo.NewRepository(db)
	_ = r.EnsureIndexes(ctx)
	v, _ := domain.New(key(9), []string{key(1), key(2)}, domain.Provenance{BandVersion: 1, RecipeRef: "ref_recipeabcdefghijklmnop"})
	_ = r.Create(ctx, v)
	v, _ = v.Archive(domain.Command{ID: "archive-command-01", ActorKey: key(1), ArchiveRef: "ref_archiveabcdefghijklmnop", At: time.Unix(100, 0)})
	_ = r.Save(ctx, v, 0, "archive-command-01")
	v, _ = v.Delete(domain.Command{ID: "delete-command-01", ActorKey: key(1), ReceiptKey: key(8), ExpectedRevision: 1, At: time.Unix(101, 0)})
	_ = r.Save(ctx, v, 1, "delete-command-01")
	saved, e := r.Find(ctx, key(9))
	if e != nil || saved.Status() != domain.StatusDeleted {
		t.Fatal(e)
	}
	var raw bson.M
	_ = db.Collection("cloth_lifecycles").FindOne(ctx, bson.M{}).Decode(&raw)
	wire, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw content", "public", "reverse"} {
		if strings.Contains(string(wire), bad) {
			t.Fatal(bad)
		}
	}
}
