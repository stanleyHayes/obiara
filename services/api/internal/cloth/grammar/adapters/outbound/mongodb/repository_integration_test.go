//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	grammarmongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

func TestDeterministicReplayAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, _ := container.ConnectionString(ctx)
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("cloth_grammar_test")
	repository := grammarmongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	input := domain.Input{Version: domain.VersionV1, PairKeys: [2]string{strings.Repeat("1", 64), strings.Repeat("2", 64)}, ThemeKeys: []string{strings.Repeat("4", 64), strings.Repeat("3", 64)}, ProvenanceKeys: []string{strings.Repeat("6", 64), strings.Repeat("5", 64)}}
	first, _ := domain.Compile(input, "command-1")
	if _, replay, err := repository.Store(ctx, first, 0); err != nil || replay {
		t.Fatal(replay, err)
	}
	input.PairKeys[0], input.PairKeys[1] = input.PairKeys[1], input.PairKeys[0]
	input.ThemeKeys[0], input.ThemeKeys[1] = input.ThemeKeys[1], input.ThemeKeys[0]
	second, _ := domain.Compile(input, "command-2")
	stored, replay, err := repository.Store(ctx, second, 0)
	if err != nil || !replay || stored.RenderSeed() != first.RenderSeed() {
		t.Fatalf("replay=%v seed=%s err=%v", replay, stored.RenderSeed(), err)
	}
	changed := input
	changed.ThemeKeys = []string{strings.Repeat("7", 64)}
	mismatch, _ := domain.Compile(changed, "command-1")
	if _, _, err = repository.Store(ctx, mismatch, 0); !errors.Is(err, domain.ErrCommandMismatch) {
		t.Fatalf("mismatch=%v", err)
	}
	var document bson.M
	if err = database.Collection("cloth_grammar_recipes").FindOne(ctx, bson.M{"_id": first.ID()}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"alice", "bob", "private text", "{{", "<script", "function(", "phone", "location"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("unsafe value %q in %s", forbidden, raw)
		}
	}
}
