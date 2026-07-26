package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var caseSvcNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestOpenCreatesCaseWithSLA(t *testing.T) {
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	service := NewCaseService(cases, func() time.Time { return caseSvcNow }, func() string { return "case_test" })

	report := domain.ReconstituteReport("rep_1", "m-1", "m-2", domain.CategoryFraud, domain.TierA, domain.SurfaceRoom, "", "", domain.StatusReceived, 1, caseSvcNow)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, safetyCase domain.Case) error {
			if safetyCase.SLADueAt() != caseSvcNow.Add(8*time.Hour) || safetyCase.Queue() != domain.QueueTriage {
				t.Fatalf("case = %#v", safetyCase)
			}
			return nil
		})

	safetyCase, err := service.Open(context.Background(), report)
	if err != nil || safetyCase.ID() != "case_test" {
		t.Fatalf("open = %#v, %v", safetyCase, err)
	}
}

func TestAssignAndResolve(t *testing.T) {
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	service := NewCaseService(cases, func() time.Time { return caseSvcNow }, func() string { return "case_test" })

	queued := domain.ReconstituteCase("case_1", "rep_1", "m-2", domain.TierB, domain.QueueTriage, caseSvcNow.Add(24*time.Hour), domain.CaseQueued, "", 1, caseSvcNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(queued, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, safetyCase domain.Case) error {
			if safetyCase.Status() != domain.CaseInReview || safetyCase.AssignedTo() != "agent-1" {
				t.Fatalf("case = %#v", safetyCase)
			}
			return nil
		})

	if _, err := service.Assign(context.Background(), "case_1", "agent-1"); err != nil {
		t.Fatal(err)
	}

	inReview := domain.ReconstituteCase("case_1", "rep_1", "m-2", domain.TierB, domain.QueueTriage, caseSvcNow.Add(24*time.Hour), domain.CaseInReview, "agent-1", 2, caseSvcNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(inReview, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, safetyCase domain.Case) error {
			if safetyCase.Status() != domain.CaseResolved {
				t.Fatalf("case = %#v", safetyCase)
			}
			return nil
		})

	if _, err := service.Resolve(context.Background(), "case_1", "suspension 14d", "agent-1"); err != nil {
		t.Fatal(err)
	}
}
