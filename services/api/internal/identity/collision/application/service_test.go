package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/domain"
)

func TestSharedDeviceCollisionIsDeniedByDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, keyer, func() time.Time { return now })

	keyer.EXPECT().Key("collision:shared_device", "device-token").Return("signal-key", nil)
	keyer.EXPECT().Key("collision:subject", "account-2").Return("subject-key", nil)
	repository.EXPECT().RegisterSignal(gomock.Any(), domain.KindSharedDevice, "signal-key", "subject-key").Return(true, nil)
	keyer.EXPECT().Key("collision:case", "shared_device:signal-key:subject-key").Return("case-key", nil)
	repository.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, reviewCase domain.Case, audit domain.AuditEvent) (domain.Case, bool, error) {
			if audit.Sequence != 1 || audit.ActorKey != "system" {
				t.Fatalf("invalid creation audit: %+v", audit)
			}
			return reviewCase, true, nil
		},
	)

	decision, err := service.Assess(context.Background(), Assessment{
		Kind: domain.KindSharedDevice, Signal: "device-token", SubjectID: "account-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || !decision.Collision || !decision.ReviewRequired {
		t.Fatalf("shared-device collision was not fail-closed: %+v", decision)
	}
}

func TestKnownNameIsNormalizedBeforePrivacyKeying(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	service := NewService(repository, keyer, time.Now)

	keyer.EXPECT().Key("collision:known_name", "ama mensah").Return("name-key", nil)
	keyer.EXPECT().Key("collision:subject", "account-2").Return("subject-key", nil)
	repository.EXPECT().RegisterSignal(gomock.Any(), domain.KindKnownName, "name-key", "subject-key").Return(false, nil)

	decision, err := service.Assess(context.Background(), Assessment{
		Kind: domain.KindKnownName, Signal: "  AMA   Mensah ", SubjectID: "account-2",
	})
	if err != nil || !decision.Allowed || decision.ReviewRequired {
		t.Fatalf("first known-name observation = %+v, %v", decision, err)
	}
}

func TestResolutionPersistsOneDeterministicAudit(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC)
	service := NewService(repository, keyer, func() time.Time { return now })
	reviewCase, _, err := domain.NewCase("case-key", domain.KindSharedDevice, "signal-key", "subject-key", now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repository.EXPECT().FindByID(gomock.Any(), "case-key").Return(reviewCase, nil)
	keyer.EXPECT().Key("collision:actor", "operator@example.test").Return("actor-key", nil)
	repository.EXPECT().Resolve(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).DoAndReturn(
		func(_ context.Context, resolved domain.Case, audit domain.AuditEvent, _ int64) error {
			if resolved.Status() != domain.StatusApproved || audit.Sequence != 2 ||
				audit.ActorKey != "actor-key" || audit.ReasonCode != "household_confirmed" {
				t.Fatalf("unexpected resolution: %+v %+v", resolved, audit)
			}
			return nil
		},
	)

	decision, err := service.Resolve(context.Background(), "case-key", domain.ResolutionApprove, "household_confirmed", "operator@example.test")
	if err != nil || !decision.Allowed {
		t.Fatalf("resolve = %+v, %v", decision, err)
	}
}
