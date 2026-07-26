package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func TestServiceKeysAndCASPersistsOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keys := NewMockKeyer(ctrl)
	now := time.Unix(200, 0)
	c, _ := domain.New(k(9), []string{k(1), k(2)}, time.Unix(100, 0))
	keys.EXPECT().Key("closure_room", "raw-room").Return(k(9), nil)
	keys.EXPECT().Key("closure_actor", "raw-member").Return(k(1), nil)
	repo.EXPECT().Find(gomock.Any(), k(9)).Return(c, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "close-command-01").DoAndReturn(func(_ context.Context, s domain.Closure, _ uint64, _ string) error {
		if s.Status() != domain.StatusClosed {
			t.Fatal("not closed")
		}
		return nil
	})
	got, err := NewService(repo, keys, func() time.Time { return now }, time.Hour).Close(context.Background(), "raw-room", "raw-member", "close-command-01", 0)
	if err != nil || got.Status() != domain.StatusClosed {
		t.Fatalf("%s %v", got.Status(), err)
	}
}
func TestServiceInactivityHasNoActorLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keys := NewMockKeyer(ctrl)
	c, _ := domain.New(k(9), []string{k(1), k(2)}, time.Unix(100, 0))
	keys.EXPECT().Key("closure_room", "raw-room").Return(k(9), nil)
	repo.EXPECT().Find(gomock.Any(), k(9)).Return(c, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "timeout-command-01").Return(nil)
	_, err := NewService(repo, keys, func() time.Time { return time.Unix(200, 0) }, time.Minute).CloseInactive(context.Background(), "raw-room", "timeout-command-01", 0)
	if err != nil {
		t.Fatal(err)
	}
}
