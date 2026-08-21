package application

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

const testPassword = "Correct-Horse-Battery-9"

func passwordPrincipal(t *testing.T) domain.Principal {
	t.Helper()
	hash, err := domain.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	return domain.ReconstitutePrincipalWithPassword(
		"adm_root", "root@example.test", []domain.Role{domain.RoleAdmin},
		domain.StatusActive, hash, 1, adminSvcNow)
}

// TestStartLoginRequiresThePasswordWhenSet is the point of the password
// factor: control of the operator's mailbox alone must no longer be enough
// to obtain an admin session.
func TestStartLoginRequiresThePasswordWhenSet(t *testing.T) {
	service, principals, _, _, _, _ := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(passwordPrincipal(t), nil)

	// gomock fails the test on any unexpected call, so the absence of
	// Create/SendMfaCode expectations asserts that no code was minted or sent.
	if err := service.StartLogin(context.Background(), "root@example.test", "wrong-password"); err != nil {
		t.Fatalf("StartLogin = %v, want nil", err)
	}
}

func TestStartLoginRejectsAnEmptyPasswordWhenSet(t *testing.T) {
	service, principals, _, _, _, _ := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(passwordPrincipal(t), nil)

	if err := service.StartLogin(context.Background(), "root@example.test", ""); err != nil {
		t.Fatalf("StartLogin = %v, want nil", err)
	}
}

func TestStartLoginSendsTheCodeOnTheCorrectPassword(t *testing.T) {
	service, principals, challenges, _, _, sender := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(passwordPrincipal(t), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(nil)

	if err := service.StartLogin(context.Background(), "root@example.test", testPassword); err != nil {
		t.Fatalf("StartLogin = %v, want nil", err)
	}
}

// TestStartLoginKeepsCodeOnlyFlowForLegacyPrincipals stops the password
// factor from locking out operators enrolled before it existed.
func TestStartLoginKeepsCodeOnlyFlowForLegacyPrincipals(t *testing.T) {
	service, principals, challenges, _, _, sender := newService(t)
	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(adminPrincipal(), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(nil)

	if err := service.StartLogin(context.Background(), "root@example.test", ""); err != nil {
		t.Fatalf("StartLogin = %v, want nil", err)
	}
}

// TestStartLoginIsNotAPasswordOracle pins the non-enumeration property: a
// wrong password, an unknown operator and a suspended operator are
// indistinguishable to an unauthenticated caller.
func TestStartLoginIsNotAPasswordOracle(t *testing.T) {
	suspended := domain.ReconstitutePrincipalWithPassword(
		"adm_s", "s@example.test", []domain.Role{domain.RoleAdmin},
		domain.StatusSuspended, passwordPrincipal(t).PasswordHash(), 1, adminSvcNow)

	cases := map[string]struct {
		email    string
		password string
		stub     func(*MockPrincipalRepository)
	}{
		"unknown operator": {
			email: "ghost@example.test", password: testPassword,
			stub: func(principals *MockPrincipalRepository) {
				principals.EXPECT().FindByEmail(gomock.Any(), "ghost@example.test").
					Return(domain.Principal{}, ErrPrincipalNotFound)
			},
		},
		"wrong password": {
			email: "root@example.test", password: "nope",
			stub: func(principals *MockPrincipalRepository) {
				principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").
					Return(passwordPrincipal(t), nil)
			},
		},
		"suspended operator": {
			email: "s@example.test", password: testPassword,
			stub: func(principals *MockPrincipalRepository) {
				principals.EXPECT().FindByEmail(gomock.Any(), "s@example.test").Return(suspended, nil)
			},
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			service, principals, _, _, _, _ := newService(t)
			testCase.stub(principals)
			if err := service.StartLogin(context.Background(), testCase.email, testCase.password); err != nil {
				t.Fatalf("StartLogin = %v, want nil so callers cannot tell the cases apart", err)
			}
		})
	}
}
