package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	// ErrStaleSession reports an optimistic-concurrency conflict: another
	// actor (e.g. a racing refresh on a second device) persisted first.
	ErrStaleSession = errors.New("session changed concurrently")
)

// IssuedSession is the one-time view handed to the client: plaintext tokens
// plus the public session fields. Tokens are never recoverable afterwards.
type IssuedSession struct {
	Session      domain.Session
	AccessToken  string
	RefreshToken string
}

// SessionService issues, rotates, authenticates and revokes member sessions
// (agent_plan.md §11). Every command carries actor metadata upstream; this
// service is the server-authoritative state boundary.
type SessionService struct {
	sessions SessionRepository
	now      func() time.Time
	newID    func() string
}

func NewSessionService(sessions SessionRepository, now func() time.Time, newID func() string) SessionService {
	return SessionService{sessions: sessions, now: now, newID: newID}
}

// Issue creates a new session for a verified member on a device.
func (service SessionService) Issue(ctx context.Context, memberID, deviceID string) (IssuedSession, error) {
	sessionID := service.newID()
	now := service.now()
	access, err := domain.IssueAccessToken(sessionID, now)
	if err != nil {
		return IssuedSession{}, err
	}
	refresh, err := domain.IssueRefreshToken(sessionID, now)
	if err != nil {
		return IssuedSession{}, err
	}
	session, err := domain.Start(sessionID, memberID, deviceID, access, refresh, now)
	if err != nil {
		return IssuedSession{}, err
	}
	if err := service.sessions.Create(ctx, session); err != nil {
		return IssuedSession{}, err
	}
	return IssuedSession{Session: session, AccessToken: access.Plain, RefreshToken: refresh.Plain}, nil
}

// Refresh rotates the refresh token. Presentation of a rotated-out token is
// reuse (theft): the session is revoked and the error reports the reuse.
func (service SessionService) Refresh(ctx context.Context, refreshPlain string) (IssuedSession, error) {
	session, err := service.findByToken(ctx, refreshPlain)
	if err != nil {
		return IssuedSession{}, err
	}

	now := service.now()
	if err := session.CheckRefresh(domain.HashToken(refreshPlain), now); err != nil {
		if errors.Is(err, domain.ErrRefreshReuse) {
			session.Revoke(now)
			_ = service.sessions.Update(ctx, session)
		}
		return IssuedSession{}, err
	}

	access, err := domain.IssueAccessToken(session.ID(), now)
	if err != nil {
		return IssuedSession{}, err
	}
	refresh, err := domain.IssueRefreshToken(session.ID(), now)
	if err != nil {
		return IssuedSession{}, err
	}
	if err := session.Rotate(access, refresh, now); err != nil {
		return IssuedSession{}, err
	}
	if err := service.sessions.Update(ctx, session); err != nil {
		return IssuedSession{}, err
	}
	return IssuedSession{Session: session, AccessToken: access.Plain, RefreshToken: refresh.Plain}, nil
}

// Authenticate validates an access token and returns its session.
func (service SessionService) Authenticate(ctx context.Context, accessPlain string) (domain.Session, error) {
	session, err := service.findByToken(ctx, accessPlain)
	if err != nil {
		return domain.Session{}, err
	}
	if err := session.CheckAccess(domain.HashToken(accessPlain), service.now()); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// Revoke ends one session (sign-out, step-up enforcement).
func (service SessionService) Revoke(ctx context.Context, sessionID string) error {
	session, err := service.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	session.Revoke(service.now())
	return service.sessions.Update(ctx, session)
}

// RevokeMemberSessions ends every session of a member (account-takeover
// response, deletion, Tier-A action propagation).
func (service SessionService) RevokeMemberSessions(ctx context.Context, memberID string) error {
	return service.sessions.RevokeAllForMember(ctx, memberID, service.now())
}

// RevokeDeviceSessions ends every session bound to a lost or risky device.
func (service SessionService) RevokeDeviceSessions(ctx context.Context, deviceID string) error {
	return service.sessions.RevokeAllForDevice(ctx, deviceID, service.now())
}

func (service SessionService) findByToken(ctx context.Context, plain string) (domain.Session, error) {
	sessionID, err := domain.SplitToken(plain)
	if err != nil {
		return domain.Session{}, err
	}
	session, err := service.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return domain.Session{}, fmt.Errorf("find session: %w", err)
	}
	return session, nil
}
