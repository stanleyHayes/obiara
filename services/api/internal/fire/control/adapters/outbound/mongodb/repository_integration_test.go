//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	controlmongo "github.com/stanleyHayes/obiara/services/api/internal/fire/control/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAuthorizationReplayConcurrencyAndPrivacy(t *testing.T) {
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
	database := client.Database("fire_control_test")
	repository := controlmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	host, target, fireKey := strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("9", 64)
	opened, _ := domain.Open("control", fireKey, host, []string{target}, command("open", host, host, domain.ActionOpened, 0, now))
	if err = repository.Create(ctx, opened); err != nil {
		t.Fatal(err)
	}
	muted, _ := opened.Mute(target, command("mute", host, target, domain.ActionMuted, 1, now))
	ejected, _ := opened.Eject(target, command("eject", host, target, domain.ActionEjected, 1, now))
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, v := range []struct {
		f domain.Fire
		c string
	}{{muted, "mute"}, {ejected, "eject"}} {
		wg.Add(1)
		go func(v struct {
			f domain.Fire
			c string
		}) { defer wg.Done(); results <- repository.Append(ctx, v.f, 1, v.c) }(v)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	var winner struct {
		f domain.Fire
		c string
	}
	for e := range results {
		if e == nil {
			success++
		} else if errors.Is(e, application.ErrConcurrentChange) {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	_ = winner
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	stored, err := repository.Find(ctx, "control")
	if err != nil || stored.Revision() != 2 {
		t.Fatal(stored.Revision(), err)
	}
	var document bson.M
	if err = database.Collection("fire_control_events").FindOne(ctx, bson.M{"controlId": "control"}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"alice", "bob", "raw", "email", "phone", "accusation", "public"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy leak %q in %s", forbidden, raw)
		}
	}
}
func command(id, actor, target string, action domain.Action, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, ReasonCode: "moderation.test", Fingerprint: domain.Fingerprint("control", action, actor, target, "moderation.test", revision), ExpectedRevision: revision, At: at}
}
