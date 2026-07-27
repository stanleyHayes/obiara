//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/application"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/domain"
)

const integrationTimeout = 3 * time.Minute

type caseOpenerStub struct {
	opened []string
}

func (stub *caseOpenerStub) OpenScamCase(_ context.Context, roomID, _ string, _ float64) error {
	stub.opened = append(stub.opened, roomID)
	return nil
}

func TestScamArcEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_scamarc_test")
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	opener := &caseOpenerStub{}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "sig_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewScamArcService(store, nil, opener, time.Now, ids)

	// First signal alone stays below the watch threshold.
	state, card, err := service.Observe(ctx, "room_1", "m-1", domain.SignalAffectionCadence)
	if err != nil || state.Ladder != domain.LadderNone || card != nil {
		t.Fatalf("after 1 signal = %#v card=%v err=%v", state, card, err)
	}

	// Second kind: (1+3)×1.25 = 5 → education with a card.
	state, card, err = service.Observe(ctx, "room_1", "m-1", domain.SignalAskPattern)
	if err != nil || state.Ladder != domain.LadderEducation || card == nil {
		t.Fatalf("after 2 kinds = %#v card=%v err=%v", state, card, err)
	}

	// Third kind: (1+3+2)×1.5 = 9 → case opens.
	state, _, err = service.Observe(ctx, "room_1", "m-1", domain.SignalEmergencyNarrative)
	if err != nil || state.Ladder != domain.LadderCase {
		t.Fatalf("after 3 kinds = %#v err=%v", state, err)
	}
	if len(opener.opened) != 1 {
		t.Fatalf("cases opened = %v", opener.opened)
	}

	// More signals don't reopen.
	if _, _, err := service.Observe(ctx, "room_1", "m-1", domain.SignalOffPlatformPull); err != nil {
		t.Fatal(err)
	}
	if len(opener.opened) != 1 {
		t.Fatalf("case reopened: %v", opener.opened)
	}

	// State persists.
	persisted, err := service.StateForRoom(ctx, "room_1")
	if err != nil || persisted.Ladder != domain.LadderCase {
		t.Fatalf("persisted = %#v, %v", persisted, err)
	}
}
