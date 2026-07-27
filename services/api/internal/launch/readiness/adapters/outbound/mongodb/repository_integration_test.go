//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoImmutableIdempotentPrivateSnapshots(t *testing.T) {
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
	db := client.Database("launch_readiness_test")
	r := adapter.New(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	f := domain.FamilyDensity{Market: "GH", ConsentedFamilies: 100, TargetFamilies: 100, DenseCircles: 5, RequiredDenseCircles: 5, EvidenceVersion: 1, Complete: true, ObservedAt: now}
	h := domain.HostCoverage{Market: "GH", Trained: 10, Certified: 10, Required: 10, TrainingVersion: 1, CertificationVersion: 1, CertifiedUntil: now.Add(time.Hour), Complete: true, ObservedAt: now}
	l := domain.LicenseCoverage{Market: "GH", Jurisdiction: "gh-accra", Licensed: 4, Required: 4, LicenseVersion: 1, LicensedUntil: now.Add(time.Hour), Complete: true, ObservedAt: now}
	s, _ := domain.Project(key(1), key(2), "GH", "gh-accra", "review-1", f, h, l, now)
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, s) }()
	go func() { ch <- r.Create(ctx, s) }()
	ok, replay := 0, 0
	for range 2 {
		e = <-ch
		if e == nil {
			ok++
		} else if errors.Is(e, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(e)
		}
	}
	if ok != 1 || replay != 1 {
		t.Fatalf("%d %d", ok, replay)
	}
	got, e := r.Find(ctx, s.ID())
	if e != nil || !got.Ready() {
		t.Fatal(e)
	}
	var raw bson.M
	if e = db.Collection("launch_readiness_snapshots").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "familyid", "memberid", "membername", "contact", "address", "crm", "outreach"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
	// Repository deliberately exposes no update method: every review is a new immutable snapshot.
}
