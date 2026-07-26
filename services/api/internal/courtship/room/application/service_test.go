package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type ids struct{}

func (ids) NewID() string { return "room-1" }
func key(n int) string    { return fmt.Sprintf("%064x", n) }
func TestOpenPersistsOnlyOpaqueTwoMemberRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, m, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockMembership(ctrl), NewMockKeyer(ctrl)
	s := NewService(r, a, m, k, ids{}, func() time.Time { return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC) })
	a.EXPECT().Require(gomock.Any(), "member-a", "courtship.room.open", "")
	m.EXPECT().RevalidatePair(gomock.Any(), "member-a", "member-b")
	k.EXPECT().Key("courtship-room:member", "member-a").Return(key(1), nil)
	k.EXPECT().Key("courtship-room:member", "member-b").Return(key(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, x domain.Room) error {
		if len(x.Members()) != 2 || x.Members()[0] == "member-a" {
			t.Fatal("non-private members")
		}
		return nil
	})
	x, e := s.Open(context.Background(), Command{ID: "open", ActorID: "member-a", ReasonCode: "member_action"}, Proposal{FirstMemberID: "member-a", SecondMemberID: "member-b"})
	if e != nil || x.Room.Revision() != 1 {
		t.Fatalf("%+v %v", x, e)
	}
}
