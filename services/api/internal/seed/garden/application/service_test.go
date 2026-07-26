package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	"go.uber.org/mock/gomock"
)

func gardenKey(n int) string { return fmt.Sprintf("%064x", n) }

func TestSummaryExpiresBeforeMemberOnlyProjection(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockRepository(controller)
	keys := NewMockKeyer(controller)
	now := time.Date(2026, 7, 26, 6, 0, 0, 0, time.UTC)
	keys.EXPECT().Key("garden_owner", "member-1").Return(gardenKey(1), nil)
	repository.EXPECT().ExpireDue(gomock.Any(), gardenKey(1), now, 100).Return(int64(1), nil)
	repository.EXPECT().ListOwner(gomock.Any(), gardenKey(1)).Return([]domain.Item{{
		SeedKey: gardenKey(2), OwnerKey: gardenKey(1), State: domain.StateExpired,
		ExpiresAt: now.Add(-time.Minute), UpdatedAt: now, Revision: 2,
	}}, nil)
	summary, err := NewService(repository, keys, func() time.Time { return now }).Summary(context.Background(), "member-1")
	if err != nil || summary.Message != "Nothing needs your attention today." {
		t.Fatalf("summary=%#v err=%v", summary, err)
	}
}

func TestProjectUsesOpaqueOwnerAndSeedKeys(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockRepository(controller)
	keys := NewMockKeyer(controller)
	now := time.Date(2026, 7, 26, 6, 0, 0, 0, time.UTC)
	item, _ := domain.New(gardenKey(2), gardenKey(1), now.Add(time.Hour), now.Add(-time.Minute))
	keys.EXPECT().Key("garden_owner", "raw-member").Return(gardenKey(1), nil)
	keys.EXPECT().Key("garden_seed", "raw-seed").Return(gardenKey(2), nil)
	repository.EXPECT().Find(gomock.Any(), gardenKey(1), gardenKey(2)).Return(item, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(
		func(_ context.Context, saved domain.Item, _ uint64) error {
			if saved.State != domain.StateDelivered || saved.OwnerKey == "raw-member" {
				t.Fatalf("saved %#v", saved)
			}
			return nil
		},
	)
	if _, err := NewService(repository, keys, func() time.Time { return now }).Project(
		context.Background(), "raw-member", "raw-seed", domain.StateDelivered, 1,
	); err != nil {
		t.Fatal(err)
	}
}
