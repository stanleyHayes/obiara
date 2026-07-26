package application

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	"go.uber.org/mock/gomock"
)

func TestServiceSpendsThroughAuthoritativeRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockSubjectKeyer(ctrl)
	now := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)
	policy, _ := domain.NewWeekPolicy("Africa/Accra")
	current, _ := domain.Issue("opaque-key", 10, policy.Start(now), now, "issue", "issue-fp")
	keyer.EXPECT().Key("raw-member").Return("opaque-key", nil)
	repository.EXPECT().Find(gomock.Any(), "opaque-key").Return(current, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), int64(1)).DoAndReturn(
		func(_ context.Context, next domain.Ledger, expected int64) error {
			if next.Balance() != 7 || expected != 1 {
				t.Fatalf("unexpected save: balance=%d expected=%d", next.Balance(), expected)
			}
			return nil
		})
	service, err := NewService(repository, keyer, policy, func() time.Time { return now }, 10)
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.Spend(context.Background(), "raw-member", "spend-command", 3)
	if err != nil || got.Balance() != 7 {
		t.Fatalf("got balance=%d err=%v", got.Balance(), err)
	}
}
