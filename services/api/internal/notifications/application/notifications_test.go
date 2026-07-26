package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/notifications/domain"
)

var notificationsNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func newService(t *testing.T) (NotificationService, *MockPreferencesRepository, *MockDeliveryCounter) {
	t.Helper()
	ctrl := gomock.NewController(t)
	preferences := NewMockPreferencesRepository(ctrl)
	counter := NewMockDeliveryCounter(ctrl)
	return NewNotificationService(preferences, counter, func() time.Time { return notificationsNow }), preferences, counter
}

func defaultPrefs(t *testing.T) domain.Preferences {
	t.Helper()
	preferences, err := domain.NewPreferences("m-1", notificationsNow)
	if err != nil {
		t.Fatal(err)
	}
	return preferences
}

func TestGetCreatesDefaultsOnFirstRead(t *testing.T) {
	service, preferences, _ := newService(t)
	preferences.EXPECT().FindByMember(gomock.Any(), "m-1").Return(domain.Preferences{}, ErrPreferencesNotFound)
	preferences.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, stored domain.Preferences) error {
			if stored.Timezone() != "Africa/Accra" || len(stored.Muted()) != 0 {
				t.Fatalf("defaults = %#v", stored)
			}
			return nil
		})

	if _, err := service.Get(context.Background(), "m-1"); err != nil {
		t.Fatal(err)
	}
}

func TestDecideClaimsSlotWhenAllowed(t *testing.T) {
	service, preferences, counter := newService(t)
	preferences.EXPECT().FindByMember(gomock.Any(), "m-1").Return(defaultPrefs(t), nil)
	counter.EXPECT().ClaimSlot(gomock.Any(), "m-1", "2026-07-26", domain.DailyCap).Return(true, nil)

	decision, err := service.Decide(context.Background(), "m-1", domain.CategoryRitual)
	if err != nil || !decision.Allowed {
		t.Fatalf("decision = %v, %v", decision, err)
	}
}

func TestDecideDailyCapExceeded(t *testing.T) {
	service, preferences, counter := newService(t)
	preferences.EXPECT().FindByMember(gomock.Any(), "m-1").Return(defaultPrefs(t), nil)
	counter.EXPECT().ClaimSlot(gomock.Any(), "m-1", "2026-07-26", domain.DailyCap).Return(false, nil)

	decision, err := service.Decide(context.Background(), "m-1", domain.CategoryRitual)
	if err != nil || decision.Allowed || decision.Reason != "daily_cap" {
		t.Fatalf("decision = %v, want suppressed daily_cap", decision)
	}
}

func TestDecideQuietHoursSkipsCounter(t *testing.T) {
	service, preferences, _ := newService(t)
	service.now = func() time.Time { return notificationsNow.Add(10 * time.Hour) } // 22:00 local
	preferences.EXPECT().FindByMember(gomock.Any(), "m-1").Return(defaultPrefs(t), nil)
	// No ClaimSlot expectation: quiet-hours suppression never consumes a slot.

	decision, err := service.Decide(context.Background(), "m-1", domain.CategoryPods)
	if err != nil || decision.Allowed || decision.Reason != "quiet_hours" {
		t.Fatalf("decision = %v, want suppressed quiet_hours", decision)
	}
}

func TestDecideSafetyBypassesCounter(t *testing.T) {
	service, preferences, _ := newService(t)
	service.now = func() time.Time { return notificationsNow.Add(10 * time.Hour) } // quiet hours
	preferences.EXPECT().FindByMember(gomock.Any(), "m-1").Return(defaultPrefs(t), nil)
	// No ClaimSlot expectation: safety never touches the cap counter.

	decision, err := service.Decide(context.Background(), "m-1", domain.CategorySafety)
	if err != nil || !decision.Allowed || decision.Reason != "safety_override" {
		t.Fatalf("decision = %v, want safety_override", decision)
	}
}
