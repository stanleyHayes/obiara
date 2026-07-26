package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
	"go.uber.org/mock/gomock"
)

func safetyKey(n int) string { return fmt.Sprintf("%064x", n) }

func TestCheckUsesPseudonymousKeyAndPersistsAllowedDecision(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockRepository(controller)
	keys := NewMockKeyer(controller)
	now := time.Date(2026, 7, 26, 10, 1, 0, 0, time.UTC)
	bucket, _ := domain.New(safetyKey(1), now)
	keys.EXPECT().Key("seed_safety_actor", "raw-member").Return(safetyKey(1), nil)
	repository.EXPECT().Find(gomock.Any(), safetyKey(1)).Return(bucket, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(
		func(_ context.Context, saved domain.Bucket, _ uint64) error {
			if saved.ActorKey == "raw-member" || saved.Sows != 1 {
				t.Fatalf("saved=%#v", saved)
			}
			return nil
		},
	)
	decision, err := NewService(repository, keys, func() time.Time { return now }).Check(
		context.Background(), "raw-member", domain.OperationSow,
	)
	if err != nil || !decision.Allowed {
		t.Fatalf("decision=%#v err=%v", decision, err)
	}
}
