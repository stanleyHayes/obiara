package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCreateProtectsDetailBeforePersistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	protector := NewMockDetailProtector(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Now()
	keyer.EXPECT().Key("member", "sender").Return(key(1), nil)
	keyer.EXPECT().Key("member", "recipient").Return(key(2), nil)
	protector.EXPECT().Protect(gomock.Any(), "raw-sensitive-detail").Return(key(3), nil)
	ids.EXPECT().NewID().Return("proposal")
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, p domain.Proposal) error {
		if p.State().DetailKey != key(3) {
			t.Fatal("raw detail reached persistence")
		}
		return nil
	})
	service := New(repo, keyer, protector, ids, func() time.Time { return now })
	result, err := service.Create(context.Background(), CreateCommand{"command", "sender", "recipient", "raw-sensitive-detail", domain.TypeMeeting, now.Add(time.Hour)})
	if err != nil || result.Proposal.Status() != domain.StatusPending {
		t.Fatal(result, err)
	}
}
func key(n int) string {
	return "000000000000000000000000000000000000000000000000000000000000000" + string(rune('0'+n))
}
