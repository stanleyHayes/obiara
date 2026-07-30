package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/domain"
)

type memoryStore struct {
	prompts  []domain.Prompt
	duels    map[string]StoredDuel
	commands map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{duels: map[string]StoredDuel{}, commands: map[string]string{}}
}
func (store *memoryStore) SaveApproved(_ context.Context, prompt domain.Prompt) error {
	store.prompts = append(store.prompts, prompt)
	return nil
}
func (store *memoryStore) ListApproved(_ context.Context, limit int) ([]domain.Prompt, error) {
	if limit > len(store.prompts) {
		limit = len(store.prompts)
	}
	return append([]domain.Prompt(nil), store.prompts[:limit]...), nil
}
func (store *memoryStore) Create(_ context.Context, duel StoredDuel, command string) error {
	if _, exists := store.commands[command]; exists {
		return ErrApplied
	}
	id := duel.Duel.Specification().ID
	store.duels[id], store.commands[command] = duel, id
	return nil
}
func (store *memoryStore) Find(_ context.Context, id string) (StoredDuel, error) {
	duel, ok := store.duels[id]
	if !ok {
		return StoredDuel{}, ErrNotFound
	}
	return duel, nil
}
func (store *memoryStore) FindByCommand(ctx context.Context, command string) (StoredDuel, error) {
	return store.Find(ctx, store.commands[command])
}
func (store *memoryStore) Append(_ context.Context, duel StoredDuel, expected uint64, command string) error {
	current := store.duels[duel.Duel.Specification().ID]
	if current.Duel.Revision() != expected {
		return ErrConflict
	}
	id := duel.Duel.Specification().ID
	store.duels[id], store.commands[command] = duel, id
	return nil
}

type testAuthority struct{}

func (testAuthority) Revalidate(_ context.Context, room, first, second string) error {
	if room == "circle-one" && first != second &&
		((first == "member-a" && second == "member-b") || (first == "member-b" && second == "member-a")) {
		return nil
	}
	return errors.New("denied")
}
func (testAuthority) RequireReviewer(_ context.Context, reviewer string) error {
	if reviewer == "reviewer-one" {
		return nil
	}
	return errors.New("denied")
}

type testKeyer struct{}

func (testKeyer) Key(namespace, value string) (string, error) {
	sum := sha256.Sum256([]byte(namespace + ":" + value))
	return hex.EncodeToString(sum[:]), nil
}

type testID string

func (id testID) NewID() string { return string(id) }

func TestApprovedCatalogAndPrivateDuelProjection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newMemoryStore()
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	catalog := NewCatalogService(store, testAuthority{}, testKeyer{}, testID("review-one"), func() time.Time { return now })
	prompt, err := catalog.Approve(ctx, "reviewer-one", PromptApproval{
		ID: "reviewed-one", Version: 1, Language: "en",
		Cue:             "A synthetic test cue with no claim of cultural authority.",
		AcceptedAnswers: []string{"private accepted form"},
		Source:          domain.Source{Kind: domain.SourceBook, Citation: "Synthetic test source"},
	})
	if err != nil || strings.Contains(mustJSON(prompt), "private accepted form") {
		t.Fatalf("approval projection leaked accepted form: %#v, %v", prompt, err)
	}
	_, err = catalog.Approve(ctx, "reviewer-one", PromptApproval{
		ID: "reviewed-two", Version: 1, Language: "en",
		Cue:             "A second synthetic test cue.",
		AcceptedAnswers: []string{"second private form"},
		Source:          domain.Source{Kind: domain.SourceBook, Citation: "Synthetic test source"},
	})
	if err != nil {
		t.Fatal(err)
	}

	duels := NewDuelService(store, store, testAuthority{}, testKeyer{}, testID("duel-one"))
	created, err := duels.Create(ctx, Command{
		ID: "create-one", RoomID: "circle-one", ActorID: "member-a",
		FirstPlayerID: "member-a", SecondPlayerID: "member-b",
	})
	if err != nil || !created.YourTurn {
		t.Fatalf("create = %#v, %v", created, err)
	}
	answered, err := duels.Answer(ctx, Command{
		ID: "answer-one", DuelID: created.ID, RoomID: "circle-one",
		ActorID: "member-a", FirstPlayerID: "member-a",
		SecondPlayerID: "member-b", ExpectedRevision: created.Revision,
	}, "private accepted form")
	if err != nil || answered.YourTurn || answered.Turns[0].YourAnswerCorrect == nil || !*answered.Turns[0].YourAnswerCorrect {
		t.Fatalf("answer = %#v, %v", answered, err)
	}
	other, err := duels.View(ctx, Command{
		DuelID: created.ID, RoomID: "circle-one", ActorID: "member-b",
		FirstPlayerID: "member-b", SecondPlayerID: "member-a",
	})
	encoded := mustJSON(other)
	if err != nil || !other.YourTurn || strings.Contains(encoded, "private accepted form") ||
		strings.Contains(encoded, "member-a") || strings.Contains(encoded, "reviewer-one") {
		t.Fatalf("other projection leaked private material: %s, %v", encoded, err)
	}
}

func mustJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
