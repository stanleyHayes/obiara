package domain

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func issuePair(t *testing.T, sessionID string) (IssuedToken, IssuedToken) {
	t.Helper()
	access, err := IssueAccessToken(sessionID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := IssueRefreshToken(sessionID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return access, refresh
}

func startSession(t *testing.T) Session {
	t.Helper()
	access, refresh := issuePair(t, "sess_test")
	session, err := Start("sess_test", "member-1", "device-1", access, refresh, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func TestStartValidatesRequiredFields(t *testing.T) {
	access, refresh := issuePair(t, "sess_x")
	cases := map[string]struct {
		id, memberID, deviceID string
		wantErr                error
	}{
		"missing id":     {"", "member-1", "device-1", ErrSessionIDRequired},
		"missing member": {"sess_x", " ", "device-1", ErrMemberIDRequired},
		"missing device": {"sess_x", "member-1", "", ErrDeviceIDRequired},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Start(tc.id, tc.memberID, tc.deviceID, access, refresh, testNow); err != tc.wantErr {
				t.Fatalf("Start error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestCheckRefreshAcceptsCurrentToken(t *testing.T) {
	session := startSession(t)
	if err := session.CheckRefresh(session.RefreshTokenHash(), testNow); err != nil {
		t.Fatalf("CheckRefresh current = %v, want nil", err)
	}
}

func TestCheckRefreshRejectsWrongAndReuse(t *testing.T) {
	session := startSession(t)
	oldRefreshHash := session.RefreshTokenHash()

	access, refresh := issuePair(t, "sess_test")
	if err := session.Rotate(access, refresh, testNow); err != nil {
		t.Fatal(err)
	}
	if err := session.CheckRefresh(oldRefreshHash, testNow); err != ErrRefreshReuse {
		t.Fatalf("CheckRefresh rotated-out = %v, want ErrRefreshReuse", err)
	}
	if err := session.CheckRefresh("deadbeef", testNow); err != ErrRefreshTokenMismatch {
		t.Fatalf("CheckRefresh wrong = %v, want ErrRefreshTokenMismatch", err)
	}
}

func TestRotateAndAccessRequireActiveSession(t *testing.T) {
	session := startSession(t)
	session.Revoke(testNow)

	access, refresh := issuePair(t, "sess_test")
	if err := session.Rotate(access, refresh, testNow); err != ErrSessionNotActive {
		t.Fatalf("Rotate revoked = %v, want ErrSessionNotActive", err)
	}
	if err := session.CheckAccess(session.AccessTokenHash(), testNow); err != ErrSessionNotActive {
		t.Fatalf("CheckAccess revoked = %v, want ErrSessionNotActive", err)
	}
	if err := session.CheckRefresh(session.RefreshTokenHash(), testNow); err != ErrSessionNotActive {
		t.Fatalf("CheckRefresh revoked = %v, want ErrSessionNotActive", err)
	}
}

func TestExpiryChecks(t *testing.T) {
	session := startSession(t)
	beyondRefresh := testNow.Add(RefreshTokenLifetime + time.Minute)
	if err := session.CheckRefresh(session.RefreshTokenHash(), beyondRefresh); err != ErrSessionExpired {
		t.Fatalf("CheckRefresh expired = %v, want ErrSessionExpired", err)
	}
	beyondAccess := testNow.Add(AccessTokenLifetime + time.Minute)
	if err := session.CheckAccess(session.AccessTokenHash(), beyondAccess); err != ErrSessionExpired {
		t.Fatalf("CheckAccess expired = %v, want ErrSessionExpired", err)
	}
}

func TestSplitToken(t *testing.T) {
	id, err := SplitToken("sess_abc.secretvalue")
	if err != nil || id != "sess_abc" {
		t.Fatalf("SplitToken = %q, %v", id, err)
	}
	for _, plain := range []string{"", "noseparator", ".secret"} {
		if _, err := SplitToken(plain); err != ErrTokenMalformed {
			t.Fatalf("SplitToken(%q) = %v, want ErrTokenMalformed", plain, err)
		}
	}
}

func TestIssuedTokenHashIsDeterministic(t *testing.T) {
	access, _ := issuePair(t, "sess_h")
	if got := HashToken(access.Plain); got != access.Hash {
		t.Fatalf("HashToken mismatch: %q vs %q", got, access.Hash)
	}
	if access.ExpiresAt.Sub(testNow) != AccessTokenLifetime {
		t.Fatalf("access lifetime = %v", access.ExpiresAt.Sub(testNow))
	}
}
