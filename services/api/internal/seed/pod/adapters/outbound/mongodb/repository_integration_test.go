//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	podmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestCapPrivacyAndConcurrentPlayback(t *testing.T) {
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
	db := client.Database("seed_pod_test")
	r := podmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	rs := make([]string, domain.MaxRecipients)
	for i := range rs {
		rs[i] = key(i + 3)
	}
	p, e := domain.Create("pod-1", key(1), key(2), rs, now.Add(time.Hour), domain.Command{ID: "create-1", ActorKey: key(1), ReasonCode: "user_requested", At: now})
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Create(ctx, p); e != nil {
		t.Fatal(e)
	}
	var raw bson.M
	if e = db.Collection("seed_pods").FindOne(ctx, bson.M{"_id": "pod-1"}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"raw-member", "raw-media", "profile", "globalQueue", "reverseDiscovery"} {
		if strings.Contains(string(encoded), bad) {
			t.Fatalf("leak %q", bad)
		}
	}
	a, _ := p.Play(domain.Command{ID: "play-a", ActorKey: key(3), ReasonCode: "user_requested", ExpectedRevision: 1, At: now.Add(time.Minute)})
	b, _ := p.Play(domain.Command{ID: "play-b", ActorKey: key(4), ReasonCode: "user_requested", ExpectedRevision: 1, At: now.Add(time.Minute)})
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, a, 1, "play-a") }()
	go func() { ch <- r.Append(ctx, b, 1, "play-b") }()
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
	final, e := r.Find(ctx, "pod-1")
	if e != nil || len(final.RecipientKeys()) != domain.MaxRecipients || final.Revision() != 2 {
		t.Fatalf("recipients=%d revision=%d err=%v", len(final.RecipientKeys()), final.Revision(), e)
	}
}
