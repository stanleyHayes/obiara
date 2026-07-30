package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestCreateRevalidatesAndStoresOpaqueOwnership(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockKeyer(ctrl)
	s := NewService(r, a, k, fixedID("story-1"), fixedID("passage-1"), time.Now)
	a.EXPECT().RevalidateAuthors(gomock.Any(), "room-private", "a", "b")
	k.EXPECT().Key("anansesem:room", "room-private").Return(ak(3), nil)
	k.EXPECT().Key("anansesem:author", "a").Return(ak(1), nil)
	k.EXPECT().Key("anansesem:author", "b").Return(ak(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, v domain.Story) error {
		if v.RoomKey() != ak(3) {
			t.Fatal("raw room")
		}
		return nil
	})
	_, e := s.Create(context.Background(), Command{ID: "create", RoomID: "room-private", ActorID: "a", FirstAuthorID: "a", SecondAuthorID: "b"}, "spider-path")
	if e != nil {
		t.Fatal(e)
	}
}

func TestViewProjectsOwnershipWithoutAuthorKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	story, err := domain.Create(
		"story-1", ak(3), "spider-path", []string{ak(1), ak(2)},
		domain.Command{ID: "create", At: now},
	)
	if err != nil {
		t.Fatal(err)
	}
	story, err = story.Add(
		"passage-1", ak(1), "The path answered quietly.", now,
		domain.Command{ID: "add", ExpectedRevision: 1, At: now.Add(time.Minute)},
	)
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(r, a, k, fixedID("story-next"), fixedID("passage-next"), func() time.Time { return now })
	a.EXPECT().RevalidateAuthors(gomock.Any(), "room-private", "a", "b")
	r.EXPECT().Find(gomock.Any(), "story-1").Return(story, nil)
	k.EXPECT().Key("anansesem:room", "room-private").Return(ak(3), nil)
	k.EXPECT().Key("anansesem:author", "a").Return(ak(1), nil).Times(2)
	k.EXPECT().Key("anansesem:author", "b").Return(ak(2), nil)

	view, err := s.View(context.Background(), Command{
		StoryID: "story-1", RoomID: "room-private", ActorID: "a",
		FirstAuthorID: "a", SecondAuthorID: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if view.YourTurn || len(view.Passages) != 1 || !view.Passages[0].Yours ||
		view.Passages[0].Content != "The path answered quietly." {
		t.Fatalf("view=%+v", view)
	}
	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{ak(1), ak(2), ak(3)} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("projection leaked key %q: %s", secret, encoded)
		}
	}
}
