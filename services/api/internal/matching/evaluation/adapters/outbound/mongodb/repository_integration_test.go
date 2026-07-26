//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	evaluationmongo "github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func TestCASApprovalPersistenceAndNoRawEvaluationData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("matching_evaluation_test")
	repo := evaluationmongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)
	e, err := domain.Create("evaluation:1", "candidate.compatibility", 1, domain.Command{ID: "create", At: now})
	if err != nil || repo.Create(ctx, e) != nil {
		t.Fatal(err)
	}
	s := domain.Snapshot{ID: "snapshot:consented", Version: 2, ConsentVersion: 8, EvaluatedAt: now}
	m := domain.Metrics{Cohort: 600, Evidence: 5000, Quality: .78, ErrorRate: .2, MaxDisparity: .04, Slices: []domain.SliceMetric{{PolicyKey: "approved.region", Cohort: 300, Quality: .77, ErrorRate: .21}, {PolicyKey: "approved.access", Cohort: 300, Quality: .79, ErrorRate: .19}}}
	e, err = e.Record(s, m, domain.Command{ID: "evaluate", ExpectedRevision: 1, At: now})
	if err != nil || repo.Append(ctx, e, 1, "evaluate") != nil {
		t.Fatal(err)
	}
	e, err = e.AttachCard(domain.ModelCard{Version: 1, Purpose: "matching.compatibility", EvaluationRef: "report:bounded", LimitationsRef: "limits:bounded", Owner: "matching.team"}, domain.Command{ID: "card", ExpectedRevision: 2, At: now})
	if err != nil || repo.Append(ctx, e, 2, "card") != nil {
		t.Fatal(err)
	}
	a, _ := e.Approve("6f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8", now.Add(time.Hour), domain.Command{ID: "approve:a", ExpectedRevision: 3, At: now})
	b, _ := e.Approve("7f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8c7b7f2a6f6a8", now.Add(time.Hour), domain.Command{ID: "approve:b", ExpectedRevision: 3, At: now})
	ch := make(chan error, 2)
	go func() { ch <- repo.Append(ctx, a, 3, "approve:a") }()
	go func() { ch <- repo.Append(ctx, b, 3, "approve:b") }()
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
		t.Fatalf("ok=%d conflict=%d", ok, conflict)
	}
	stored, err := repo.Find(ctx, "evaluation:1")
	if err != nil || !stored.Ready(now) {
		t.Fatalf("ready=%v err=%v", stored.Ready(now), err)
	}
	cur, err := db.Collection("matching_offline_evaluations").Find(ctx, bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	var docs []bson.M
	if err = cur.All(ctx, &docs); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, forbidden := range []string{"human@example.invalid", "rawdata", "rawcontent", "inferredtrait", "sensitiveattribute", "vendor", "productiondecision", "onlinepayload"} {
		if strings.Contains(value, forbidden) {
			t.Fatalf("leaked %q: %s", forbidden, value)
		}
	}
}
