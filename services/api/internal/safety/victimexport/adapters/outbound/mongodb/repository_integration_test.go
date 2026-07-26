//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	exportmongo "github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/application"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const member = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const refKey = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const redact = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
const token = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"

func TestUseRevokeCASIdempotencyAndPrivacy(t *testing.T) {
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
	db := client.Database("victim_export_test")
	repo := exportmongo.NewRepository(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)
	x, e := domain.Request("export:1", member, domain.PurposeVictimSupport, []domain.Reference{{Kind: domain.KindIncidentSummary, RefKey: refKey, RedactionKey: redact}}, domain.Command{ID: "request", At: now})
	if e != nil || repo.Create(ctx, x) != nil {
		t.Fatal(e)
	}
	x, e = x.Authorize(x.References(), token, domain.Command{ID: "authorize", ExpectedRevision: 1, At: now})
	if e != nil || repo.Append(ctx, x, 1, "authorize") != nil {
		t.Fatal(e)
	}
	used, _ := x.Use(token, domain.Command{ID: "use", ExpectedRevision: 2, At: now.Add(time.Hour)})
	revoked, _ := x.Revoke(domain.Command{ID: "revoke", ExpectedRevision: 2, At: now.Add(time.Hour)})
	ch := make(chan error, 2)
	go func() { ch <- repo.Append(ctx, used, 2, "use") }()
	go func() { ch <- repo.Append(ctx, revoked, 2, "revoke") }()
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
	stored, e := repo.Find(ctx, "export:1")
	if e != nil {
		t.Fatal(e)
	}
	if stored.Status() == domain.StatusUsed {
		if _, e = stored.Use(token, domain.Command{ID: "reuse", ExpectedRevision: 3, At: now.Add(2 * time.Hour)}); e == nil {
			t.Fatal("one-time token reused")
		}
	} else {
		if _, e = stored.Use(token, domain.Command{ID: "revoked-use", ExpectedRevision: 3, At: now.Add(2 * time.Hour)}); e == nil {
			t.Fatal("revoked token used")
		}
	}
	winner := "use"
	candidate := used
	if stored.Status() == domain.StatusRevoked {
		winner = "revoke"
		candidate = revoked
	}
	if e = repo.Append(ctx, candidate, 2, winner); !errors.Is(e, application.ErrApplied) {
		t.Fatalf("replay=%v", e)
	}
	var docs []bson.M
	cur, e := db.Collection("victim_export_authorizations").Find(ctx, bson.M{})
	if e != nil {
		t.Fatal(e)
	}
	if e = cur.All(ctx, &docs); e != nil {
		t.Fatal(e)
	}
	raw, _ := bson.MarshalExtJSON(docs, false, false)
	value := strings.ToLower(string(raw))
	for _, bad := range []string{"reporter@example", "accused", "thirdparty", "rawcontent", "rawevidence", "deliveryaddress", "caseaction"} {
		if strings.Contains(value, bad) {
			t.Fatalf("leaked %q: %s", bad, value)
		}
	}
}
