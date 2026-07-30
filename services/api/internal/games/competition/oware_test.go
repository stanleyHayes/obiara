package competition

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/adapters/outbound/privacy"
	competitionapp "github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	owareapp "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	owaresession "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
)

type memoryOwareSessions struct{ game owaresession.Session }

func (store *memoryOwareSessions) Create(_ context.Context, game owaresession.Session) error {
	store.game = game
	return nil
}
func (store *memoryOwareSessions) Find(_ context.Context, id string) (owaresession.Session, error) {
	if store.game.ID() != id {
		return owaresession.Session{}, owareapp.ErrNotFound
	}
	return store.game, nil
}
func (store *memoryOwareSessions) FindByCommand(context.Context, string) (owaresession.Session, error) {
	return owaresession.Session{}, owareapp.ErrNotFound
}
func (store *memoryOwareSessions) FindCurrentByRoom(_ context.Context, room string) (owaresession.Session, error) {
	if store.game.RoomRef() != room {
		return owaresession.Session{}, owareapp.ErrNotFound
	}
	return store.game, nil
}
func (store *memoryOwareSessions) Append(_ context.Context, game owaresession.Session, _ uint64, _ string) error {
	store.game = game
	return nil
}

func TestOwareResultVerifierRequiresExactCompletedBoundWinner(t *testing.T) {
	keyer, err := privacy.NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	room, _ := keyer.Key("game-competition:oware-match", "competition-1:round-1:match-0")
	first, _ := keyer.Key("game-competition:entrant", "member-a")
	second, _ := keyer.Key("game-competition:entrant", "member-b")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	var game owaresession.Session
	rng := rand.New(rand.NewSource(29))
	for trial := 0; trial < 200; trial++ {
		game, err = owaresession.Create(
			fmt.Sprintf("oware-result-%d", trial), room, []string{first, second}, 24*time.Hour, now,
			owaresession.Command{ID: fmt.Sprintf("create-%d", trial), At: now},
		)
		if err != nil {
			t.Fatal(err)
		}
		for move := 0; game.Status() == owaresession.StatusActive && move < 500; move++ {
			player := game.Turn()
			legal := game.Board().LegalMoves(player)
			if len(legal) == 0 {
				break
			}
			actor := game.Players()[int(player)]
			next, moveErr := game.Move(actor, legal[rng.Intn(len(legal))], now.Add(time.Duration(move+1)*time.Second), owaresession.Command{
				ID: fmt.Sprintf("move-%d-%d", trial, move), ExpectedRevision: game.Revision(),
				At: now.Add(time.Duration(move+1) * time.Second),
			})
			if moveErr != nil {
				t.Fatal(moveErr)
			}
			game = next
		}
		if game.Status() == owaresession.StatusCompleted && game.Board().Winner() >= 0 && game.Board().Winner() <= 1 {
			break
		}
	}
	if game.Status() != owaresession.StatusCompleted || game.Board().Winner() < 0 || game.Board().Winner() > 1 {
		t.Fatalf("deterministic board did not produce a decisive result: status=%s winner=%d", game.Status(), game.Board().Winner())
	}
	store := &memoryOwareSessions{game: game}
	verifier := owareResultVerifier{sessions: store, keyer: keyer}
	winnerID := "member-a"
	if game.Players()[game.Board().Winner()] == second {
		winnerID = "member-b"
	}
	if err = verifier.Revalidate(context.Background(), game.ID(), "competition-1", "round-1:match-0", winnerID); err != nil {
		t.Fatalf("exact result rejected: %v", err)
	}
	if err = verifier.Revalidate(context.Background(), game.ID(), "competition-1", "round-1:match-1", winnerID); !errors.Is(err, competitionapp.ErrNotAvailable) {
		t.Fatalf("different match accepted: %v", err)
	}
	loserID := "member-b"
	if winnerID == loserID {
		loserID = "member-a"
	}
	if err = verifier.Revalidate(context.Background(), game.ID(), "competition-1", "round-1:match-0", loserID); !errors.Is(err, competitionapp.ErrNotAvailable) {
		t.Fatalf("wrong winner accepted: %v", err)
	}
}

func TestFairPlayVerifierAcceptsOnlyExactExpiredBoundSession(t *testing.T) {
	keyer, err := privacy.NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	room, _ := keyer.Key("game-competition:oware-match", "competition-2:round-1:match-0")
	first, _ := keyer.Key("game-competition:entrant", "member-a")
	second, _ := keyer.Key("game-competition:entrant", "member-b")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	game, err := owaresession.Create(
		"oware-expired", room, []string{first, second}, time.Minute, now,
		owaresession.Command{ID: "create-expired", At: now},
	)
	if err != nil {
		t.Fatal(err)
	}
	game, err = game.Expire(now.Add(time.Minute), owaresession.Command{
		ID: "expire", ExpectedRevision: game.Revision(), At: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	verifier := owareFairPlayVerifier{sessions: &memoryOwareSessions{game: game}, keyer: keyer}
	if err = verifier.Revalidate(context.Background(), "competition-2", "round-1:match-0", game.ID()); err != nil {
		t.Fatalf("exact expired evidence rejected: %v", err)
	}
	if err = verifier.Revalidate(context.Background(), "competition-2", "round-1:match-1", game.ID()); !errors.Is(err, competitionapp.ErrNotAvailable) {
		t.Fatalf("cross-match evidence accepted: %v", err)
	}
}
