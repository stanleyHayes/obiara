package domain

import (
	"testing"
	"time"
)

var adminNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestNewPrincipalValidation(t *testing.T) {
	if _, err := NewPrincipal("adm_1", "not-an-email", []Role{RoleVerifier}, adminNow); err != ErrInvalidEmail {
		t.Fatalf("bad email = %v", err)
	}
	if _, err := NewPrincipal("adm_1", "ops@example.test", nil, adminNow); err != ErrNoRoles {
		t.Fatalf("no roles = %v", err)
	}
	if _, err := NewPrincipal("adm_1", "ops@example.test", []Role{Role("superuser")}, adminNow); err != ErrInvalidRole {
		t.Fatalf("bad role = %v", err)
	}
	if _, err := NewPrincipal("adm_1", "ops@example.test", []Role{RoleVerifier, RoleVerifier}, adminNow); err != ErrInvalidRole {
		t.Fatalf("duplicate role = %v", err)
	}
	principal, err := NewPrincipal("adm_1", "ops@example.test", []Role{RoleVerifier, RoleTSAgent}, adminNow)
	if err != nil {
		t.Fatal(err)
	}
	if !principal.HasRole(RoleTSAgent) || principal.HasRole(RoleAdmin) {
		t.Fatalf("roles = %#v", principal.Roles())
	}
}

func TestChallengeFlow(t *testing.T) {
	challenge := NewChallenge("ch_1", "adm_1", "123456", adminNow)
	if err := challenge.Verify("000000", adminNow); err != ErrMfaMismatch {
		t.Fatalf("wrong code = %v", err)
	}
	if challenge.Attempts() != 1 {
		t.Fatalf("attempts = %d", challenge.Attempts())
	}
	if err := challenge.Verify("123456", adminNow); err != nil {
		t.Fatal(err)
	}
	if challenge.ConsumedAt() == nil {
		t.Fatal("challenge must be consumed")
	}
	if err := challenge.Verify("123456", adminNow); err != ErrMfaConsumed {
		t.Fatalf("reuse = %v", err)
	}

	expired := NewChallenge("ch_2", "adm_1", "123456", adminNow)
	if err := expired.Verify("123456", adminNow.Add(MfaLifetime+time.Minute)); err != ErrMfaExpired {
		t.Fatalf("expired = %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	session := NewSession("sess_1", "adm_1", []Role{RoleVerifier}, adminNow)
	if !session.Active(adminNow) {
		t.Fatal("fresh session must be active")
	}
	if session.SteppedUp() {
		t.Fatal("session must not start stepped up")
	}
	if err := session.MarkSteppedUp(adminNow); err != nil {
		t.Fatal(err)
	}
	if !session.SteppedUp() {
		t.Fatal("step-up not recorded")
	}
	session.Revoke()
	if session.Active(adminNow) {
		t.Fatal("revoked session must be inactive")
	}
	if err := session.MarkSteppedUp(adminNow); err != ErrSessionClosed {
		t.Fatalf("step-up on revoked = %v", err)
	}

	expired := NewSession("sess_2", "adm_1", nil, adminNow.Add(-SessionLifetime))
	if expired.Active(adminNow) {
		t.Fatal("expired session must be inactive")
	}
}

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != mfaCodeDigits {
		t.Fatalf("code = %q", code)
	}
}
