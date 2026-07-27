package application

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

func scheduledFireFixture(capacity, going int) domain.Fire {
	return domain.ReconstituteFire("fire_1", "host-1", "circle-1", "Friday Fire", fireStart, capacity, going, domain.StatusScheduled, 1, fireNow)
}

func TestCloseToEmbersReturnsRoster(t *testing.T) {
	service, fires := newService(t)
	fires.EXPECT().FindByID(gomock.Any(), "fire_1").Return(scheduledFireFixture(50, 2), nil)
	fires.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fire domain.Fire) error {
			if fire.Status() != domain.StatusEmbers {
				t.Fatalf("status = %q", fire.Status())
			}
			return nil
		})
	going := []domain.RSVP{
		domain.ReconstituteRSVP("fire_1", "m-1", domain.RSVPGoing, 0, 1, fireNow),
		domain.ReconstituteRSVP("fire_1", "m-2", domain.RSVPGoing, 0, 1, fireNow),
	}
	fires.EXPECT().ListGoing(gomock.Any(), "fire_1").Return(going, nil)

	attendees, err := service.CloseToEmbers(context.Background(), "fire_1", "host-1")
	if err != nil || len(attendees) != 2 {
		t.Fatalf("attendees = %#v, %v", attendees, err)
	}
}

func TestCloseToEmbersRejectsNonHost(t *testing.T) {
	service, fires := newService(t)
	fires.EXPECT().FindByID(gomock.Any(), "fire_1").Return(scheduledFireFixture(50, 2), nil)
	// No UpdateStatus expectation.

	if _, err := service.CloseToEmbers(context.Background(), "fire_1", "intruder"); err != domain.ErrNotHost {
		t.Fatalf("non-host close = %v, want ErrNotHost", err)
	}
}
