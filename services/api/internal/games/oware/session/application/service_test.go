package application

import (
	"context"
	"fmt"
	session "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestCreateUsesOpaqueRoomAndPlayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, rooms, a, k := NewMockRepository(ctrl), NewMockRoomEmbedding(ctrl), NewMockAuthorizer(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	s := NewService(r, rooms, a, k, fixedID("session-1"), func() time.Time { return now })
	a.EXPECT().RequireParticipant(gomock.Any(), "room-private", "a")
	rooms.EXPECT().Revalidate(gomock.Any(), "room-private", "a", "b")
	k.EXPECT().Key("oware-session:room", "room-private").Return(ak(3), nil)
	k.EXPECT().Key("oware-session:player", "a").Return(ak(1), nil)
	k.EXPECT().Key("oware-session:player", "b").Return(ak(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, v session.Session) error {
		if v.RoomRef() != ak(3) {
			t.Fatal("raw room")
		}
		return nil
	})
	p, e := s.Create(context.Background(), Command{ID: "create", RoomID: "room-private", ActorID: "a"}, "b", time.Hour)
	if e != nil || p.Status != session.StatusActive {
		t.Fatal(e)
	}
}
