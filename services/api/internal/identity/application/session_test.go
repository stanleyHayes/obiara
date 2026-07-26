package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return testNow }

func newService(t *testing.T) (SessionService, *MockSessionRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repository := NewMockSessionRepository(ctrl)
	return NewSessionService(repository, fixedNow, func() string { return "sess_test" }), repository
}

func storedSession(t *testing.T) domain.Session {
	t.Helper()
	access, err := domain.IssueAccessToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := domain.IssueRefreshToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	session, err := domain.Start("sess_test", "member-1", "device-1", access, refresh, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func TestIssueCreatesSessionWithTokenPair(t *testing.T) {
	service, repository := newService(t)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, session domain.Session) error {
			if session.MemberID() != "member-1" || session.DeviceID() != "device-1" {
				t.Fatalf("session = %#v", session)
			}
			if session.Status() != domain.StatusActive {
				t.Fatalf("status = %q", session.Status())
			}
			return nil
		})

	issued, err := service.Issue(context.Background(), "member-1", "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if issued.AccessToken == "" || issued.RefreshToken == "" {
		t.Fatal("plaintext tokens must be returned exactly once")
	}
	if domain.HashToken(issued.AccessToken) != issued.Session.AccessTokenHash() {
		t.Fatal("access token hash mismatch")
	}
	if domain.HashToken(issued.RefreshToken) != issued.Session.RefreshTokenHash() {
		t.Fatal("refresh token hash mismatch")
	}
}

func TestRefreshRotatesTokens(t *testing.T) {
	service, repository := newService(t)
	session := storedSession(t)
	oldRefreshPlain, err := domain.IssueRefreshToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a session whose stored refresh hash matches oldRefreshPlain.
	access, _ := domain.IssueAccessToken("sess_test", testNow)
	session, err = domain.Start("sess_test", "member-1", "device-1", access, oldRefreshPlain, testNow)
	if err != nil {
		t.Fatal(err)
	}

	repository.EXPECT().FindByID(gomock.Any(), "sess_test").Return(session, nil)
	repository.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Session) error {
			if updated.RefreshTokenHash() == session.RefreshTokenHash() {
				t.Fatal("refresh hash must rotate")
			}
			if updated.ReplacedRefreshHash() != session.RefreshTokenHash() {
				t.Fatal("rotated-out hash must be retained for reuse detection")
			}
			return nil
		})

	issued, err := service.Refresh(context.Background(), oldRefreshPlain.Plain)
	if err != nil {
		t.Fatal(err)
	}
	if issued.RefreshToken == oldRefreshPlain.Plain {
		t.Fatal("refresh token must rotate")
	}
}

func TestRefreshReuseRevokesSession(t *testing.T) {
	service, repository := newService(t)
	session := storedSession(t)
	replaced, err := domain.IssueRefreshToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	access, _ := domain.IssueAccessToken("sess_test", testNow)
	current, err := domain.IssueRefreshToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Rotate(access, current, testNow); err != nil {
		t.Fatal(err)
	}
	// Force the rotated-out hash to be the one the attacker replays.
	session = domain.Reconstitute(session.ID(), session.MemberID(), session.DeviceID(), session.Status(),
		session.AccessTokenHash(), session.AccessExpiresAt(), session.RefreshTokenHash(),
		replaced.Hash, session.RefreshExpiresAt(), session.Version(), session.CreatedAt(), session.UpdatedAt())

	repository.EXPECT().FindByID(gomock.Any(), "sess_test").Return(session, nil)
	repository.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Session) error {
			if updated.Status() != domain.StatusRevoked {
				t.Fatalf("reuse must revoke the session, status = %q", updated.Status())
			}
			return nil
		})

	if _, err := service.Refresh(context.Background(), replaced.Plain); !errors.Is(err, domain.ErrRefreshReuse) {
		t.Fatalf("Refresh reuse error = %v, want ErrRefreshReuse", err)
	}
}

func TestAuthenticateValidatesAccessToken(t *testing.T) {
	service, repository := newService(t)
	session := storedSession(t)
	repository.EXPECT().FindByID(gomock.Any(), "sess_test").Return(session, nil)

	access, err := domain.IssueAccessToken("sess_test", testNow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), access.Plain); !errors.Is(err, domain.ErrRefreshTokenMismatch) {
		t.Fatalf("Authenticate wrong secret = %v, want mismatch", err)
	}
	if _, err := service.Authenticate(context.Background(), "garbage"); !errors.Is(err, domain.ErrTokenMalformed) {
		t.Fatalf("Authenticate malformed = %v, want ErrTokenMalformed", err)
	}
}

func TestRevokeDeviceDelegates(t *testing.T) {
	service, repository := newService(t)
	repository.EXPECT().RevokeAllForDevice(gomock.Any(), "device-9", testNow).Return(nil)
	if err := service.RevokeDeviceSessions(context.Background(), "device-9"); err != nil {
		t.Fatal(err)
	}
}
