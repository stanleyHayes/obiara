package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	"go.uber.org/mock/gomock"
)

func TestServiceRelightKeysIdentityAndPersists(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockRepository(controller)
	keys := NewMockKeyer(controller)
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	memberOne, memberTwo := fmt.Sprintf("%064x", 1), fmt.Sprintf("%064x", 2)
	pace, _ := domain.New("pace_room_key", []string{memberOne, memberTwo}, "command_start", memberOne, now.Add(-4*24*time.Hour))
	pace, _ = pace.Advance(domain.Command{ID: "command_rest", ExpectedRevision: 1, At: pace.ResponseAt()})

	keys.EXPECT().Key("pace_room", "room-raw").Return("pace_room_key", nil)
	keys.EXPECT().Key("pace_member", "member-raw").Return(memberOne, nil)
	repository.EXPECT().Find(gomock.Any(), "pace_room_key").Return(pace, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(2), "command_relight").DoAndReturn(
		func(_ context.Context, changed domain.Pace, _ uint64, _ string) error {
			if changed.Status() != domain.StatusResting || len(changed.RelightGrants()) != 1 {
				t.Fatal("unilateral relight changed the room state")
			}
			return nil
		},
	)
	service := NewService(repository, keys, func() time.Time { return now })
	if _, err := service.Relight(context.Background(), "room-raw", "member-raw", "command_relight", 2); err != nil {
		t.Fatal(err)
	}
}
