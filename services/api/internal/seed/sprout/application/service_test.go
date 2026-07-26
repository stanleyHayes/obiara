package application

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	"go.uber.org/mock/gomock"
)

func TestUnilateralSproutReturnsNoDoorway(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	keyer.EXPECT().Key("participant", "alice").Return("alice-key", nil)
	keyer.EXPECT().Key("participant", "bob").Return("bob-key", nil)
	keyer.EXPECT().Key("seed", "seed-raw").Return("seed-key", nil)
	ids.EXPECT().NewID().Return("intent-1")
	repository.EXPECT().RecordIntent(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, intent domain.Intent) (*domain.Doorway, bool, error) {
		if intent.ActorKey != "alice-key" || intent.TargetKey != "bob-key" {
			t.Fatalf("intent=%#v", intent)
		}
		return nil, false, nil
	})
	service := New(repository, keyer, ids, time.Now)
	result, err := service.Sprout(context.Background(), SproutCommand{"command", "alice", "bob", "seed-raw"})
	if err != nil || result.Doorway != nil {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestExchangeUsesOpaqueReferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Now()
	current, _ := domain.Open("doorway", "alice-key", "bob-key", now)
	repository.EXPECT().FindDoorway(gomock.Any(), "doorway").Return(current, nil)
	keyer.EXPECT().Key("participant", "alice").Return("alice-key", nil)
	keyer.EXPECT().Key("message", "raw-message").Return("message-key", nil)
	repository.EXPECT().AppendExchange(gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(func(_ context.Context, next domain.Doorway, _ uint64) (domain.Doorway, bool, error) {
		if next.Exchanges()[0].MessageKey != "message-key" {
			t.Fatal("raw ref persisted")
		}
		return next, false, nil
	})
	service := New(repository, keyer, ids, func() time.Time { return now })
	result, err := service.Exchange(context.Background(), ExchangeCommand{"exchange", "doorway", "alice", "raw-message"})
	if err != nil || len(result.Doorway.Exchanges()) != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}
