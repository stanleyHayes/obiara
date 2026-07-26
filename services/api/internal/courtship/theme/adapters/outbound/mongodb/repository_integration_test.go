//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/theme/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/theme/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/theme/domain"
)

func TestRepositoryAtomicRevealReplayCASAndPrivacy(t *testing.T) {
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
	database := client.Database("obiara_courtship_theme_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	theme, err := domain.Open("theme-1", []string{key(1), key(2)}, command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, theme); err != nil {
		t.Fatal(err)
	}
	first, _ := theme.Submit(key(10), command("first", key(1), 1, now.Add(time.Second)))
	if err := repository.Append(ctx, first, 1, "first"); err != nil {
		t.Fatal(err)
	}
	var concealed struct {
		Projection domain.Projection `bson:"projection"`
	}
	if err := database.Collection("courtship_theme_one").FindOne(ctx, bson.M{"_id": "theme-1"}).Decode(&concealed); err != nil {
		t.Fatal(err)
	}
	if concealed.Projection.Revealed || len(concealed.Projection.Submissions) != 0 {
		t.Fatalf("stored projection leaked before both submissions: %#v", concealed.Projection)
	}

	secondA, _ := first.Submit(key(11), command("second-a", key(2), 2, now.Add(2*time.Second)))
	secondB, _ := first.Submit(key(12), command("second-b", key(2), 2, now.Add(2*time.Second)))
	results := make(chan error, 2)
	go func() { results <- repository.Append(ctx, secondA, 2, "second-a") }()
	go func() { results <- repository.Append(ctx, secondB, 2, "second-b") }()
	successes, conflicts := 0, 0
	for range 2 {
		switch err := <-results; {
		case err == nil:
			successes++
		case errors.Is(err, application.ErrOptimisticConflict):
			conflicts++
		default:
			t.Fatalf("append error=%v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	stored, err := repository.Find(ctx, "theme-1")
	if err != nil || !stored.Projection().Revealed || len(stored.Projection().Submissions) != 2 {
		t.Fatalf("projection=%#v error=%v", stored.Projection(), err)
	}
	winner := stored.Commands()[2].ID
	if err := repository.Append(ctx, stored, 2, winner); !errors.Is(err, application.ErrCommandApplied) {
		t.Fatalf("replay error=%v", err)
	}

	var raw bson.M
	if err := database.Collection("courtship_theme_one").FindOne(ctx, bson.M{"_id": "theme-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{
		"member-private", "content-private", "public", "rank", "popularity", "likeCount", "audience",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(forbidden)) {
			t.Fatalf("private theme leaked %q: %s", forbidden, encoded)
		}
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
