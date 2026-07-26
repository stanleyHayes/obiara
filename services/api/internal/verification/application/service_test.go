package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var testDOB = time.Date(1998, time.April, 12, 0, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return testNow }

func newService(t *testing.T) (VerificationService, *MockCaseRepository, *MockVerificationProvider, *MockTierTransitions) {
	t.Helper()
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	provider := NewMockVerificationProvider(ctrl)
	tiers := NewMockTierTransitions(ctrl)
	service := NewVerificationService(cases, provider, tiers, fixedNow, func() string { return "vc_test" })
	return service, cases, provider, tiers
}

func TestMatchApprovesAndPromotes(t *testing.T) {
	service, cases, provider, tiers := newService(t)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "match", ProviderRef: "ref-1", Reason: "issuer match"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			if c.Status() != domain.StatusApproved {
				t.Fatalf("status = %q", c.Status())
			}
			return nil
		})
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), gomock.Any()).Return(nil)

	verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", verificationCase.Status())
	}
}

func TestMismatchRejects(t *testing.T) {
	service, cases, provider, _ := newService(t)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "mismatch", ProviderRef: "ref-2", Reason: "no issuer record"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			if c.Status() != domain.StatusRejected {
				t.Fatalf("status = %q", c.Status())
			}
			return nil
		})

	if _, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-X", testDOB); !errors.Is(err, ErrProviderRejected) {
		t.Fatalf("SubmitGhanaCard = %v, want ErrProviderRejected", err)
	}
}

func TestOutageAndUncertaintyQueueManual(t *testing.T) {
	for name, setup := range map[string]func(*MockVerificationProvider){
		"outage": func(provider *MockVerificationProvider) {
			provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(ProviderResult{}, context.DeadlineExceeded)
		},
		"uncertain": func(provider *MockVerificationProvider) {
			provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(ProviderResult{Outcome: "uncertain", Reason: "unreadable"}, nil)
		},
	} {
		t.Run(name, func(t *testing.T) {
			service, cases, provider, _ := newService(t)
			cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			setup(provider)
			cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, c domain.VerificationCase) error {
					if c.Status() != domain.StatusQueuedManual {
						t.Fatalf("status = %q, want queued_manual", c.Status())
					}
					return nil
				})

			verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
			if err != nil {
				t.Fatal(err)
			}
			if verificationCase.Status() != domain.StatusQueuedManual {
				t.Fatalf("status = %q", verificationCase.Status())
			}
		})
	}
}

func TestDecideManualApprovalPromotes(t *testing.T) {
	service, cases, _, tiers := newService(t)
	queued := domain.ReconstituteCase("vc_1", "id_1", "GHA-1", domain.StatusQueuedManual, "", "provider uncertain", testDOB, 2, testNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "vc_1").Return(queued, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), "agent-1").Return(nil)

	verificationCase, err := service.DecideManual(context.Background(), "vc_1", true, "verified in person", "agent-1")
	if err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", verificationCase.Status())
	}
}
