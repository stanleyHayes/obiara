package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
	"go.uber.org/mock/gomock"
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
