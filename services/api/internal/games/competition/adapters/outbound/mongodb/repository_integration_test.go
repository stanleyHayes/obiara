//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	competitionmongo "github.com/stanleyHayes/obiara/services/api/internal/games/competition/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func command(id string, r uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ExpectedRevision: r, At: at}
}
func TestResultCASReviewAppealAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
	defer client.Disconnect(context.Background())
	db := client.Database("competition_test")
	r := competitionmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	x, e := domain.Create("competition-1", key(9), []string{key(4), key(1), key(3), key(2)}, command("create", 0, now))
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Create(ctx, x); e != nil {
		t.Fatal(e)
	}
	m := x.Matches()[0]
	a, _ := x.RecordResult(m.ID, m.FirstKey, key(7), command("result-a", 1, now))
	b, _ := x.RecordResult(m.ID, m.SecondKey, key(8), command("result-b", 1, now))
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, a, 1, "result-a") }()
	go func() { ch <- r.Append(ctx, b, 1, "result-b") }()
	ok, conflict := 0, 0
	for range 2 {
		v := <-ch
		if v == nil {
			ok++
		} else if errors.Is(v, application.ErrConflict) {
			conflict++
		} else {
			t.Fatal(v)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("%d %d", ok, conflict)
	}
	x, e = r.Find(ctx, "competition-1")
	if e != nil {
		t.Fatal(e)
	}
	before := x.Ladder()
	x, e = x.OpenReview("review-1", m.ID, key(6), key(1), now, command("review", 2, now))
	if e != nil || r.Append(ctx, x, 2, "review") != nil {
		t.Fatal(e)
	}
	after := x.Ladder()
	if fmt.Sprint(before) != fmt.Sprint(after) || x.Reviews()[0].Decision != domain.DecisionNone {
		t.Fatal("evidence became accusation or ladder change")
	}
	x, e = x.ResolveReview("review-1", domain.DecisionRulesAction, now, command("resolve", 3, now))
	if e != nil || r.Append(ctx, x, 3, "resolve") != nil {
		t.Fatal(e)
	}
	x, e = x.Appeal("review-1", key(2), command("appeal", 4, now))
	if e != nil || r.Append(ctx, x, 4, "appeal") != nil {
		t.Fatal(e)
	}
	x, e = x.ResolveAppeal("review-1", domain.DecisionNoAction, now, command("appeal-resolve", 5, now))
	if e != nil || r.Append(ctx, x, 5, "appeal-resolve") != nil {
		t.Fatal(e)
	}
	var raw bson.M
	if e = db.Collection("game_competitions").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"cohort-private", "entrant-private", "matching", "discovery", "popular", "payment", "payToRank", "trustVisibility", "accusation", "freeText", "public", "listing"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
