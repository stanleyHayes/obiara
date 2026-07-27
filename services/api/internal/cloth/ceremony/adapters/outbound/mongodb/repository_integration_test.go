//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	ceremonymongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDualConfirmPublishAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, _ := container.ConnectionString(ctx)
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("ceremony_test")
	repository := ceremonymongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	a, b, d := strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("3", 64)
	opened, _ := domain.Open("ceremony", [2]string{a, b}, command("open", a, domain.ActionOpened, "", 0, now))
	if err = repository.Create(ctx, opened); err != nil {
		t.Fatal(err)
	}
	one, _ := opened.Confirm(command("c1", a, domain.ActionConfirmed, "", 1, now))
	two, _ := opened.Confirm(command("c2", b, domain.ActionConfirmed, "", 1, now))
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, v := range []struct {
		x domain.Ceremony
		c string
	}{{one, "c1"}, {two, "c2"}} {
		wg.Add(1)
		go func(v struct {
			x domain.Ceremony
			c string
		}) {
			defer wg.Done()
			results <- repository.Append(ctx, v.x, 1, v.c)
		}(v)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for e := range results {
		if e == nil {
			success++
		} else if errors.Is(e, application.ErrConcurrentChange) {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	stored, _ := repository.Find(ctx, "ceremony")
	missing := b
	if stored.State().Confirmations[0] == b {
		missing = a
	}
	completed, _ := stored.Confirm(command("retry", missing, domain.ActionConfirmed, "", stored.Revision(), now))
	if err = repository.Append(ctx, completed, stored.Revision(), "retry"); err != nil {
		t.Fatal(err)
	}
	stored, _ = repository.Find(ctx, "ceremony")
	proposed, _ := stored.ProposeAnnouncement(d, command("propose", a, domain.ActionAnnouncementProposed, d, stored.Revision(), now))
	_ = repository.Append(ctx, proposed, stored.Revision(), "propose")
	stored, _ = repository.Find(ctx, "ceremony")
	first, _ := stored.ConsentAnnouncement(command("a1", a, domain.ActionAnnouncementConsented, d, stored.Revision(), now))
	_ = repository.Append(ctx, first, stored.Revision(), "a1")
	stored, _ = repository.Find(ctx, "ceremony")
	second, _ := stored.ConsentAnnouncement(command("a2", b, domain.ActionAnnouncementConsented, d, stored.Revision(), now))
	_ = repository.Append(ctx, second, stored.Revision(), "a2")
	stored, _ = repository.Find(ctx, "ceremony")
	published, _ := stored.PublishAnnouncement(command("publish", a, domain.ActionAnnouncementPublished, d, stored.Revision(), now))
	if err = repository.Append(ctx, published, stored.Revision(), "publish"); err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err = database.Collection("cloth_ceremonies").FindOne(ctx, bson.M{"_id": "ceremony"}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"alice", "bob", "raw", "relationship", "reviewer", "material", "content", "publiclink"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy leak %q in %s", forbidden, raw)
		}
	}
}
func command(id, actor string, action domain.Action, destination string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, Fingerprint: domain.Fingerprint("ceremony", action, actor, destination, revision), ExpectedRevision: revision, At: at}
}
