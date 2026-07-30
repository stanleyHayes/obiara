package ampe

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/privacy"
)

type memoryRounds struct {
	round    domain.Round
	commands map[string]domain.Round
}

func (store *memoryRounds) Create(_ context.Context, round domain.Round, command string) error {
	store.round = round
	store.commands[command] = round
	return nil
}
func (store *memoryRounds) Find(_ context.Context, id string) (domain.Round, error) {
	if store.round.Specification().ID != id {
		return domain.Round{}, application.ErrNotFound
	}
	return store.round, nil
}
func (store *memoryRounds) FindByCommand(_ context.Context, command string) (domain.Round, error) {
	round, ok := store.commands[command]
	if !ok {
		return domain.Round{}, application.ErrNotFound
	}
	return round, nil
}
func (store *memoryRounds) Append(_ context.Context, round domain.Round, expected uint64, command string) error {
	if store.round.Sequence() != expected {
		return application.ErrConflict
	}
	store.round = round
	store.commands[command] = round
	return nil
}

type allowPair struct{}

func (allowPair) Revalidate(context.Context, string, string, string) error { return nil }

type fixedRoundID string

func (id fixedRoundID) NewID() string { return string(id) }

type memoryPresence struct{ values map[string]presenceDoc }

func (store *memoryPresence) touch(_ context.Context, roundID, playerKey string, now time.Time) (presenceDoc, error) {
	id := roundID + ":" + playerKey
	value, ok := store.values[id]
	if !ok {
		value = presenceDoc{ID: id, RoundID: roundID, PlayerKey: playerKey, FirstSeen: now}
	}
	value.LastSeen, value.ExpiresAt = now, now.Add(presenceRetention)
	store.values[id] = value
	return value, nil
}
func (store *memoryPresence) find(_ context.Context, roundID, playerKey string) (presenceDoc, error) {
	value, ok := store.values[roundID+":"+playerKey]
	if !ok {
		return presenceDoc{}, mongo.ErrNoDocuments
	}
	return value, nil
}

func TestPresencePausesAfterGraceAndReconnectsWithoutForfeit(t *testing.T) {
	keyer, err := privacy.NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	roundStore := &memoryRounds{commands: map[string]domain.Round{}}
	rounds := application.NewService(roundStore, allowPair{}, keyer, fixedRoundID("ampe-presence"))
	commandA := application.Command{
		ID: "create", RoundID: "ampe-presence", RoomID: "circle-private",
		ActorID: "member-a", FirstPlayerID: "member-a", SecondPlayerID: "member-b",
	}
	if _, err = rounds.Create(context.Background(), commandA); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	presenceStore := &memoryPresence{values: map[string]presenceDoc{}}
	presence := Presence{rounds: rounds, repository: presenceStore, keyer: keyer, now: func() time.Time { return now }}
	first, err := presence.Observe(context.Background(), commandA)
	if err != nil || first.Paused {
		t.Fatalf("initial grace failed: projection=%+v err=%v", first, err)
	}
	now = now.Add(presenceWindow + time.Second)
	paused, err := presence.Observe(context.Background(), commandA)
	if err != nil || !paused.Paused || paused.Other.Connected {
		t.Fatalf("absence did not pause: projection=%+v err=%v", paused, err)
	}
	commandB := application.Command{
		RoundID: "ampe-presence", RoomID: "circle-private",
		ActorID: "member-b", FirstPlayerID: "member-b", SecondPlayerID: "member-a",
	}
	now = now.Add(time.Second)
	reconnected, err := presence.Observe(context.Background(), commandB)
	if err != nil || reconnected.Paused || !reconnected.You.Connected || reconnected.Complete {
		t.Fatalf("safe reconnect failed: projection=%+v err=%v", reconnected, err)
	}
	if errors.Is(err, application.ErrConflict) {
		t.Fatal("presence reconciliation leaked a conflict")
	}
}
