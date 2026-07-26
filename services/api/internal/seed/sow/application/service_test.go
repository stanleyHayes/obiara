package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
	"go.uber.org/mock/gomock"
)

func TestSendScreensBeforeAtomicAcceptance(t *testing.T) {
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	screening.EXPECT().Screen(gomock.Any(), "hello", []string{"raw-media"}).Return(ScreeningDecision{Approved: true, Reference: "raw-screen"}, nil)
	keyer.EXPECT().Key("allowance-subject", "raw-actor").Return("actor-key", nil)
	keyer.EXPECT().Key("screening", "raw-screen").Return("screen-key", nil)
	keyer.EXPECT().Key("media", "raw-media").Return("media-key", nil)
	ids.EXPECT().NewID().Return("sow-1")
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) {
		if s.ActorKey != "actor-key" || s.Media[0].Key != "media-key" || s.AllowanceUnits != 1 {
			t.Fatalf("unsafe candidate %#v", s)
		}
		return s, false, nil
	})
	service := New(screening, acceptance, keyer, ids, func() time.Time { return now }, 1)
	result, err := service.Send(context.Background(), Command{ID: "command-1", ActorID: "raw-actor", Body: " hello ", MediaRefs: []string{"raw-media"}, Confirmed: true})
	if err != nil || result.Sow.ID != "sow-1" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestSendRejectsWithoutConfirmationOrScreening(t *testing.T) {
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	service := New(screening, NewMockAcceptance(ctrl), NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", Body: "body"}); !errors.Is(err, domain.ErrNotConfirmed) {
		t.Fatalf("got %v", err)
	}
	screening.EXPECT().Screen(gomock.Any(), "body", gomock.Any()).Return(ScreeningDecision{Approved: false}, nil)
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", Body: "body", Confirmed: true}); !errors.Is(err, domain.ErrScreeningRejected) {
		t.Fatalf("got %v", err)
	}
}

func FuzzFingerprintIsDeterministicAndInputBound(f *testing.F) {
	f.Add("c", "a", "body", "media")
	f.Fuzz(func(t *testing.T, c, a, b, m string) {
		one := fingerprint(c, a, b, []string{m}, 1)
		two := fingerprint(c, a, b, []string{m}, 1)
		if one != two || len(one) != 64 {
			t.Fatal("unstable fingerprint")
		}
		if m != "x" && one == fingerprint(c, a, b, []string{"x"}, 1) {
			t.Fatal("media not bound")
		}
	})
}
