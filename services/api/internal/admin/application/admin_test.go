package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

var adminSvcNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func newService(t *testing.T) (AdminService, *MockPrincipalRepository, *MockChallengeRepository, *MockSessionRepository, *MockAccessAudit, *MockCodeSender) {
	t.Helper()
	ctrl := gomock.NewController(t)
	principals := NewMockPrincipalRepository(ctrl)
	challenges := NewMockChallengeRepository(ctrl)
	sessions := NewMockSessionRepository(ctrl)
	audit := NewMockAccessAudit(ctrl)
	sender := NewMockCodeSender(ctrl)
	service := NewAdminService(principals, challenges, sessions, audit, sender, func() time.Time { return adminSvcNow }, func() string { return "adm_test" })
	return service, principals, challenges, sessions, audit, sender
}

func adminPrincipal() domain.Principal {
	return domain.ReconstitutePrincipal("adm_root", "root@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusActive, 1, adminSvcNow)
}

func steppedUpSession(id, principalID string, roles []domain.Role) domain.Session {
	return domain.ReconstituteSession(id, principalID, roles, true, adminSvcNow.Add(time.Hour), false, 2, adminSvcNow)
}

func TestEnrollRequiresAdminRole(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	verifier := domain.ReconstitutePrincipal("adm_v", "v@example.test", []domain.Role{domain.RoleVerifier}, domain.StatusActive, 1, adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_v").Return(steppedUpSession("sess_v", "adm_v", []domain.Role{domain.RoleVerifier}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_v").Return(verifier, nil)

	if _, err := service.Enroll(context.Background(), "sess_v", "new@example.test", []domain.Role{domain.RoleTSAgent}); err != ErrNotAdmin {
		t.Fatalf("Enroll = %v, want ErrNotAdmin (FR-801 least privilege)", err)
	}
}

func TestEnrollCreatesAndAudits(t *testing.T) {
	service, principals, _, sessions, audit, _ := newService(t)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	audit.EXPECT().Append(gomock.Any(), "adm_root", "admin.enroll", "adm_test", gomock.Any()).Return(nil)

	principal, err := service.Enroll(context.Background(), "sess_root", "new@example.test", []domain.Role{domain.RoleTSAgent})
	if err != nil {
		t.Fatal(err)
	}
	if principal.ID() != "adm_test" {
		t.Fatalf("principal = %#v", principal)
	}
}

func TestLoginFlow(t *testing.T) {
	service, principals, challenges, sessions, _, sender := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(adminPrincipal(), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(nil)

	if err := service.StartLogin(context.Background(), "root@example.test"); err != nil {
		t.Fatal(err)
	}

	// Complete: code matches the challenge hash for "654321".
	challenge := domain.NewChallenge("ch_1", "adm_root", "654321", adminSvcNow)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(adminPrincipal(), nil)
	challenges.EXPECT().LatestForPrincipal(gomock.Any(), "adm_root").Return(challenge, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	sessions.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, session domain.Session) error {
			if session.PrincipalID() != "adm_root" || !session.Active(adminSvcNow) {
				t.Fatalf("session = %#v", session)
			}
			return nil
		})

	session, err := service.CompleteLogin(context.Background(), "root@example.test", "654321")
	if err != nil {
		t.Fatal(err)
	}
	if session.SteppedUp() {
		t.Fatal("login session must not be stepped up")
	}
}

func TestCompleteLoginWrongCodePersistsAttempt(t *testing.T) {
	service, principals, challenges, _, _, _ := newService(t)
	challenge := domain.NewChallenge("ch_1", "adm_root", "654321", adminSvcNow)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(adminPrincipal(), nil)
	challenges.EXPECT().LatestForPrincipal(gomock.Any(), "adm_root").Return(challenge, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Challenge) error {
			if updated.Attempts() != 1 {
				t.Fatalf("attempts = %d", updated.Attempts())
			}
			return nil
		})

	if _, err := service.CompleteLogin(context.Background(), "root@example.test", "000000"); err != domain.ErrMfaMismatch {
		t.Fatalf("CompleteLogin = %v, want mismatch", err)
	}
}

func TestStepUpMarksSessionAndAudits(t *testing.T) {
	service, principals, challenges, sessions, audit, sender := newService(t)
	session := domain.NewSession("sess_1", "adm_root", []domain.Role{domain.RoleAdmin}, adminSvcNow)

	sessions.EXPECT().FindByID(gomock.Any(), "sess_1").Return(session, nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(nil)

	if err := service.StepUpStart(context.Background(), "sess_1"); err != nil {
		t.Fatal(err)
	}

	challenge := domain.NewChallenge("ch_2", "adm_root", "111111", adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_1").Return(session, nil)
	challenges.EXPECT().LatestForPrincipal(gomock.Any(), "adm_root").Return(challenge, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	sessions.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Session) error {
			if !updated.SteppedUp() {
				t.Fatal("session must be stepped up")
			}
			return nil
		})
	audit.EXPECT().Append(gomock.Any(), "adm_root", "admin.step_up", "sess_1", gomock.Any()).Return(nil)

	if _, err := service.StepUpComplete(context.Background(), "sess_1", "111111"); err != nil {
		t.Fatal(err)
	}
}

func TestAdminRoleChangeNeedsDistinctSteppedUpApprover(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	target := domain.ReconstitutePrincipal("adm_target", "target@example.test", []domain.Role{domain.RoleVerifier}, domain.StatusActive, 4, adminSvcNow)
	first := steppedUpSession("sess_first", "adm_root", []domain.Role{domain.RoleAdmin})
	sessions.EXPECT().FindByID(gomock.Any(), first.ID()).Return(first, nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindByID(gomock.Any(), target.ID()).Return(target, nil)
	principals.EXPECT().CreateRoleChange(gomock.Any(), gomock.Any()).Return(nil)

	change, err := service.ProposeAdminRoleChange(
		context.Background(), first.ID(), target.ID(),
		[]domain.Role{domain.RoleVerifier, domain.RoleAdmin},
		"Grant command-centre coverage",
	)
	if err != nil {
		t.Fatal(err)
	}

	// The proposing principal cannot approve its own change, even with a
	// stepped-up session.
	sessions.EXPECT().FindByID(gomock.Any(), first.ID()).Return(first, nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindRoleChange(gomock.Any(), change.ID()).Return(change, nil)
	principals.EXPECT().FindByID(gomock.Any(), target.ID()).Return(target, nil)
	if _, err := service.ApproveAdminRoleChange(context.Background(), first.ID(), change.ID()); err != domain.ErrSameApprover {
		t.Fatalf("self approval = %v, want ErrSameApprover", err)
	}

	secondPrincipal := domain.ReconstitutePrincipal("adm_second", "second@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusActive, 1, adminSvcNow)
	second := steppedUpSession("sess_second", secondPrincipal.ID(), []domain.Role{domain.RoleAdmin})
	sessions.EXPECT().FindByID(gomock.Any(), second.ID()).Return(second, nil)
	principals.EXPECT().FindByID(gomock.Any(), secondPrincipal.ID()).Return(secondPrincipal, nil)
	principals.EXPECT().FindRoleChange(gomock.Any(), change.ID()).Return(change, nil)
	principals.EXPECT().FindByID(gomock.Any(), target.ID()).Return(target, nil)
	principals.EXPECT().ApproveRoleChange(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	updated, err := service.ApproveAdminRoleChange(context.Background(), second.ID(), change.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !updated.HasRole(domain.RoleAdmin) {
		t.Fatal("approved target must hold admin role")
	}
}
