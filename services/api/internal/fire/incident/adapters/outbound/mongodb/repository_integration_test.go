//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	incidentmongo "github.com/stanleyHayes/obiara/services/api/internal/fire/incident/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestSingleCaseRoutingConcurrencyAndPrivacy(t *testing.T) {
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
	db := client.Database("fire_incident_test")
	r := incidentmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	a, _ := domain.Create("case-a", key(1), key(2), domain.CategoryHarassment, key(3), now, domain.Command{ID: "trigger", At: now})
	b, _ := domain.Create("case-b", key(1), key(2), domain.CategoryHarassment, key(3), now, domain.Command{ID: "trigger", At: now})
	ch := make(chan error, 2)
	go func() { ch <- r.Create(ctx, a) }()
	go func() { ch <- r.Create(ctx, b) }()
	saved, replay := 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			saved++
		} else if errors.Is(x, application.ErrApplied) {
			replay++
		} else {
			t.Fatal(x)
		}
	}
	if saved != 1 || replay != 1 {
		t.Fatalf("%d %d", saved, replay)
	}
	incident, e := r.FindByCommand(ctx, "trigger")
	if e != nil {
		t.Fatal(e)
	}
	routeID := "trigger:route"
	routed, e := incident.Route(now.Add(time.Second), domain.Command{ID: routeID, ExpectedRevision: 1, At: now.Add(time.Second)})
	if e != nil || r.Append(ctx, routed, 1, routeID) != nil {
		t.Fatal(e)
	}
	if e = r.Append(ctx, routed, 1, routeID); !errors.Is(e, application.ErrApplied) {
		t.Fatalf("route replay=%v", e)
	}
	count, e := db.Collection("fire_incidents").CountDocuments(ctx, bson.M{})
	if e != nil || count != 1 {
		t.Fatalf("cases=%d %v", count, e)
	}
	var raw bson.M
	if e = db.Collection("fire_incidents").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"fire-private", "actor-private", "raw-evidence", "accusation", "freeText", "content", "reporter", "host", "public", "listing", "reverse"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
	stored, e := r.FindByCase(ctx, incident.CaseID())
	if e != nil || !stored.Project().Routed {
		t.Fatalf("%+v %v", stored, e)
	}
}
