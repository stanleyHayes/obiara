package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestServiceUsesOpaqueKeysAndPersistsPause(t *testing.T) {
	c := gomock.NewController(t)
	r := NewMockRepository(c)
	k := NewMockKeyer(c)
	now := time.Now().UTC()
	stone, _ := domain.New(key(1), []string{key(2), key(3)})
	k.EXPECT().Key("pause_room", "raw-room").Return(key(1), nil)
	k.EXPECT().Key("pause_actor", "raw-member").Return(key(2), nil)
	r.EXPECT().Find(gomock.Any(), key(1)).Return(stone, nil)
	r.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "command-01").Return(nil)
	changed, err := NewService(r, k, func() time.Time { return now }).Apply(context.Background(), "raw-room", "raw-member", "command-01", domain.ActionPause, 0)
	if err != nil || changed.Status() != domain.StatusPaused {
		t.Fatalf("changed=%#v err=%v", changed, err)
	}
}
