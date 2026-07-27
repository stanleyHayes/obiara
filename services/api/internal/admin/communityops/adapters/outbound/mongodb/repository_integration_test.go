//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestMongoConcurrencyIdempotencyPrivacy(t *testing.T) {
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
	db := client.Database("communityops_test")
	repo := adapter.New(db)
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	d := domain.Density{CircleKey: key(1), FireKey: key(2), Participants: 20, Capacity: 50, Version: 1, ObservedAt: now}
	n := domain.Notice{TemplateKey: "fire.cancel", Version: 1, Locale: "en-gh", AudienceCount: 20}
	n.Digest = domain.NoticeDigest(n.TemplateKey, 1, n.Locale, 20)
	p, _ := domain.Propose(key(9), key(8), domain.CancelFire, domain.ReasonScheduleConflict, key(7), d, nil, n, "propose-1", now)
	ch := make(chan error, 2)
	go func() { ch <- repo.Create(ctx, p) }()
	go func() { ch <- repo.Create(ctx, p) }()
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
	next, _ := p.AcknowledgeNotice(key(8), n.Digest, "ack-1", now)
	ch = make(chan error, 2)
	go func() { ch <- repo.Save(ctx, next, 1, "ack-1") }()
	go func() { ch <- repo.Save(ctx, next, 1, "ack-1") }()
	ok, replay = 0, 0
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
	var raw bson.M
	if e = db.Collection("community_operation_proposals").FindOne(ctx, bson.M{}).Decode(&raw); e != nil {
		t.Fatal(e)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "memberid", "membername", "content", "messagebody", "notificationtoken", "address"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("privacy leak %q: %s", bad, encoded)
		}
	}
}
