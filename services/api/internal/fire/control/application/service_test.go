package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestEjectRevalidatesThenAtomicallyRevokesAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	validator := NewMockRevalidator(ctrl)
	realtime := NewMockRealtimeControl(ctrl)
	now := time.Now()
	current, _ := domain.Open("control", key(9), key(1), []string{key(2)}, domain.Command{ID: "open", ActorKey: key(1), ReasonCode: "fire.opened", Fingerprint: key(8), At: now})
	repo.EXPECT().Find(gomock.Any(), "control").Return(current, nil)
	keyer.EXPECT().Key("participant", "host").Return(key(1), nil)
	keyer.EXPECT().Key("participant", "target").Return(key(2), nil)
	validator.EXPECT().Authorize(gomock.Any(), current.State(), domain.ActionEjected, key(1), key(2)).Return(nil)
	realtime.EXPECT().EjectAndRevoke(gomock.Any(), key(9), key(2), "eject").Return(nil)
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "eject").Return(nil)
	service := New(repo, keyer, ids, validator, realtime, func() time.Time { return now })
	if _, err := service.Eject(context.Background(), Command{"eject", "control", "host", "target", "moderation.test", 1}); err != nil {
		t.Fatal(err)
	}
}
func key(n int) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = '0'
	}
	b[63] = byte('0' + n)
	return string(b)
}
