package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestSetHashesRoomAndActorBeforeGrant(t *testing.T) {
	c := gomock.NewController(t)
	r := NewMockRepository(c)
	k := NewMockKeyer(c)
	x, _ := domain.New(key(1), []string{key(2), key(3)})
	k.EXPECT().Key("honesty_room", "raw-room").Return(key(1), nil)
	k.EXPECT().Key("honesty_actor", "raw-member").Return(key(2), nil)
	r.EXPECT().Find(gomock.Any(), key(1)).Return(x, nil)
	r.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "command-01").Return(nil)
	changed, e := NewService(r, k, time.Now).Set(context.Background(), "raw-room", "raw-member", "command-01", true, 0)
	if e != nil || len(changed.Grants()) != 1 {
		t.Fatalf("changed=%#v err=%v", changed, e)
	}
}
