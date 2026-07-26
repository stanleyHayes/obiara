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
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/domain"
)

func TestRepositoryOrderConcealConcurrencyReplayAndPrivacy(t *testing.T) {
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
	database := client.Database("obiara_theme_progression_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	progression, err := domain.Open("progression-1", []string{key(1), key(2)}, key(9), command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, progression); err != nil {
		t.Fatal(err)
	}
	if _, err := progression.Submit(domain.ThemeThree, key(20), command("skip", key(1), 1, now.Add(time.Second))); !errors.Is(err, domain.ErrLocked) {
		t.Fatalf("skip error=%v", err)
	}

	first, _ := progression.Submit(domain.ThemeTwo, key(20), command("two-a", key(1), 1, now.Add(time.Second)))
	if err := repository.Append(ctx, first, 1, "two-a"); err != nil {
		t.Fatal(err)
	}
	var concealed struct {
		Projection domain.Projection `bson:"projection"`
	}
	if err := database.Collection("courtship_theme_progressions").FindOne(ctx, bson.M{"_id": "progression-1"}).Decode(&concealed); err != nil {
		t.Fatal(err)
	}
	if concealed.Projection.Themes[0].Revealed || len(concealed.Projection.Themes[0].Submissions) != 0 ||
		concealed.Projection.Themes[1].Unlocked {
		t.Fatalf("stored projection leaked/unlocked early: %#v", concealed.Projection)
	}

	secondA, _ := first.Submit(domain.ThemeTwo, key(21), command("two-b-a", key(2), 2, now.Add(2*time.Second)))
	secondB, _ := first.Submit(domain.ThemeTwo, key(22), command("two-b-b", key(2), 2, now.Add(2*time.Second)))
	results := make(chan error, 2)
	go func() { results <- repository.Append(ctx, secondA, 2, "two-b-a") }()
	go func() { results <- repository.Append(ctx, secondB, 2, "two-b-b") }()
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
	progression, err = repository.Find(ctx, "progression-1")
	if err != nil || !themeState(progression, domain.ThemeTwo).Revealed ||
		!themeState(progression, domain.ThemeThree).Unlocked ||
		themeState(progression, domain.ThemeFour).Unlocked {
		t.Fatalf("projection=%#v error=%v", progression.Projection(), err)
	}
	winner := progression.Commands()[2].ID
	if err := repository.Append(ctx, progression, 2, winner); !errors.Is(err, application.ErrCommandApplied) {
		t.Fatalf("replay error=%v", err)
	}

	revisionTime := 3
	for _, theme := range []domain.ThemeNumber{domain.ThemeThree, domain.ThemeFour} {
		for member := 1; member <= 2; member++ {
			next, submitErr := progression.Submit(theme, key(100+revisionTime), command(
				fmt.Sprintf("%d-%d", theme, member), key(member), progression.Revision(),
				now.Add(time.Duration(revisionTime)*time.Second),
			))
			if submitErr != nil {
				t.Fatal(submitErr)
			}
			if err := repository.Append(ctx, next, progression.Revision(), next.Commands()[next.Revision()-1].ID); err != nil {
				t.Fatal(err)
			}
			progression = next
			if member == 1 && (themeState(progression, theme).Revealed || len(themeState(progression, theme).Submissions) != 0) {
				t.Fatalf("theme %d leaked after one response", theme)
			}
			revisionTime++
		}
	}

	var raw bson.M
	if err := database.Collection("courtship_theme_progressions").FindOne(ctx, bson.M{"_id": "progression-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, _ := bson.MarshalExtJSON(raw, false, false)
	for _, forbidden := range []string{
		"member-private", "content-private", "public", "payment", "purchase", "tier",
		"score", "streak", "rank", "popularity", "audience",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("private progression leaked %q: %s", forbidden, encoded)
		}
	}
}

func themeState(progression domain.Progression, number domain.ThemeNumber) domain.ThemeState {
	for _, state := range progression.Projection().Themes {
		if state.Number == number {
			return state
		}
	}
	return domain.ThemeState{}
}
func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
