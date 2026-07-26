package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/notifications/domain"
)

var dispatchNow = time.Date(2026, time.July, 27, 6, 30, 0, 0, time.UTC) // Monday 06:30 UTC

func TestDispatchCalendarSkipsSuppressed(t *testing.T) {
	ctrl := gomock.NewController(t)
	members := NewMockMemberSource(ctrl)
	preferences := NewMockPreferenceReader(ctrl)
	decider := NewMockDecider(ctrl)
	fires := NewMockFireSource(ctrl)

	members.EXPECT().ListActiveIDs(gomock.Any(), gomock.Any()).Return([]string{"m-1"}, nil)
	preferences.EXPECT().TimezoneFor(gomock.Any(), "m-1").Return("Africa/Accra", nil)
	decider.EXPECT().Decide(gomock.Any(), "m-1", domain.CategoryRitual).
		Return(domain.Decision{Allowed: false, Reason: "quiet_hours"}, nil).AnyTimes()

	dispatcher := NewDispatcher(members, preferences, decider, fires, nil, nil, func() time.Time { return dispatchNow })
	// A nil outbox would panic on dispatch; suppression must prevent any dispatch.
	if err := dispatcher.DispatchCalendar(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestHeraldSkipsFutureFires(t *testing.T) {
	ctrl := gomock.NewController(t)
	fires := NewMockFireSource(ctrl)
	fires.EXPECT().StartingWithin(gomock.Any(), gomock.Any(), gomock.Any()).Return([]FireWindow{
		{FireID: "fire_far", StartsAt: dispatchNow.Add(12 * time.Hour), Attendees: []string{"m-1"}},
	}, nil)

	// The herald is due 45 minutes before start; 12 hours out, nothing
	// dispatches (a nil outbox would panic otherwise).
	dispatcher := NewDispatcher(nil, nil, nil, fires, nil, nil, func() time.Time { return dispatchNow })
	if err := dispatcher.DispatchHeralds(context.Background()); err != nil {
		t.Fatal(err)
	}
}
