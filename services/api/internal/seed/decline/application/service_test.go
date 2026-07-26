package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
)

type fixedIDs struct{ value string }

func (source fixedIDs) NewID() string { return source.value }

func TestDeclinePersistsOnlyOpaqueKeysAndNeutralNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := New(store, keyer, fixedIDs{"decline-1"}, func() time.Time { return now })

	keyer.EXPECT().Key("decliner", "member-private").Return(key("a"), nil)
	keyer.EXPECT().Key("seed", "seed-private").Return(key("b"), nil)
	keyer.EXPECT().Key("notification-recipient", "owner-private").Return(key("c"), nil)
	keyer.EXPECT().Key("notification-event", "command-1").Return(key("d"), nil)
	store.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, decline domain.Decline, notification domain.Notification) (domain.Decline, bool, error) {
			if decline.DeclinerKey() != key("a") || decline.SeedKey() != key("b") ||
				decline.RecipientKey() != key("c") || notification.Kind() != domain.NotificationKind {
				t.Fatalf("decline=%#v notification=%#v", decline, notification)
			}
			return decline, false, nil
		},
	)
	if _, err := service.Decline(context.Background(), Command{
		ID: "command-1", DeclinerID: "member-private", SeedID: "seed-private", OwnerID: "owner-private",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestEligibleReturnsOnlyBooleanFromServerExclusion(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	keyer := NewMockKeyer(ctrl)
	service := New(store, keyer, fixedIDs{"unused"}, time.Now)
	at := time.Now().UTC()
	keyer.EXPECT().Key("decliner", "member").Return(key("a"), nil)
	keyer.EXPECT().Key("seed", "seed").Return(key("b"), nil)
	store.EXPECT().IsExcluded(gomock.Any(), key("a"), key("b"), at).Return(true, nil)
	eligible, err := service.Eligible(context.Background(), "member", "seed", at)
	if err != nil || eligible {
		t.Fatalf("eligible=%v error=%v", eligible, err)
	}
}

func key(value string) string { return strings.Repeat(value, 64) }
