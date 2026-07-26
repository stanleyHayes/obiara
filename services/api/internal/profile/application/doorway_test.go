package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

var doorwayNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestSetDoorwayQuestionUpserts(t *testing.T) {
	ctrl := gomock.NewController(t)
	questions := NewMockDoorwayRepository(ctrl)
	service := NewDoorwayService(questions, func() time.Time { return doorwayNow })

	questions.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, question domain.DoorwayQuestion) error {
			if question.Text() != "What does home mean to you?" || !question.Custom() {
				t.Fatalf("question = %#v", question)
			}
			return nil
		})

	if _, err := service.Set(context.Background(), "m-1", "  What does home mean to you? ", true); err != nil {
		t.Fatal(err)
	}
}

func TestSetDoorwayQuestionRejectsUnsafe(t *testing.T) {
	ctrl := gomock.NewController(t)
	questions := NewMockDoorwayRepository(ctrl)
	service := NewDoorwayService(questions, func() time.Time { return doorwayNow })
	// No Upsert expectation: unsafe input must not persist.

	if _, err := service.Set(context.Background(), "m-1", "text me +233 55 000 0101", true); err != domain.ErrUnsafeProfile {
		t.Fatalf("Set = %v, want ErrUnsafeProfile", err)
	}
}

func TestVaultAddRespectsCapacity(t *testing.T) {
	ctrl := gomock.NewController(t)
	vault := NewMockVaultRepository(ctrl)
	service := NewVaultService(vault, func() time.Time { return doorwayNow }, func() string { return "vi_test" })

	vault.EXPECT().CountByMember(gomock.Any(), "m-1").Return(domain.VaultCapacity, nil)
	if _, err := service.Add(context.Background(), "m-1", "asset-1", 0); err != domain.ErrVaultFull {
		t.Fatalf("Add at capacity = %v, want ErrVaultFull", err)
	}

	vault.EXPECT().CountByMember(gomock.Any(), "m-1").Return(3, nil)
	vault.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)
	if _, err := service.Add(context.Background(), "m-1", "asset-1", 3); err != nil {
		t.Fatal(err)
	}
}

func TestVaultViewAppliesVeil(t *testing.T) {
	ctrl := gomock.NewController(t)
	vault := NewMockVaultRepository(ctrl)
	service := NewVaultService(vault, func() time.Time { return doorwayNow }, func() string { return "vi_test" })

	item, _ := domain.NewVaultItem("vi_1", "m-1", "asset-1", 0, doorwayNow)
	vault.EXPECT().ListByMember(gomock.Any(), "m-1").Return([]domain.VaultItem{item}, nil)

	views, err := service.ViewFor(context.Background(), "m-1", "m-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || !views[0].Veiled {
		t.Fatalf("stranger view = %#v", views)
	}

	vault.EXPECT().ListByMember(gomock.Any(), "m-1").Return([]domain.VaultItem{item}, nil)
	views, _ = service.ViewFor(context.Background(), "m-1", "m-1")
	if views[0].Veiled {
		t.Fatal("owner view must be clear")
	}
}
