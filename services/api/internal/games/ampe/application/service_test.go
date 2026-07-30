package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

type memoryRepository struct {
	rounds   map[string]domain.Round
	commands map[string]string
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{rounds: map[string]domain.Round{}, commands: map[string]string{}}
}

func (repository *memoryRepository) Create(_ context.Context, round domain.Round, command string) error {
	if _, exists := repository.commands[command]; exists {
		return ErrApplied
	}
	repository.rounds[round.Specification().ID] = round
	repository.commands[command] = round.Specification().ID
	return nil
}

func (repository *memoryRepository) Find(_ context.Context, id string) (domain.Round, error) {
	round, exists := repository.rounds[id]
	if !exists {
		return domain.Round{}, ErrNotFound
	}
	return round, nil
}

func (repository *memoryRepository) FindByCommand(_ context.Context, command string) (domain.Round, error) {
	return repository.Find(context.Background(), repository.commands[command])
}

func (repository *memoryRepository) Append(_ context.Context, round domain.Round, expected uint64, command string) error {
	current := repository.rounds[round.Specification().ID]
	if current.Sequence() != expected {
		return ErrConflict
	}
	repository.rounds[round.Specification().ID] = round
	repository.commands[command] = round.Specification().ID
	return nil
}

type allowPair struct{}

func (allowPair) Revalidate(_ context.Context, room, first, second string) error {
	if room != "circle-one" || first == second ||
		!((first == "member-a" && second == "member-b") ||
			(first == "member-b" && second == "member-a")) {
		return errors.New("denied")
	}
	return nil
}

type digestKeyer struct{}

func (digestKeyer) Key(namespace, value string) (string, error) {
	sum := sha256.Sum256([]byte(namespace + ":" + value))
	return hex.EncodeToString(sum[:]), nil
}

type fixedID string

func (id fixedID) NewID() string { return string(id) }

func TestPrivateRoundProjectsNoKeysAndRevealsAtomically(t *testing.T) {
	t.Parallel()
	repository := newMemoryRepository()
	service := NewService(repository, allowPair{}, digestKeyer{}, fixedID("round-one"))
	ctx := context.Background()

	created, err := service.Create(ctx, Command{
		ID: "create-one", RoomID: "circle-one", ActorID: "member-a",
		FirstPlayerID: "member-a", SecondPlayerID: "member-b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if encoded, _ := json.Marshal(created); string(encoded) == "" ||
		containsAny(string(encoded), "member-a", "member-b", "circle-one") {
		t.Fatalf("projection leaked a raw identifier: %s", encoded)
	}
	replayed, err := service.Create(ctx, Command{
		ID: "create-one", RoomID: "circle-one", ActorID: "member-a",
		FirstPlayerID: "member-a", SecondPlayerID: "member-b",
	})
	if err != nil || replayed.ID != created.ID {
		t.Fatalf("idempotent create = %#v, %v", replayed, err)
	}

	a := Command{RoundID: created.ID, RoomID: "circle-one", ActorID: "member-a", FirstPlayerID: "member-a", SecondPlayerID: "member-b"}
	b := Command{RoundID: created.ID, RoomID: "circle-one", ActorID: "member-b", FirstPlayerID: "member-b", SecondPlayerID: "member-a"}
	a.ID = "a-ready"
	afterAReady, err := service.Apply(ctx, a, domain.ActionReady, "")
	if err != nil {
		t.Fatal(err)
	}
	b.ID, b.ExpectedSequence = "b-ready", afterAReady.Sequence
	afterBReady, err := service.Apply(ctx, b, domain.ActionReady, "")
	if err != nil {
		t.Fatal(err)
	}
	a.ID, a.ExpectedSequence = "a-lock", afterBReady.Sequence
	afterALock, err := service.Apply(ctx, a, domain.ActionLock, domain.ChoiceTogether)
	if err != nil {
		t.Fatal(err)
	}
	if afterALock.Complete || afterALock.OtherReveal != nil {
		t.Fatal("first lock exposed a reveal")
	}
	b.ID, b.ExpectedSequence = "b-lock", afterALock.Sequence
	afterBLock, err := service.Apply(ctx, b, domain.ActionLock, domain.ChoiceApart)
	if err != nil {
		t.Fatal(err)
	}
	if !afterBLock.Complete || afterBLock.YourReveal == nil || afterBLock.OtherReveal == nil {
		t.Fatal("second lock did not reveal both choices atomically")
	}
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if len(candidate) > 0 && stringContains(value, candidate) {
			return true
		}
	}
	return false
}

func stringContains(value, candidate string) bool {
	for index := 0; index+len(candidate) <= len(value); index++ {
		if value[index:index+len(candidate)] == candidate {
			return true
		}
	}
	return false
}
