//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	queuemongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRetryAndMultiDeviceOrdering(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := apimongo.Connect(ctx, uri+separator+"directConnection=true")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("courtship_queue_test")
	repository := queuemongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	room := strings.Repeat("1", 64)
	state, err := repository.State(ctx, room)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	makeEvent := func(command, device, payload string) (domain.State, domain.Event, error) {
		return state.Accept(domain.Command{ID: command, DeviceKey: device, ActorKey: strings.Repeat("3", 64), PayloadKey: payload, Fingerprint: strings.Repeat(payload[:1], 64), BaseSequence: 0, At: now})
	}
	nextA, eventA, _ := makeEvent("device-a-command", strings.Repeat("a", 64), strings.Repeat("4", 64))
	nextB, eventB, _ := makeEvent("device-b-command", strings.Repeat("b", 64), strings.Repeat("5", 64))
	type result struct {
		event  domain.Event
		replay bool
		err    error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for index, candidate := range []struct {
		next  domain.State
		event domain.Event
	}{{nextA, eventA}, {nextB, eventB}} {
		_ = index
		wg.Add(1)
		go func(c struct {
			next  domain.State
			event domain.Event
		}) {
			defer wg.Done()
			event, replay, e := repository.Append(ctx, c.next, c.event, 0)
			results <- result{event, replay, e}
		}(candidate)
	}
	wg.Wait()
	close(results)
	success, stale := 0, 0
	var winner domain.Event
	for result := range results {
		if result.err == nil {
			success++
			winner = result.event
		} else if errors.Is(result.err, domain.ErrStaleDevice) {
			stale++
		} else {
			t.Fatal(result.err)
		}
	}
	if success != 1 || stale != 1 {
		t.Fatalf("success=%d stale=%d", success, stale)
	}
	stored, replay, err := repository.Append(ctx, nextA, winner, 0)
	if err != nil || !replay || stored.Sequence != 1 {
		t.Fatalf("retry replay=%v event=%#v err=%v", replay, stored, err)
	}
	state, err = repository.State(ctx, room)
	if err != nil {
		t.Fatal(err)
	}
	next, event, err := state.Accept(domain.Command{ID: "next-command", DeviceKey: strings.Repeat("a", 64), ActorKey: strings.Repeat("3", 64), PayloadKey: strings.Repeat("6", 64), Fingerprint: strings.Repeat("7", 64), BaseSequence: 1, At: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = repository.Append(ctx, next, event, state.Revision); err != nil {
		t.Fatal(err)
	}
	events, err := repository.Events(ctx, room, 0, 10)
	if err != nil || len(events) != 2 || events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("events=%#v err=%v", events, err)
	}
	var document bson.M
	if err = database.Collection("courtship_queue_events").FindOne(ctx, bson.M{"roomKey": room}).Decode(&document); err != nil {
		t.Fatal(err)
	}
	raw, _ := bson.MarshalExtJSON(document, false, false)
	for _, forbidden := range []string{"raw-room", "raw-device", "raw-actor", "raw-payload", "email", "phone"} {
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("privacy leak %q in %s", forbidden, raw)
		}
	}
}
