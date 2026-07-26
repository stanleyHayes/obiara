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
	harvestmongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func key(number int) string { return fmt.Sprintf("%064x", number) }
func payload() domain.Payload {
	return domain.Payload{
		RecipeKey: key(3), RecipeVersion: "grammar-v1", RenderSeed: key(4),
		ProductionTokens: []string{"warp_even", "weft_close", "edge_soft", "tone_warm", "mark_sparse", "finish_matte"},
		Format:           "woven_band", DeliveryRef: key(5), PolicyVersion: "policy-v1",
	}
}
func command(id string, actor, revision int, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: key(actor), ExpectedRevision: uint64(revision), At: at}
}

func TestConsentHandoffProviderBoundaryAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("cloth_harvest_test")
	repository := harvestmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	harvest, err := domain.Create("harvest-1", []string{key(1), key(2)}, payload(), command("create", 1, 0, now))
	if err != nil || repository.Create(ctx, harvest) != nil {
		t.Fatal(err)
	}

	first, _ := harvest.Approve(command("approve-a", 1, 1, now.Add(time.Minute)))
	second, _ := harvest.Approve(command("approve-b", 2, 1, now.Add(time.Minute)))
	type result struct {
		id      string
		harvest domain.Harvest
		err     error
	}
	results := make(chan result, 2)
	go func() { results <- result{"approve-a", first, repository.Append(ctx, first, 1, "approve-a")} }()
	go func() { results <- result{"approve-b", second, repository.Append(ctx, second, 1, "approve-b")} }()
	var loser result
	successes, conflicts := 0, 0
	for range 2 {
		result := <-results
		if result.err == nil {
			successes++
		} else if errors.Is(result.err, application.ErrConflict) {
			conflicts++
			loser = result
		} else {
			t.Fatal(result.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("approval CAS successes=%d conflicts=%d", successes, conflicts)
	}
	current, err := repository.Find(ctx, "harvest-1")
	if err != nil {
		t.Fatal(err)
	}
	actor := 1
	if loser.id == "approve-b" {
		actor = 2
	}
	converged, err := current.Approve(command(loser.id, actor, 2, now.Add(time.Minute)))
	if err != nil || repository.Append(ctx, converged, 2, loser.id) != nil || converged.Status() != domain.StatusReady {
		t.Fatalf("approval convergence=%+v error=%v", converged, err)
	}

	a, _ := converged.Handoff("handoff-a", command("handoff-a-command", 1, 3, now.Add(2*time.Minute)))
	b, _ := converged.Handoff("handoff-b", command("handoff-b-command", 1, 3, now.Add(2*time.Minute)))
	handoffs := make(chan error, 2)
	go func() { handoffs <- repository.Append(ctx, a, 3, "handoff-a-command") }()
	go func() { handoffs <- repository.Append(ctx, b, 3, "handoff-b-command") }()
	successes, conflicts = 0, 0
	for range 2 {
		result := <-handoffs
		if result == nil {
			successes++
		} else if errors.Is(result, application.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("handoff CAS successes=%d conflicts=%d", successes, conflicts)
	}
	saved, err := repository.Find(ctx, "harvest-1")
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := saved.ProviderEnvelope(now.Add(3 * time.Minute))
	if err != nil || envelope.HandoffID == "" || len(envelope.ProductionTokens) != 6 {
		t.Fatalf("envelope=%+v error=%v", envelope, err)
	}
	accepted, err := saved.Callback(domain.StatusAccepted, "accepted_spec", domain.Command{ID: "accepted", ExpectedRevision: 4, At: now.Add(4 * time.Minute)})
	if err != nil || repository.Append(ctx, accepted, 4, "accepted") != nil {
		t.Fatal(err)
	}
	completed, err := accepted.Callback(domain.StatusCompleted, "production_complete", domain.Command{ID: "completed", ExpectedRevision: 5, At: now.Add(5 * time.Minute)})
	if err != nil || repository.Append(ctx, completed, 5, "completed") != nil {
		t.Fatal(err)
	}
	completed, err = repository.FindByHandoff(ctx, envelope.HandoffID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = completed.ProviderEnvelope(now.Add(6 * time.Minute)); !errors.Is(err, domain.ErrTransition) {
		t.Fatalf("completed provider access=%v", err)
	}

	var raw bson.M
	if err = database.Collection("cloth_harvests").FindOne(ctx, bson.M{"_id": "harvest-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, err := bson.MarshalExtJSON(raw, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"member-private", "recipe-private", "delivery-private", "street", "address",
		"payment", "card", "reflection", "transcript", "promptResponse", "public", "listing",
		"name", "email", "phone", "relationship", "circle",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(forbidden)) {
			t.Fatalf("stored/provider payload leaked %q: %s", forbidden, encoded)
		}
	}
	count, err := database.Collection("cloth_harvests").CountDocuments(ctx, bson.M{"handoffId": bson.M{"$exists": true, "$ne": ""}})
	if err != nil || count != 1 {
		t.Fatalf("handoffs=%d error=%v", count, err)
	}
}
