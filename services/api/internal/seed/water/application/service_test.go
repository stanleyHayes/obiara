package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func key(n int) string          { return fmt.Sprintf("%064x", n) }
func TestSecondWaterCreatesOnlyOpaqueRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, c, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockPairConsent(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, _ := domain.Start("water-1", []string{key(1), key(2)}, domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})
	s := NewService(r, a, c, k, fixedID("water-x"), fixedID("raw-room-id"), func() time.Time { return now })
	r.EXPECT().Find(gomock.Any(), "water-1").Return(w, nil)
	a.EXPECT().Require(gomock.Any(), "member-b", "seed.water.mutual", "water-1")
	k.EXPECT().Key("seed-water:member", "member-b").Return(key(2), nil)
	c.EXPECT().Revalidate(gomock.Any(), key(1), key(2))
	k.EXPECT().Key("seed-water:room", "raw-room-id").Return(key(9), nil)
	r.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "second").DoAndReturn(func(_ context.Context, w domain.Water, _ uint64, _ string) error {
		if w.RoomKey() != key(9) {
			t.Fatal("room reference was not opaque")
		}
		return nil
	})
	x, e := s.Water(context.Background(), Command{ID: "second", WaterID: "water-1", ActorID: "member-b", ReasonCode: "member_watered", ExpectedRevision: 1})
	if e != nil || x.Water.RoomKey() != key(9) {
		t.Fatalf("%+v %v", x, e)
	}
}
