package adminauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
)

type stubAdmin struct {
	session   admindomain.Session
	principal admindomain.Principal
	err       error
}

func (stub stubAdmin) Authenticate(context.Context, string) (admindomain.Session, admindomain.Principal, error) {
	return stub.session, stub.principal, stub.err
}

func principalWith(t *testing.T, roles ...admindomain.Role) admindomain.Principal {
	t.Helper()
	principal, err := admindomain.NewPrincipal("adm_1", "op@obiara.test", roles, time.Now())
	if err != nil {
		t.Fatalf("NewPrincipal: %v", err)
	}
	return principal
}

func TestTrustAndSafetyRolesMayReview(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleTSAgent, admindomain.RoleAdmin} {
		authority := NewAuthority(stubAdmin{principal: principalWith(t, role)})
		if err := authority.Authorize(context.Background(), "sess_1", application.CapQueue); err != nil {
			t.Errorf("%s was refused: %v", role, err)
		}
	}
}

func TestOtherRolesMayNotReview(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleFinance, admindomain.RoleVerifier, admindomain.RoleHost} {
		authority := NewAuthority(stubAdmin{principal: principalWith(t, role)})
		if err := authority.Authorize(context.Background(), "sess_1", application.CapDecide); !errors.Is(err, application.ErrDenied) {
			t.Errorf("%s reached the desk: %v", role, err)
		}
	}
}

// TestSteppedUpSessionSatisfiesTheGate covers the MFA bridge: evidence
// access and decisions reuse the console's own step-up rather than keeping
// a second notion of recency.
func TestSteppedUpSessionSatisfiesTheGate(t *testing.T) {
	now := time.Now()
	session := admindomain.ReconstituteSession("ses_1", "adm_1", nil, true, now.Add(time.Hour), false, 2, now)
	gate := NewMFAGate(stubAdmin{session: session, principal: principalWith(t, admindomain.RoleTSAgent)})
	if !gate.Recent(context.Background(), "sess_1", now) {
		t.Error("a stepped-up, active session did not satisfy the gate")
	}
}

func TestSessionWithoutStepUpFailsTheGate(t *testing.T) {
	now := time.Now()
	session := admindomain.ReconstituteSession("ses_1", "adm_1", nil, false, now.Add(time.Hour), false, 1, now)
	gate := NewMFAGate(stubAdmin{session: session})
	if gate.Recent(context.Background(), "sess_1", now) {
		t.Error("a session that never stepped up satisfied the gate")
	}
}

// TestExpiredSessionFailsTheGate matters because the stepped-up flag
// outlives the session that earned it.
func TestExpiredSessionFailsTheGate(t *testing.T) {
	now := time.Now()
	expired := admindomain.ReconstituteSession("ses_1", "adm_1", nil, true, now.Add(-time.Minute), false, 2, now)
	gate := NewMFAGate(stubAdmin{session: expired})
	if gate.Recent(context.Background(), "sess_1", now) {
		t.Error("an expired but stepped-up session satisfied the gate")
	}
}

func TestUnreadableSessionFailsBoth(t *testing.T) {
	stub := stubAdmin{err: errors.New("session not found")}
	if err := NewAuthority(stub).Authorize(context.Background(), "x", application.CapRead); !errors.Is(err, application.ErrDenied) {
		t.Errorf("authorize error = %v, want ErrDenied", err)
	}
	if NewMFAGate(stub).Recent(context.Background(), "x", time.Now()) {
		t.Error("an unreadable session satisfied the MFA gate")
	}
}
