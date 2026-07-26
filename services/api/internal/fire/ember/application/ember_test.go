package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/domain"
)

var emberNow = time.Date(2026, time.July, 26, 22, 0, 0, 0, time.UTC)

func newService(t *testing.T) (EmberService, *MockEmberRepository, *MockAttendanceChecker, *MockDoorwayOpener) {
	t.Helper()
	ctrl := gomock.NewController(t)
	embers := NewMockEmberRepository(ctrl)
	attendance := NewMockAttendanceChecker(ctrl)
	opener := NewMockDoorwayOpener(ctrl)
	return NewEmberService(embers, attendance, opener, func() time.Time { return emberNow }, func() string { return "ember_test" }), embers, attendance, opener
}

func TestIssueRequiresCoAttendance(t *testing.T) {
	service, _, attendance, _ := newService(t)
	attendance.EXPECT().Attended(gomock.Any(), "fire_1", "m-1").Return(true, nil)
	attendance.EXPECT().Attended(gomock.Any(), "fire_1", "m-2").Return(false, nil)

	if _, err := service.Issue(context.Background(), "fire_1", "m-1", "m-2"); err != ErrNotCoAttendee {
		t.Fatalf("non-attendee = %v, want ErrNotCoAttendee (FR-402)", err)
	}
}

func TestIssueOneWay(t *testing.T) {
	service, embers, attendance, _ := newService(t)
	attendance.EXPECT().Attended(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
	embers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	embers.EXPECT().FindDirected(gomock.Any(), "fire_1", "m-2", "m-1").Return(domain.Ember{}, ErrEmberNotFound)

	ember, err := service.Issue(context.Background(), "fire_1", "m-1", "m-2")
	if err != nil {
		t.Fatal(err)
	}
	if ember.Status() != domain.StatusIssued {
		t.Fatalf("status = %q", ember.Status())
	}
}

func TestMutualEmberOpensDoorway(t *testing.T) {
	service, embers, attendance, opener := newService(t)
	attendance.EXPECT().Attended(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
	embers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	reverse, _ := domain.NewEmber("ember_rev", "fire_1", "m-2", "m-1", emberNow)
	embers.EXPECT().FindDirected(gomock.Any(), "fire_1", "m-2", "m-1").Return(reverse, nil)
	embers.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	opener.EXPECT().OpenForPair(gomock.Any(), "m-1", "m-2", "ember:ember_test").Return(nil)

	ember, err := service.Issue(context.Background(), "fire_1", "m-1", "m-2")
	if err != nil {
		t.Fatal(err)
	}
	if ember.Status() != domain.StatusMutual {
		t.Fatalf("status = %q, want mutual", ember.Status())
	}
}

func TestRedeemPersists(t *testing.T) {
	service, embers, _, _ := newService(t)
	issued := domain.ReconstituteEmber("ember_1", "fire_1", "m-1", "m-2", domain.StatusIssued, emberNow.Add(time.Hour), 1, emberNow, nil)
	embers.EXPECT().FindByID(gomock.Any(), "ember_1").Return(issued, nil)
	embers.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ember domain.Ember) error {
			if ember.Status() != domain.StatusRedeemed {
				t.Fatalf("status = %q", ember.Status())
			}
			return nil
		})

	if _, err := service.Redeem(context.Background(), "ember_1", "m-2"); err != nil {
		t.Fatal(err)
	}
}

func TestRedeemExpiredStillPersistsState(t *testing.T) {
	service, embers, _, _ := newService(t)
	issued := domain.ReconstituteEmber("ember_1", "fire_1", "m-1", "m-2", domain.StatusIssued, emberNow.Add(-time.Hour), 1, emberNow.Add(-25*time.Hour), nil)
	embers.EXPECT().FindByID(gomock.Any(), "ember_1").Return(issued, nil)
	embers.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ember domain.Ember) error {
			if ember.Status() != domain.StatusExpired {
				t.Fatalf("status = %q, want expired recorded", ember.Status())
			}
			return nil
		})

	if _, err := service.Redeem(context.Background(), "ember_1", "m-2"); err != domain.ErrEmberExpired {
		t.Fatalf("redeem = %v, want expired", err)
	}
}
