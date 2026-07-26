package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
	"go.uber.org/mock/gomock"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func TestLifecycleUsesProviderConfirmedRefundFact(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	confirmations := NewMockRefundConfirmationSource(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	service := New(repo, confirmations, ids, clock)
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	pass, err := service.Grant(context.Background(), key(2), "obiara.pass", key(3), 2, now.Add(30*24*time.Hour), 24*time.Hour, "grant-1")
	if err != nil {
		t.Fatal(err)
	}
	clock.EXPECT().Now().Return(now.Add(time.Hour))
	repo.EXPECT().Find(gomock.Any(), key(1)).Return(pass, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1), "cancel-1").Return(nil)
	pass, err = service.Cancel(context.Background(), key(1), "cancel-1")
	if err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().Find(gomock.Any(), key(1)).Return(pass, nil)
	ids.EXPECT().NewID().Return(key(4))
	clock.EXPECT().Now().Return(now.Add(2 * time.Hour))
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(2), "refund-1").Return(nil)
	pass, err = service.RequestRefund(context.Background(), key(1), "refund-1")
	if err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().Find(gomock.Any(), key(1)).Return(pass, nil)
	confirmations.EXPECT().Confirmed(gomock.Any(), key(4)).Return(key(5), now.Add(3*time.Hour), nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(3), "confirm-1").DoAndReturn(func(_ context.Context, p domain.Pass, _ uint64, _ string) error {
		if p.Status(now) != domain.StatusRefunded {
			t.Fatal("not provider confirmed")
		}
		return nil
	})
	if _, err = service.ConfirmRefund(context.Background(), key(1), "confirm-1"); err != nil {
		t.Fatal(err)
	}
}
