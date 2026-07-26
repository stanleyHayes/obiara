package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var actionNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func inReviewCase(tier domain.Tier) domain.Case {
	return domain.ReconstituteCase("case_1", "rep_1", "m-2", tier, domain.QueueTriage, actionNow.Add(24*time.Hour), domain.CaseInReview, "agent-1", 2, actionNow, nil)
}

func newActionService(t *testing.T) (ActionService, *MockCaseRepository, *MockActionLog, *MockIdentityEnforcement, *MockSessionRevoker, *MockDeviceBlocklister) {
	t.Helper()
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	actions := NewMockActionLog(ctrl)
	identity := NewMockIdentityEnforcement(ctrl)
	sessions := NewMockSessionRevoker(ctrl)
	devices := NewMockDeviceBlocklister(ctrl)
	service := NewActionService(cases, actions, identity, sessions, devices, func() time.Time { return actionNow }, func() string { return "act_test" })
	return service, cases, actions, identity, sessions, devices
}

func TestApplyBanPropagatesEverywhere(t *testing.T) {
	service, cases, actions, identity, sessions, devices := newActionService(t)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(inReviewCase(domain.TierA), nil)
	actions.EXPECT().CountForSubject(gomock.Any(), "m-2").Return(0, nil)
	identity.EXPECT().Block(gomock.Any(), "m-2").Return(nil)
	devices.EXPECT().Blocklist(gomock.Any(), "m-2", "ban:case_1", gomock.Any()).Return(nil)
	sessions.EXPECT().RevokeMemberSessions(gomock.Any(), "m-2").Return(nil)
	actions.EXPECT().Append(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, record domain.ActionRecord) error {
			if record.Action != domain.ActionBan || record.Priors != 0 || record.ActorID != "agent-1" {
				t.Fatalf("record = %#v", record)
			}
			return nil
		})
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, safetyCase domain.Case) error {
			if safetyCase.Status() != domain.CaseResolved {
				t.Fatalf("case = %#v", safetyCase)
			}
			return nil
		})

	if err := service.Apply(context.Background(), "case_1", domain.ActionBan, "agent-1"); err != nil {
		t.Fatal(err)
	}
}

func TestApplySuspensionComputesExpiry(t *testing.T) {
	service, cases, actions, identity, sessions, _ := newActionService(t)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(inReviewCase(domain.TierB), nil)
	actions.EXPECT().CountForSubject(gomock.Any(), "m-2").Return(0, nil)
	identity.EXPECT().Suspend(gomock.Any(), "m-2", actionNow.Add(30*24*time.Hour)).Return(nil)
	sessions.EXPECT().RevokeMemberSessions(gomock.Any(), "m-2").Return(nil)
	actions.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	if err := service.Apply(context.Background(), "case_1", domain.ActionSuspend30d, "agent-1"); err != nil {
		t.Fatal(err)
	}
}

func TestApplyWarningTouchesNothing(t *testing.T) {
	service, cases, actions, _, _, _ := newActionService(t)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(inReviewCase(domain.TierC), nil)
	actions.EXPECT().CountForSubject(gomock.Any(), "m-2").Return(0, nil)
	actions.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	// No identity/session/device expectations: warnings change nothing.

	if err := service.Apply(context.Background(), "case_1", domain.ActionWarning, "agent-1"); err != nil {
		t.Fatal(err)
	}
}

func TestLadderViolationLeavesNoTrace(t *testing.T) {
	service, cases, actions, _, _, _ := newActionService(t)
	cases.EXPECT().FindByID(gomock.Any(), "case_1").Return(inReviewCase(domain.TierC), nil)
	actions.EXPECT().CountForSubject(gomock.Any(), "m-2").Return(0, nil)
	// No Append/identity/session expectations.

	if err := service.Apply(context.Background(), "case_1", domain.ActionBan, "agent-1"); err == nil {
		t.Fatal("ban on first tier-C must be rejected by the ladder")
	}
}
