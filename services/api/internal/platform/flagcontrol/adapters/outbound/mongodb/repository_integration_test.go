//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	flagmongo "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/application"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

const (
	actorA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	actorB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	actorC = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestConcurrentCASIdempotencyAuditValidationAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer box.Terminate(context.Background())
	uri, err := box.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("flag_control_test")
	repo := flagmongo.NewRepository(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	p, _ := domain.NewProposal("proposal:a", "command:a", actorA, domain.CapabilityPayments, domain.EnvironmentProduction, domain.MarketGH, domain.ActionKill, domain.ReasonIncident, start, start.Add(domain.MaxLifetime))
	proposedAudit, _ := domain.NewAudit("audit:proposed", p, actorA, domain.AuditProposed, start)
	if err = repo.CreateWithAudit(ctx, p, proposedAudit); err != nil {
		t.Fatal(err)
	}
	replay, _ := domain.NewProposal("proposal:b", "command:a", actorA, domain.CapabilityPayments, domain.EnvironmentProduction, domain.MarketGH, domain.ActionKill, domain.ReasonIncident, start, start.Add(domain.MaxLifetime))
	replayAudit, _ := domain.NewAudit("audit:replay", replay, actorA, domain.AuditProposed, start)
	if err = repo.CreateWithAudit(ctx, replay, replayAudit); !errors.Is(err, application.ErrApplied) {
		t.Fatalf("replay=%v", err)
	}
	left, _ := p.Approve(actorB, start.Add(time.Minute))
	right, _ := p.Approve(actorC, start.Add(time.Minute))
	leftAudit, _ := domain.NewAudit("audit:left", left, actorB, domain.AuditApproved, start.Add(time.Minute))
	rightAudit, _ := domain.NewAudit("audit:right", right, actorC, domain.AuditApproved, start.Add(time.Minute))
	ch := make(chan error, 2)
	go func() { ch <- repo.SaveWithAudit(ctx, left, p.Version(), leftAudit) }()
	go func() { ch <- repo.SaveWithAudit(ctx, right, p.Version(), rightAudit) }()
	ok, conflict := 0, 0
	for range 2 {
		switch e := <-ch; {
		case e == nil:
			ok++
		case errors.Is(e, application.ErrConflict):
			conflict++
		default:
			t.Fatal(e)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("ok=%d conflict=%d", ok, conflict)
	}
	stored, err := repo.FindByCommand(ctx, "command:a")
	if err != nil || stored.Status() != domain.StatusApproved || stored.Version() != 2 {
		t.Fatalf("stored=%+v err=%v", stored.State(), err)
	}
	expired, fallback, err := stored.Expire(start.Add(domain.MaxLifetime))
	if err != nil || fallback.Enabled || !fallback.Killed {
		t.Fatalf("fallback=%+v err=%v", fallback, err)
	}
	expiredAudit, _ := domain.NewAudit("audit:expired", expired, actorA, domain.AuditExpired, start.Add(domain.MaxLifetime))
	if err = repo.SaveWithAudit(ctx, expired, stored.Version(), expiredAudit); err != nil {
		t.Fatal(err)
	}
	if got, findErr := repo.Find(ctx, "proposal:a"); findErr != nil || got.Status() != domain.StatusExpired {
		t.Fatalf("status=%s err=%v", got.Status(), findErr)
	}

	_, err = db.Collection("platform_flag_control_proposals").InsertOne(ctx, bson.M{"_id": "bad", "commandId": "bad", "fingerprint": actorA, "proposerKey": actorA, "capability": "payments", "environment": "production", "market": "GH", "action": "enable", "reason": "incident", "status": "proposed", "version": 1, "createdAt": start, "expiresAt": start.Add(3 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Find(ctx, "bad"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("malformed=%v", err)
	}
	for _, collection := range []string{"platform_flag_control_proposals", "platform_flag_control_audits"} {
		cur, e := db.Collection(collection).Find(ctx, bson.M{})
		if e != nil {
			t.Fatal(e)
		}
		var docs []bson.M
		if e = cur.All(ctx, &docs); e != nil {
			t.Fatal(e)
		}
		raw, _ := bson.MarshalExtJSON(docs, false, false)
		value := strings.ToLower(string(raw))
		for _, bad := range []string{"member@example", "memberid", "cohort", "authorization", "secret", "token", "password", "freetext", "global"} {
			if strings.Contains(value, bad) {
				t.Fatalf("%s leaked %q: %s", collection, bad, value)
			}
		}
	}
}
