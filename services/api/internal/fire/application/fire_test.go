package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

var fireNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var fireStart = fireNow.Add(6 * time.Hour)

func newService(t *testing.T) (FireService, *MockFireRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	fires := NewMockFireRepository(ctrl)
	return NewFireService(fires, func() time.Time { return fireNow }, func() string { return "fire_test" }), fires
}

func TestScheduleValidatesAndPersists(t *testing.T) {
	service, fires := newService(t)
	fires.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fire domain.Fire) error {
			if fire.ID() != "fire_test" || fire.Capacity() != 50 || fire.Status() != domain.StatusScheduled {
				t.Fatalf("fire = %#v", fire)
			}
			return nil
		})

	fire, err := service.Schedule(context.Background(), "host-1", "circle-1", "Friday Fire", fireStart, 50)
	if err != nil {
		t.Fatal(err)
	}
	if fire.ID() != "fire_test" {
		t.Fatalf("id = %q", fire.ID())
	}
}

func TestScheduleRejectsInvalid(t *testing.T) {
	service, _ := newService(t)
	if _, err := service.Schedule(context.Background(), "host-1", "circle-1", "F", fireStart, 0); err != domain.ErrInvalidCapacity {
		t.Fatalf("capacity 0 = %v", err)
	}
	if _, err := service.Schedule(context.Background(), " ", "circle-1", "F", fireStart, 10); err != domain.ErrHostRequired {
		t.Fatalf("blank host = %v", err)
	}
}

func TestRSVPRequiresTier1(t *testing.T) {
	service, _ := newService(t)
	// No AdmitTx expectation: tier 0 never reaches persistence (FR-401).
	if _, err := service.RSVP(context.Background(), "fire_1", "m-1", 0); err != domain.ErrTierTooLow {
		t.Fatalf("tier 0 = %v, want ErrTierTooLow (FR-401)", err)
	}
}

func TestRSVPDelegatesToAtomicAdmission(t *testing.T) {
	service, fires := newService(t)
	want := domain.ReconstituteRSVP("fire_1", "m-1", domain.RSVPWaitlisted, 3, 1, fireNow)
	fires.EXPECT().AdmitTx(gomock.Any(), "fire_1", "m-1", fireNow).Return(want, nil)

	got, err := service.RSVP(context.Background(), "fire_1", "m-1", 1)
	if err != nil || got != want {
		t.Fatalf("rsvp = %#v, %v", got, err)
	}
}

func TestCancelDelegatesToAtomicCancellation(t *testing.T) {
	service, fires := newService(t)
	promoted := domain.ReconstituteRSVP("fire_1", "m-2", domain.RSVPGoing, 0, 2, fireNow)
	fires.EXPECT().CancelTx(gomock.Any(), "fire_1", "m-1", fireNow).Return(&promoted, nil)

	got, err := service.Cancel(context.Background(), "fire_1", "m-1")
	if err != nil || got == nil || got.MemberID() != "m-2" {
		t.Fatalf("promoted = %#v, %v", got, err)
	}
}
