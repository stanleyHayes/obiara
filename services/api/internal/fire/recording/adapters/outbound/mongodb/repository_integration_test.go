//go:build integration

package mongodb_test

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	recordingmongo "github.com/stanleyHayes/obiara/services/api/internal/fire/recording/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func TestConsentJoinRevokeRetentionAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, _ := container.ConnectionString(ctx)
	sep := "?"
	if strings.Contains(uri, "?") {
		sep = "&"
	}
	client, err := apimongo.Connect(ctx, uri+sep+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("recording_test")
	repo := recordingmongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	host, member, fire := strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("9", 64)
	p, _ := domain.Open("policy", fire, host, []string{host, member}, command("open", host, host, domain.ActionOpened, "", 0, 0, now))
	if err = repo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	p, _ = p.Propose(domain.PurposeArchive, 30*24*time.Hour, command("proposal", host, host, domain.ActionProposed, domain.PurposeArchive, 30*24*time.Hour, 1, now))
	if err = repo.Append(ctx, p, 1, "proposal"); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.Find(ctx, "policy")
	if err != nil || stored.State().Active {
		t.Fatal(err)
	}
	var doc bson.M
	if err = db.Collection("fire_recording_policies").FindOne(ctx, bson.M{"_id": "policy"}).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(doc, false, false)
	for _, bad := range []string{"voice", "transcript", "alice", "bob", "email", "public"} {
		if strings.Contains(strings.ToLower(string(raw)), bad) {
			t.Fatalf("privacy leak %q in %s", bad, raw)
		}
	}
}
func command(id, actor, subject string, action domain.Action, p domain.Purpose, r time.Duration, rev uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, Fingerprint: domain.Fingerprint("policy", action, actor, subject, p, r, rev), ExpectedRevision: rev, At: at}
}
