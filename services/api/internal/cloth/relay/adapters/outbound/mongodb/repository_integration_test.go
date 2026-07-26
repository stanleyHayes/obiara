//go:build integration

package mongodb_test

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	relaymongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestQuestionRevokeRedactionPrivacy(t *testing.T) {
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
	db := client.Database("relay")
	r := relaymongo.NewRepository(db)
	_ = r.EnsureIndexes(ctx)
	v, _ := domain.New(key(9), []string{key(1), key(2)}, key(3))
	_ = r.Create(ctx, v)
	commands := []domain.Command{{ID: "submit-command-01", ActorKey: key(3), QuestionRef: "ref_questionabcdefghijklmnop", PromptRef: "ref_promptabcdefghijklmnop", At: time.Now()}, {ID: "grant-command-01", ActorKey: key(1), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: 1, At: time.Now()}, {ID: "grant-command-02", ActorKey: key(2), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: 2, At: time.Now()}, {ID: "revoke-command-01", ActorKey: key(1), QuestionRef: "ref_questionabcdefghijklmnop", ExpectedRevision: 3, At: time.Now()}}
	for i, c := range commands {
		old := v
		if i == 0 {
			v, _ = v.Submit(c)
		} else if i < 3 {
			v, _ = v.Grant(c)
		} else {
			v, _ = v.Revoke(c)
		}
		if e = r.Save(ctx, v, old.Revision(), c.ID); e != nil {
			t.Fatal(e)
		}
	}
	saved, _ := r.Find(ctx, key(9))
	if _, e = saved.Project(key(3), "ref_questionabcdefghijklmnop"); e != domain.ErrDenied {
		t.Fatalf("%v", e)
	}
	var raw bson.M
	_ = db.Collection("cloth_relays").FindOne(ctx, bson.M{}).Decode(&raw)
	wire, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw content", "full-thread", "public", "reverse"} {
		if strings.Contains(string(wire), bad) {
			t.Fatal(bad)
		}
	}
}
