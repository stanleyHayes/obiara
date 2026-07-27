//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	fairmongo "github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const (
	a = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	c = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	d = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	e = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

func TestConcurrentImmutableProjectionRehydrationAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, err := box.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("fairness_test")
	repo := fairmongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	definition, _ := domain.NewDefinition(domain.DefinitionSpec{ID: "fairness.v1", ReviewID: "review.v1", ReviewerKey: a, Version: 1, MaxParityGapPermille: 50, ReviewedAt: time.Unix(1, 0)})
	snapshot := domain.Snapshot{ID: a, QuarterKey: b, SourceWatermark: c, Version: 1, WindowStartedAt: time.Unix(1, 0), WindowEndedAt: time.Unix(2, 0), Cohorts: []domain.CohortAggregate{{CohortKey: a, Eligible: 100, Exposed: 50}, {CohortKey: b, Eligible: 100, Exposed: 55}}, PreviousRegretEligible: 1000, PreviousRegretReports: 20, CurrentRegretEligible: 1000, CurrentRegretReports: 10, ColorismAuditComplete: true, CompleteMetrics: []domain.Metric{domain.MetricExposureParity, domain.MetricColorismDrift, domain.MetricRegretTrend, domain.MetricTierASafety}}
	first, _ := domain.Evaluate(c, definition, snapshot, time.Unix(3, 0))
	second, _ := domain.Evaluate(d, definition, snapshot, time.Unix(4, 0))
	ch := make(chan error, 2)
	go func() { ch <- repo.Insert(ctx, first) }()
	go func() { ch <- repo.Insert(ctx, second) }()
	ok, applied := 0, 0
	for range 2 {
		switch e := <-ch; {
		case e == nil:
			ok++
		case errors.Is(e, application.ErrApplied):
			applied++
		default:
			t.Fatal(e)
		}
	}
	if ok != 1 || applied != 1 {
		t.Fatalf("ok=%d applied=%d", ok, applied)
	}
	stored, err := repo.Find(ctx, b, 1)
	if err != nil || stored.Fingerprint != first.Fingerprint || stored.Outcome != domain.OutcomePass {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	if len(stored.Cohorts) != 2 || stored.Cohorts[0].Eligible != 100 {
		t.Fatalf("cohorts=%+v", stored.Cohorts)
	}

	rawDoc := bson.M{}
	bytes, err := bson.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if err = bson.Unmarshal(bytes, &rawDoc); err != nil {
		t.Fatal(err)
	}
	rawDoc["_id"] = e
	rawDoc["quarterKey"] = d
	rawDoc["definitionVersion"] = uint64(99)
	rawDoc["fingerprint"] = a
	if _, err = db.Collection("analytics_fairness_reports").InsertOne(ctx, rawDoc); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Find(ctx, d, 99); !errors.Is(err, domain.ErrInvalidSnapshot) {
		t.Fatalf("malformed read=%v", err)
	}

	cur, err := db.Collection("analytics_fairness_reports").Find(ctx, bson.M{"quarterKey": b})
	if err != nil {
		t.Fatal(err)
	}
	var docs []bson.M
	if err = cur.All(ctx, &docs); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, forbidden := range []string{"member@example", "memberid", "ethnicity", "tribe", "skin-tone", "skintone", "conversation", "messagebody", "matchingrank", "automaticcorrection", "rawcontent", "freetext"} {
		if strings.Contains(value, forbidden) {
			t.Fatalf("leaked %q: %s", forbidden, value)
		}
	}
}
