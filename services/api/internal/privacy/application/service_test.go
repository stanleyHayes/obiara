package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/privacy/domain"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return testNow }

func newService(t *testing.T) (PrivacyService, *MockRequestRepository, *MockLegalHoldRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	requests := NewMockRequestRepository(ctrl)
	holds := NewMockLegalHoldRepository(ctrl)
	return NewPrivacyService(requests, holds, fixedNow, func() string { return "pr_test" }), requests, holds
}

func TestRequestExportOpens(t *testing.T) {
	service, requests, _ := newService(t)
	requests.EXPECT().FindOpenByAccountAndKind(gomock.Any(), "id_1", domain.KindExport).Return(domain.PrivacyRequest{}, ErrRequestNotFound)
	requests.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	request, err := service.RequestExport(context.Background(), "id_1")
	if err != nil {
		t.Fatal(err)
	}
	if request.Kind() != domain.KindExport || request.DueAt() != testNow.Add(domain.ExportDueWithin) {
		t.Fatalf("request = %#v", request)
	}
}

func TestDuplicateOpenRequestRejected(t *testing.T) {
	service, requests, _ := newService(t)
	existing, _ := domain.NewRequest("pr_1", "id_1", domain.KindExport, testNow)
	requests.EXPECT().FindOpenByAccountAndKind(gomock.Any(), "id_1", domain.KindExport).Return(existing, nil)

	if _, err := service.RequestExport(context.Background(), "id_1"); err != ErrOpenRequestExists {
		t.Fatalf("RequestExport = %v, want ErrOpenRequestExists", err)
	}
}

func TestDeletionBlockedByLegalHold(t *testing.T) {
	service, _, holds := newService(t)
	hold, _ := domain.NewLegalHold("id_1", "court order", "agent-1", testNow)
	holds.EXPECT().ActiveFor(gomock.Any(), "id_1").Return(hold, nil)

	if _, err := service.RequestDeletion(context.Background(), "id_1"); err != domain.ErrLegalHoldActive {
		t.Fatalf("RequestDeletion = %v, want ErrLegalHoldActive", err)
	}
}

func TestPlaceHoldBlocksOpenDeletion(t *testing.T) {
	service, requests, holds := newService(t)
	holds.EXPECT().Place(gomock.Any(), gomock.Any()).Return(nil)
	open, _ := domain.NewRequest("pr_1", "id_1", domain.KindDeletion, testNow)
	requests.EXPECT().FindOpenByAccountAndKind(gomock.Any(), "id_1", domain.KindDeletion).Return(open, nil)
	requests.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, request domain.PrivacyRequest) error {
			if request.Status() != domain.StatusBlocked {
				t.Fatalf("status = %q, want blocked", request.Status())
			}
			return nil
		})

	if err := service.PlaceHold(context.Background(), "id_1", "court order", "agent-1"); err != nil {
		t.Fatal(err)
	}
}

func TestProcessorExecutesDeletion(t *testing.T) {
	ctrl := gomock.NewController(t)
	requests := NewMockRequestRepository(ctrl)
	assembler := NewMockExportAssembler(ctrl)
	erasure := NewMockErasureRunner(ctrl)
	sessions := NewMockSessionRevoker(ctrl)
	processor := NewProcessor(requests, assembler, erasure, sessions, fixedNow)

	open, _ := domain.NewRequest("pr_1", "id_1", domain.KindDeletion, testNow)
	requests.EXPECT().NextExecutable(gomock.Any(), 5).Return([]domain.PrivacyRequest{open}, nil)
	requests.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	erasure.EXPECT().Erase(gomock.Any(), "pr_1", "id_1").Return(nil)
	sessions.EXPECT().RevokeMemberSessions(gomock.Any(), "id_1").Return(nil)

	if err := processor.RunBatch(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
}

func TestProcessorExportDoesNotErase(t *testing.T) {
	ctrl := gomock.NewController(t)
	requests := NewMockRequestRepository(ctrl)
	assembler := NewMockExportAssembler(ctrl)
	erasure := NewMockErasureRunner(ctrl)
	sessions := NewMockSessionRevoker(ctrl)
	processor := NewProcessor(requests, assembler, erasure, sessions, fixedNow)

	open, _ := domain.NewRequest("pr_1", "id_1", domain.KindExport, testNow)
	requests.EXPECT().NextExecutable(gomock.Any(), 5).Return([]domain.PrivacyRequest{open}, nil)
	requests.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	assembler.EXPECT().Assemble(gomock.Any(), "pr_1", "id_1").Return("archive://pr_1", nil)
	// No Erase or RevokeMemberSessions expectation: any call fails.

	if err := processor.RunBatch(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
}

func TestProcessorResumesIdempotentProcessingAfterWorkerRestart(t *testing.T) {
	ctrl := gomock.NewController(t)
	requests := NewMockRequestRepository(ctrl)
	assembler := NewMockExportAssembler(ctrl)
	processor := NewProcessor(requests, assembler, NewMockErasureRunner(ctrl), NewMockSessionRevoker(ctrl), fixedNow)

	processing, _ := domain.NewRequest("pr_resume", "id_1", domain.KindExport, testNow)
	if err := processing.StartProcessing(); err != nil {
		t.Fatal(err)
	}
	requests.EXPECT().NextExecutable(gomock.Any(), 5).Return([]domain.PrivacyRequest{processing}, nil)
	assembler.EXPECT().Assemble(gomock.Any(), "pr_resume", "id_1").Return("privacy-export:pr_resume", nil)
	requests.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, request domain.PrivacyRequest) error {
		if request.Status() != domain.StatusCompleted {
			t.Fatalf("status = %q", request.Status())
		}
		return nil
	})

	if err := processor.RunBatch(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
}
