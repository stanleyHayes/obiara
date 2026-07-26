//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	runsheetmongo "github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestTimerConcurrencyReplayAndPrivacy(t *testing.T) {
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
	db := client.Database("fire_runsheet_test")
	r := runsheetmongo.NewRepository(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	segments := []domain.Segment{{Type: domain.TypeTalk, TitleCode: "opening", PlannedDuration: time.Minute}, {Type: domain.TypeGame, TitleCode: "game", PlannedDuration: time.Minute, CapabilityRef: domain.AllowedGameCapabilities()[0]}, {Type: domain.TypeClose, TitleCode: "close", PlannedDuration: time.Minute}}
	sheet, e := domain.Create("sheet-1", key(1), key(2), 1, segments, domain.Command{ID: "create", At: now})
	if e != nil || r.Create(ctx, sheet) != nil {
		t.Fatal(e)
	}
	started, e := sheet.Start(now, domain.Command{ID: "start", ExpectedRevision: 1, At: now})
	if e != nil || r.Append(ctx, started, 1, "start") != nil {
		t.Fatal(e)
	}
	saved, e := r.Find(ctx, "sheet-1")
	if e != nil {
		t.Fatal(e)
	}
	p := saved.Project(now.Add(time.Hour))
	if p.Status != domain.StatusRunning || p.CurrentIndex != 0 || p.Remaining != 0 || saved.Revision() != 2 {
		t.Fatalf("timer mutated state %+v", p)
	}
	advanced, _ := saved.Advance(now.Add(time.Hour), domain.Command{ID: "advance", ExpectedRevision: 2, At: now.Add(time.Hour)})
	skipped, _ := saved.Skip(now.Add(time.Hour), domain.Command{ID: "skip", ExpectedRevision: 2, At: now.Add(time.Hour)})
	ch := make(chan error, 2)
	go func() { ch <- r.Append(ctx, advanced, 2, "advance") }()
	go func() { ch <- r.Append(ctx, skipped, 2, "skip") }()
	ok, conflict := 0, 0
	for range 2 {
		x := <-ch
		if x == nil {
			ok++
		} else if errors.Is(x, application.ErrConflict) {
			conflict++
		} else {
			t.Fatal(x)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("%d %d", ok, conflict)
	}
	final, e := r.Find(ctx, "sheet-1")
	if e != nil || final.Current() != 1 || final.Revision() != 3 {
		t.Fatalf("%+v %v", final, e)
	}
	winner := final.Commands()[2].ID
	if e = r.Append(ctx, final, 2, winner); !errors.Is(e, application.ErrApplied) {
		t.Fatalf("replay=%v", e)
	}
	var raw bson.M
	if e = db.Collection("fire_runsheets").FindOne(ctx, bson.M{"_id": "sheet-1"}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"fire-private", "host-private", "dynamicCode", "script", "eject", "punish", "penalty", "public", "listing"} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(bad)) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
