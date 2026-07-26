//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/application"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

const integrationTimeout = 3 * time.Minute

func TestSubanLedgerEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_suban_test")
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "sub_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewSubanService(store, time.Now, ids)

	// Three follow-throughs earn keeps_word, recomputed per read.
	for i := 0; i < 3; i++ {
		if err := service.Record(ctx, "m-1", domain.KindMeetingFollowThrough, domain.Provenance{Source: "meeting", Ref: "mtg-1"}); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	marks, err := service.Marks(ctx, "m-1")
	if err != nil || len(marks) != 1 || marks[0] != domain.MarkKeepsWord {
		t.Fatalf("marks = %v, %v", marks, err)
	}

	// A finding suppresses; marks recompute to none without any ledger edit.
	if err := service.Record(ctx, "m-1", domain.KindHarassmentFinding, domain.Provenance{Source: "panel", Ref: "case-1"}); err != nil {
		t.Fatal(err)
	}
	marks, _ = service.Marks(ctx, "m-1")
	if len(marks) != 0 {
		t.Fatalf("marks after finding = %v, want suppressed", marks)
	}

	// The member-visible ledger holds every event in order.
	events, err := service.Events(ctx, "m-1")
	if err != nil || len(events) != 4 {
		t.Fatalf("ledger = %d events, %v", len(events), err)
	}
	if events[3].Kind != domain.KindHarassmentFinding {
		t.Fatalf("last event = %#v", events[3])
	}

	// Anti-gaming: the 11th same-kind event within 30 days is capped.
	for i := 0; i < domain.PeriodCap; i++ {
		if err := service.Record(ctx, "m-2", domain.KindKindClosure, domain.Provenance{Source: "closure", Ref: "c"}); err != nil {
			t.Fatalf("record %d for m-2: %v", i, err)
		}
	}
	if err := service.Record(ctx, "m-2", domain.KindKindClosure, domain.Provenance{Source: "closure", Ref: "c"}); err != application.ErrPeriodCapReached {
		t.Fatalf("cap = %v, want ErrPeriodCapReached", err)
	}
}
