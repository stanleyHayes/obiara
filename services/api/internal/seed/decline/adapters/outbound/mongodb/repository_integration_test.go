//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
)

type ids struct{ next atomic.Int64 }

func (source *ids) NewID() string {
	return fmt.Sprintf("decline-%d", source.next.Add(1))
}

func TestRepositoryExclusionReplayAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uri, "?") {
		uri += "&directConnection=true"
	} else {
		uri += "?directConnection=true"
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	database := client.Database("obiara_seed_decline_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	keyer, err := privacy.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := application.New(repository, keyer, &ids{}, func() time.Time { return now })
	command := application.Command{
		ID: "command-1", DeclinerID: "member-private", SeedID: "seed-private", OwnerID: "owner-private",
	}
	result, err := service.Decline(ctx, command)
	if err != nil || result.Replayed {
		t.Fatalf("first result=%#v error=%v", result, err)
	}
	replay, err := service.Decline(ctx, command)
	if err != nil || !replay.Replayed {
		t.Fatalf("replay result=%#v error=%v", replay, err)
	}
	mismatch := command
	mismatch.SeedID = "different-seed"
	if _, err := service.Decline(ctx, mismatch); !errors.Is(err, domain.ErrCommandMismatch) {
		t.Fatalf("mismatch error=%v", err)
	}

	for _, check := range []struct {
		at   time.Time
		want bool
	}{
		{at: now, want: false},
		{at: now.Add(domain.ExclusionPeriod - time.Nanosecond), want: false},
		{at: now.Add(domain.ExclusionPeriod), want: true},
		{at: now.Add(domain.ExclusionPeriod + time.Nanosecond), want: true},
	} {
		eligible, err := service.Eligible(ctx, command.DeclinerID, command.SeedID, check.at)
		if err != nil || eligible != check.want {
			t.Fatalf("eligible at %v=%v, want %v, error=%v", check.at, eligible, check.want, err)
		}
	}

	count, err := database.Collection("seed_declines").CountDocuments(ctx, bson.M{})
	if err != nil || count != 1 {
		t.Fatalf("seed_declines count=%d error=%v", count, err)
	}
	var document bson.M
	if err := database.Collection("seed_declines").FindOne(ctx, bson.M{}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"member-private", "seed-private", "owner-private", "reason", "reject", "public"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("stored decline leaked %q: %s", forbidden, encoded)
		}
	}
	notification := nestedDocument(t, document["notification"])
	for _, required := range []string{"eventKey", "recipientKey", "kind", "occurredAt"} {
		if _, exists := notification[required]; !exists {
			t.Fatalf("notification is missing %q: %#v", required, notification)
		}
	}
	if notification["kind"] != domain.NotificationKind || len(notification) != 4 {
		t.Fatalf("notification is not minimal and neutral: %#v", notification)
	}
}

func nestedDocument(t *testing.T, value any) bson.M {
	t.Helper()
	switch value := value.(type) {
	case bson.M:
		return value
	case bson.D:
		document := make(bson.M, len(value))
		for _, element := range value {
			document[element.Key] = element.Value
		}
		return document
	default:
		t.Fatalf("nested document type=%T", value)
		return nil
	}
}
