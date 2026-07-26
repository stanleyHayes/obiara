package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestBlockKeysAndPersistsCAS(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keys := NewMockKeyer(ctrl)
	s, _ := domain.New(key(9), []string{key(1), key(2)})
	keys.EXPECT().Key("safety_room", "room").Return(key(9), nil)
	keys.EXPECT().Key("safety_actor", "member").Return(key(1), nil)
	repo.EXPECT().Find(gomock.Any(), key(9)).Return(s, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "block-command-01").Return(nil)
	got, e := NewService(repo, keys, func() time.Time { return time.Unix(100, 0) }).Block(context.Background(), "room", "member", "block-command-01", 0)
	if e != nil || !got.Blocked() {
		t.Fatalf("%v %v", got.Blocked(), e)
	}
}
func TestReportCreatesPrivateReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keys := NewMockKeyer(ctrl)
	s, _ := domain.New(key(9), []string{key(1), key(2)})
	keys.EXPECT().Key("safety_room", "room").Return(key(9), nil)
	keys.EXPECT().Key("safety_actor", "member").Return(key(1), nil)
	repo.EXPECT().Find(gomock.Any(), key(9)).Return(s, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "report-command-01").Return(nil)
	got, e := NewService(repo, keys, func() time.Time { return time.Unix(100, 0) }).Report(context.Background(), "room", "member", "report-command-01", domain.CategoryThreat, "enc_abcdefghijklmnopqrstuvwxyz", 0)
	if e != nil || len(got.Reviews()) != 1 {
		t.Fatalf("%v %v", got.Reviews(), e)
	}
}
