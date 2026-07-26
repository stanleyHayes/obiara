//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	gatemongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDualConsentRevokeConcurrencyAndPrivacy(t *testing.T) {
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
	database := client.Database("cloth_gate_test")
	repository := gatemongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	a, b := strings.Repeat("1", 64), strings.Repeat("2", 64)
	cap := domain.Capability{strings.Repeat("3", 64), strings.Repeat("4", 64), strings.Repeat("5", 64)}
	opened, _ := domain.Open("policy", domain.VersionV1, [2]string{a, b}, domain.Command{ID: "open", ActorKey: a, Fingerprint: domain.Fingerprint("policy", domain.ActionOpened, a, domain.Capability{}, 0), At: now})
	if err = repository.Create(ctx, opened); err != nil {
		t.Fatal(err)
	}
	grantA, _ := opened.Grant(cap, domain.Command{ID: "grant-a", ActorKey: a, Fingerprint: domain.Fingerprint("policy", domain.ActionGranted, a, cap, 1), ExpectedRevision: 1, At: now})
	grantB, _ := opened.Grant(cap, domain.Command{ID: "grant-b", ActorKey: b, Fingerprint: domain.Fingerprint("policy", domain.ActionGranted, b, cap, 1), ExpectedRevision: 1, At: now})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, candidate := range []struct {
		p domain.Policy
		c string
	}{{grantA, "grant-a"}, {grantB, "grant-b"}} {
		wg.Add(1)
		go func(v struct {
			p domain.Policy
			c string
		}) {
			defer wg.Done()
			results <- repository.Append(ctx, v.p, 1, v.c)
		}(candidate)
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
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	stored, err := repository.Find(ctx, "policy")
	if err != nil || stored.Allows(cap) {
		t.Fatalf("unilateral effective=%v err=%v", stored.Allows(cap), err)
	}
	missing := b
	if stored.State().Grants[0].MemberKey == b {
		missing = a
	}
	second, _ := stored.Grant(cap, domain.Command{ID: "retry-other", ActorKey: missing, Fingerprint: domain.Fingerprint("policy", domain.ActionGranted, missing, cap, stored.Revision()), ExpectedRevision: stored.Revision(), At: now})
	if err = repository.Append(ctx, second, stored.Revision(), "retry-other"); err != nil {
		t.Fatal(err)
	}
	stored, _ = repository.Find(ctx, "policy")
	if !stored.Allows(cap) {
		t.Fatal("dual consent did not allow")
	}
	revoked, _ := stored.Revoke(cap, domain.Command{ID: "revoke", ActorKey: a, Fingerprint: domain.Fingerprint("policy", domain.ActionRevoked, a, cap, stored.Revision()), ExpectedRevision: stored.Revision(), At: now})
	if err = repository.Append(ctx, revoked, stored.Revision(), "revoke"); err != nil {
		t.Fatal(err)
	}
	stored, _ = repository.Find(ctx, "policy")
	if stored.Allows(cap) {
		t.Fatal("revoke not immediate")
	}
	var document bson.M
	if err = database.Collection("cloth_gate_policies").FindOne(ctx, bson.M{"_id": "policy"}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"alice", "bob", "raw-", "public", "circle", "content", "link"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy/inheritance leak %q in %s", forbidden, raw)
		}
	}
}
