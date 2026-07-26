//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	sourcemongo "github.com/stanleyHayes/obiara/services/api/internal/seed/source/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

func opaque(n int) string { return fmt.Sprintf("%064x", n) }
func TestBoundsWithdrawalConcurrencyAndPrivacyProof(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	database := client.Database("obiara_seed_source_test")
	repository := sourcemongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	candidates := make([]string, domain.MaxCandidates)
	for i := range candidates {
		candidates[i] = opaque(i + 10)
	}
	rawRequester, rawSource, rawCandidate := "member:private", "circle:private", "candidate:private"
	opened, err := domain.Open("request-1", opaque(1), domain.Source{Type: domain.SourceCircle, Key: opaque(2)}, candidates, now.Add(time.Hour), domain.Command{ID: "open-1", ActorKey: opaque(1), ReasonCode: "user_requested", At: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, opened); err != nil {
		t.Fatal(err)
	}
	var raw bson.M
	if err := database.Collection("seed_source_requests").FindOne(ctx, bson.M{"_id": "request-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{rawRequester, rawSource, rawCandidate, "profile", "reverseGraph", "memberList"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("privacy leak %q in %s", forbidden, encoded)
		}
	}
	stored, err := repository.Find(ctx, "request-1")
	if err != nil || len(stored.CandidateIDs()) != domain.MaxCandidates {
		t.Fatalf("stored candidates=%d err=%v", len(stored.CandidateIDs()), err)
	}

	withdrawn, err := stored.Withdraw(domain.Command{ID: "withdraw-1", ActorKey: opaque(1), ExpectedRevision: 1, ReasonCode: "user_requested", At: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Append(ctx, withdrawn, 1, "withdraw-1"); err != nil {
		t.Fatal(err)
	}
	final, err := repository.Find(ctx, "request-1")
	if err != nil || final.Status() != domain.StatusWithdrawn || final.Revision() != 2 {
		t.Fatalf("withdrawal status=%s revision=%d err=%v", final.Status(), final.Revision(), err)
	}
	if final.Events()[0] != opened.Events()[0] {
		t.Fatal("opening audit event was rewritten")
	}

	second, err := domain.Open("request-2", opaque(1), domain.Source{Type: domain.SourceCohort, Key: opaque(3)}, nil, now.Add(time.Hour), domain.Command{ID: "open-2", ActorKey: opaque(1), ReasonCode: "policy_approved", At: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, second); err != nil {
		t.Fatal(err)
	}
	expired, _ := second.Expire(domain.Command{ID: "expire-2", ActorKey: opaque(4), ExpectedRevision: 1, ReasonCode: "ttl_expired", At: now.Add(time.Hour)})
	withdrawn2, _ := second.Withdraw(domain.Command{ID: "withdraw-2", ActorKey: opaque(1), ExpectedRevision: 1, ReasonCode: "user_requested", At: now.Add(time.Hour)})
	results := make(chan error, 2)
	go func() { results <- repository.Append(ctx, expired, 1, "expire-2") }()
	go func() { results <- repository.Append(ctx, withdrawn2, 1, "withdraw-2") }()
	var saved, conflicted int
	for range 2 {
		if saveErr := <-results; saveErr == nil {
			saved++
		} else if errors.Is(saveErr, application.ErrOptimisticConflict) {
			conflicted++
		} else {
			t.Fatal(saveErr)
		}
	}
	if saved != 1 || conflicted != 1 {
		t.Fatalf("saved=%d conflicted=%d", saved, conflicted)
	}
}
