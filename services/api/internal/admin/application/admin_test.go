package application

import (
	"context"
	"errors"
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
	service, principals, _, sessions, _, _ := newService(t)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	// The mutation and the audit entry commit atomically in the repository.
	principals.EXPECT().CreateWithAudit(gomock.Any(), gomock.Any(), "adm_root", "admin.enroll", gomock.Any()).Return(nil)

	principal, err := service.Enroll(context.Background(), "sess_root", "new@example.test", []domain.Role{domain.RoleTSAgent})
	if err != nil {
		t.Fatal(err)
	}
	if principal.ID() != "adm_test" {
		t.Fatalf("principal = %#v", principal)
	}
}

func TestChangeStatusSuspendsAtomicallyWithLastAdminGuard(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	target := domain.ReconstitutePrincipal("adm_target", "target@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusActive, 3, adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_target").Return(target, nil)
	principals.EXPECT().UpdateWithAudit(gomock.Any(), gomock.Any(), true, "adm_root", "admin.principal.status.suspended", gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Principal, guard bool, _ string, _ string, _ time.Time) error {
			if !guard {
				t.Fatal("suspending an active admin must carry the last-admin guard")
			}
			if updated.Status() != domain.StatusSuspended {
				t.Fatalf("status = %s", updated.Status())
			}
			return nil
		})

	updated, err := service.ChangeStatus(context.Background(), "sess_root", "adm_target", domain.StatusSuspended)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status() != domain.StatusSuspended {
		t.Fatalf("status = %s", updated.Status())
	}
}

func TestChangeStatusLastAdminErrorPropagatesFromGuardedTransaction(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	target := domain.ReconstitutePrincipal("adm_target", "target@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusActive, 3, adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_target").Return(target, nil)
	// The repository enforces the invariant inside the transaction.
	principals.EXPECT().UpdateWithAudit(gomock.Any(), gomock.Any(), true, gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrLastAdmin)

	if _, err := service.ChangeStatus(context.Background(), "sess_root", "adm_target", domain.StatusSuspended); err != ErrLastAdmin {
		t.Fatalf("ChangeStatus = %v, want ErrLastAdmin", err)
	}
}

func TestChangeStatusReactivateSkipsLastAdminGuard(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	target := domain.ReconstitutePrincipal("adm_target", "target@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusSuspended, 3, adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_target").Return(target, nil)
	principals.EXPECT().UpdateWithAudit(gomock.Any(), gomock.Any(), false, "adm_root", "admin.principal.status.active", gomock.Any()).Return(nil)

	updated, err := service.ChangeStatus(context.Background(), "sess_root", "adm_target", domain.StatusActive)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status() != domain.StatusActive {
		t.Fatalf("status = %s", updated.Status())
	}
}

func TestProposeAdminRoleChangeGuardsLastAdminAtomically(t *testing.T) {
	service, principals, _, sessions, _, _ := newService(t)
	target := domain.ReconstitutePrincipal("adm_target", "target@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusActive, 3, adminSvcNow)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").Return(steppedUpSession("sess_root", "adm_root", []domain.Role{domain.RoleAdmin}), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_target").Return(target, nil)
	principals.EXPECT().CreateRoleChange(gomock.Any(), gomock.Any(), true).Return(ErrLastAdmin)

	_, err := service.ProposeAdminRoleChange(
		context.Background(), "sess_root", "adm_target",
		[]domain.Role{domain.RoleVerifier},
		"Rotate off command centre",
	)
	if err != ErrLastAdmin {
		t.Fatalf("ProposeAdminRoleChange = %v, want ErrLastAdmin", err)
	}
}

func TestStartLoginNoOpsForUnknownOrSuspendedEmails(t *testing.T) {
	service, principals, _, _, _, _ := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "ghost@example.test").Return(domain.Principal{}, ErrPrincipalNotFound)
	suspended := domain.ReconstitutePrincipal("adm_s", "s@example.test", []domain.Role{domain.RoleAdmin}, domain.StatusSuspended, 1, adminSvcNow)
	principals.EXPECT().FindByEmail(gomock.Any(), "s@example.test").Return(suspended, nil)

	// No challenge is minted and no code is sent (gomock fails on unexpected
	// calls), so the endpoint cannot enumerate valid principals.
	if err := service.StartLogin(context.Background(), "ghost@example.test", ""); err != nil {
		t.Fatalf("unknown email: StartLogin = %v, want nil", err)
	}
	if err := service.StartLogin(context.Background(), "s@example.test", ""); err != nil {
		t.Fatalf("suspended email: StartLogin = %v, want nil", err)
	}
}

func TestStartLoginPropagatesStoreErrors(t *testing.T) {
	service, principals, _, _, _, _ := newService(t)
	storeErr := errors.New("mongodb unavailable")
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(domain.Principal{}, storeErr)

	if err := service.StartLogin(context.Background(), "root@example.test", ""); !errors.Is(err, storeErr) {
		t.Fatalf("StartLogin = %v, want store error", err)
	}
}

func TestLoginFlow(t *testing.T) {
	service, principals, challenges, sessions, _, sender := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(adminPrincipal(), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(nil)

	if err := service.StartLogin(context.Background(), "root@example.test", ""); err != nil {
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
	principals.EXPECT().CreateRoleChange(gomock.Any(), gomock.Any(), false).Return(nil)

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
