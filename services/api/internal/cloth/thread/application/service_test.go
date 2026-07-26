package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func setup(t *testing.T) (*MockRepository, *MockKeyer, *MockRevealEvidence, domain.Thread) {
	t.Helper()
	ctrl := gomock.NewController(t)
	r := NewMockRepository(ctrl)
	k := NewMockKeyer(ctrl)
	e := NewMockRevealEvidence(ctrl)
	v, _ := domain.New(key(9), []string{key(1), key(2)})
	return r, k, e, v
}
func TestIssueRequiresDurableThemeOneEvidence(t *testing.T) {
	r, k, e, v := setup(t)
	k.EXPECT().Key("thread_pair", "pair").Return(key(9), nil)
	k.EXPECT().Key("thread_member", "member").Return(key(1), nil)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	e.EXPECT().ThemeOneRevealed(gomock.Any(), key(9), "ref_revealabcdefghijklmnop").Return(false, nil)
	_, err := NewService(r, k, e, func() time.Time { return time.Unix(100, 0) }).Issue(context.Background(), "pair", "member", "thread-command-01", "ref_revealabcdefghijklmnop", "ref_recipeabcdefghijklmnop", 1, 0)
	if err != domain.ErrDenied {
		t.Fatalf("%v", err)
	}
}
func TestIssuePersistsOnceWithCASAndReplaySkipsSave(t *testing.T) {
	r, k, e, v := setup(t)
	k.EXPECT().Key("thread_pair", "pair").Return(key(9), nil).Times(2)
	k.EXPECT().Key("thread_member", "member").Return(key(1), nil).Times(2)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	e.EXPECT().ThemeOneRevealed(gomock.Any(), key(9), "ref_revealabcdefghijklmnop").Return(true, nil).Times(2)
	r.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "thread-command-01").DoAndReturn(func(_ context.Context, s domain.Thread, _ uint64, _ string) error { v = s; return nil })
	svc := NewService(r, k, e, func() time.Time { return time.Unix(100, 0) })
	if _, err := svc.Issue(context.Background(), "pair", "member", "thread-command-01", "ref_revealabcdefghijklmnop", "ref_recipeabcdefghijklmnop", 1, 0); err != nil {
		t.Fatal(err)
	}
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	if _, err := svc.Issue(context.Background(), "pair", "member", "thread-command-01", "ref_revealabcdefghijklmnop", "ref_recipeabcdefghijklmnop", 1, 0); err != nil {
		t.Fatal(err)
	}
}
func TestViewPairMemberOnly(t *testing.T) {
	r, k, _, v := setup(t)
	v, _ = v.Issue(domain.Command{ID: "thread-command-01", ActorKey: key(1), RevealRef: "ref_revealabcdefghijklmnop", RecipeRef: "ref_recipeabcdefghijklmnop", BandVersion: 1, At: time.Unix(100, 0)})
	k.EXPECT().Key("thread_pair", "pair").Return(key(9), nil)
	k.EXPECT().Key("thread_member", "outsider").Return(key(8), nil)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	if _, err := NewService(r, k, nil, time.Now).View(context.Background(), "pair", "outsider"); err != domain.ErrDenied {
		t.Fatalf("%v", err)
	}
}
