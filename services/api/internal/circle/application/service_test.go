package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

func TestServiceRequestsMembershipAndMapsConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	now := testNow().Add(time.Minute)
	service := NewService(repository, func() time.Time { return now })
	circle := applicationCircle(t)
	command := Command{ID: "request-1", CircleID: "circle-1", ActorID: "member-1", ExpectedRevision: 1}

	repository.EXPECT().Find(gomock.Any(), "circle-1").Return(circle, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1), "request-1").Return(nil)
	result, err := service.Request(context.Background(), command)
	if err != nil || result.Circle.Allows("member-1", domain.CapabilityView) {
		t.Fatalf("result = %#v, error = %v", result, err)
	}

	command.ID = "request-2"
	command.ActorID = "member-2"
	repository.EXPECT().Find(gomock.Any(), "circle-1").Return(circle, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1), "request-2").Return(ErrOptimisticConflict)
	if _, err := service.Request(context.Background(), command); !errors.Is(err, domain.ErrStaleRevision) {
		t.Fatalf("error = %v, want %v", err, domain.ErrStaleRevision)
	}
}

func TestServiceAccessIsDenyByDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	service := NewService(repository, time.Now)
	circle := applicationCircle(t)
	repository.EXPECT().Find(gomock.Any(), "circle-1").Return(circle, nil).Times(2)
	allowed, err := service.Allows(context.Background(), "circle-1", "unknown", domain.CapabilityView)
	if err != nil || allowed {
		t.Fatalf("unknown access = %v, error = %v", allowed, err)
	}
	allowed, err = service.Allows(context.Background(), "circle-1", "owner-1", domain.CapabilityManage)
	if err != nil || !allowed {
		t.Fatalf("owner access = %v, error = %v", allowed, err)
	}
}

func applicationCircle(t *testing.T) domain.Circle {
	t.Helper()
	circle, err := domain.Create("circle-1", domain.TypeCampus, "owner-1", domain.Command{
		ID: "create-1", ActorID: "owner-1", Kind: "circle.create",
		Payload: string(domain.TypeCampus), RecordedAt: testNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return circle
}

func testNow() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
