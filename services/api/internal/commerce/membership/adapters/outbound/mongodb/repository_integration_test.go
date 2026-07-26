//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adapter "github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

func TestMongoConcurrencyIdempotencyAndPrivacy(t *testing.T) {
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
	db := client.Database("membership_test")
	repo := adapter.New(db)
	if err = repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	pass, _ := domain.New(domain.Grant{ID: key(1), MemberKey: key(2), PassID: "obiara.pass", PassVersion: 1, ReceiptRef: key(3), GrantedAt: now, PaidThrough: now.Add(30 * 24 * time.Hour), GraceUntil: now.Add(31 * 24 * time.Hour), GraceDuration: 24 * time.Hour}, "grant-1")
	ch := make(chan error, 2)
	go func() { ch <- repo.Create(ctx, pass) }()
	go func() { ch <- repo.Create(ctx, pass) }()
	saved, replayed := 0, 0
	for range 2 {
		err = <-ch
		if err == nil {
			saved++
		} else if errors.Is(err, application.ErrApplied) {
			replayed++
		} else {
			t.Fatal(err)
		}
	}
	if saved != 1 || replayed != 1 {
		t.Fatalf("%d %d", saved, replayed)
	}
	loaded, _ := repo.Find(ctx, key(1))
	cancelled, _ := loaded.Cancel("cancel-1", now.Add(time.Hour))
	ch = make(chan error, 2)
	go func() { ch <- repo.Save(ctx, cancelled, loaded.Revision(), "cancel-1") }()
	go func() { ch <- repo.Save(ctx, cancelled, loaded.Revision(), "cancel-1") }()
	saved, replayed = 0, 0
	for range 2 {
		err = <-ch
		if err == nil {
			saved++
		} else if errors.Is(err, application.ErrApplied) {
			replayed++
		} else {
			t.Fatal(err)
		}
	}
	if saved != 1 || replayed != 1 {
		t.Fatalf("save %d %d", saved, replayed)
	}
	var raw bson.M
	if err = db.Collection("membership_passes").FindOne(ctx, bson.M{}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, bad := range []string{"email", "phone", "card", "accountnumber", "romance", "rank", "visibility", "seed", "rawpayment"} {
		if strings.Contains(strings.ToLower(string(encoded)), bad) {
			t.Fatalf("leak %q: %s", bad, encoded)
		}
	}
}
