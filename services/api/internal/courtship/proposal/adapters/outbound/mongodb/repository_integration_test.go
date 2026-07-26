//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	proposalmongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMutualDecisionReplayAndPrivacy(t *testing.T) {
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
	database := client.Database("proposal_test")
	repository := proposalmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	sender, recipient, detail := strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("3", 64)
	created, _ := domain.Create("proposal", domain.TypeExclusivity, sender, recipient, detail, now.Add(time.Hour), domain.Command{ID: "create", ActorKey: sender, Fingerprint: domain.Fingerprint("create", domain.ActionCreated, sender, 0), At: now})
	if err = repository.Create(ctx, created); err != nil {
		t.Fatal(err)
	}
	accept, _ := created.Accept(domain.Command{ID: "accept", ActorKey: recipient, Fingerprint: domain.Fingerprint("proposal", domain.ActionAccepted, recipient, 1), ExpectedRevision: 1, At: now})
	reject, _ := created.Reject(domain.Command{ID: "reject", ActorKey: recipient, Fingerprint: domain.Fingerprint("proposal", domain.ActionRejected, recipient, 1), ExpectedRevision: 1, At: now})
	type result struct {
		proposal domain.Proposal
		command  string
		err      error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for _, candidate := range []struct {
		p domain.Proposal
		c string
	}{{accept, "accept"}, {reject, "reject"}} {
		wg.Add(1)
		go func(v struct {
			p domain.Proposal
			c string
		}) {
			defer wg.Done()
			e := repository.Append(ctx, v.p, 1, v.c)
			results <- result{v.p, v.c, e}
		}(candidate)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	var winner result
	for r := range results {
		if r.err == nil {
			success++
			winner = r
		} else if errors.Is(r.err, application.ErrConcurrentChange) {
			conflict++
		} else {
			t.Fatal(r.err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	if err = repository.Append(ctx, winner.proposal, 1, winner.command); !errors.Is(err, application.ErrCommandApplied) {
		t.Fatalf("replay=%v", err)
	}
	stored, err := repository.Find(ctx, "proposal")
	if err != nil || stored.Status() != winner.proposal.Status() {
		t.Fatalf("stored=%v err=%v", stored.Status(), err)
	}
	var document bson.M
	if err = database.Collection("courtship_proposals").FindOne(ctx, bson.M{"_id": "proposal"}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"phone", "location", "raw-", "public", "relationshipstatus", "relationship_status"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy leak %q in %s", forbidden, raw)
		}
	}
}
