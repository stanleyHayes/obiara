package domain

import (
	"errors"
	"strings"
	"time"
)

// Status is the lifecycle state of a session. Transitions are explicit
// (agent_plan.md §7.4): active → revoked, active → expired (by time only).
type Status string

const (
	StatusActive  Status = "active"
	StatusRevoked Status = "revoked"
)

var (
	ErrSessionIDRequired    = errors.New("session id is required")
	ErrMemberIDRequired     = errors.New("member id is required")
	ErrDeviceIDRequired     = errors.New("device id is required")
	ErrSessionNotActive     = errors.New("session is not active")
	ErrSessionExpired       = errors.New("session is expired")
	ErrRefreshTokenMismatch = errors.New("refresh token does not match session")
	// ErrRefreshReuse signals presentation of a rotated-out refresh token,
	// which is treated as token theft: the session must be revoked.
	ErrRefreshReuse = errors.New("rotated refresh token presented")
)

// Session is a rotated-refresh member session bound to one device
// (agent_plan.md §11). It is server-authoritative; clients only hold tokens.
type Session struct {
	id                  string
	memberID            string
	deviceID            string
	status              Status
	accessTokenHash     string
	accessExpiresAt     time.Time
	refreshTokenHash    string
	replacedRefreshHash string
	refreshExpiresAt    time.Time
	version             int64
	createdAt           time.Time
	updatedAt           time.Time
}

// Start creates an active session with its first token pair.
func Start(id, memberID, deviceID string, access, refresh IssuedToken, now time.Time) (Session, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Session{}, ErrSessionIDRequired
	}
	memberID = strings.TrimSpace(memberID)
	if memberID == "" {
		return Session{}, ErrMemberIDRequired
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return Session{}, ErrDeviceIDRequired
	}
	now = now.UTC()
	return Session{
		id:               id,
		memberID:         memberID,
		deviceID:         deviceID,
		status:           StatusActive,
		accessTokenHash:  access.Hash,
		accessExpiresAt:  access.ExpiresAt,
		refreshTokenHash: refresh.Hash,
		refreshExpiresAt: refresh.ExpiresAt,
		version:          1,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// Reconstitute rebuilds a stored session without policy checks. Only
// persistence adapters may call it; expired/revoked sessions must load.
func Reconstitute(id, memberID, deviceID string, status Status, accessHash string, accessExpiresAt time.Time, refreshHash, replacedRefreshHash string, refreshExpiresAt time.Time, version int64, createdAt, updatedAt time.Time) Session {
	return Session{
		id:                  id,
		memberID:            memberID,
		deviceID:            deviceID,
		status:              status,
		accessTokenHash:     accessHash,
		accessExpiresAt:     accessExpiresAt,
		refreshTokenHash:    refreshHash,
		replacedRefreshHash: replacedRefreshHash,
		refreshExpiresAt:    refreshExpiresAt,
		version:             version,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

// Rotate moves the session to a fresh token pair. The rotated-out refresh
// hash is retained so a later presentation of it can be detected as reuse.
func (session *Session) Rotate(access, refresh IssuedToken, now time.Time) error {
	if session.status != StatusActive {
		return ErrSessionNotActive
	}
	if session.RefreshExpired(now) {
		return ErrSessionExpired
	}
	session.accessTokenHash = access.Hash
	session.accessExpiresAt = access.ExpiresAt
	session.replacedRefreshHash = session.refreshTokenHash
	session.refreshTokenHash = refresh.Hash
	session.refreshExpiresAt = refresh.ExpiresAt
	session.updatedAt = now.UTC()
	return nil
}

// RenewAccess replaces only the access token (e.g. after authentication of a
// still-valid refresh-less flow). Kept separate from Rotate so refresh
// rotation always goes through the reuse-detecting path.
func (session *Session) RenewAccess(access IssuedToken, now time.Time) error {
	if session.status != StatusActive {
		return ErrSessionNotActive
	}
	session.accessTokenHash = access.Hash
	session.accessExpiresAt = access.ExpiresAt
	session.updatedAt = now.UTC()
	return nil
}

// Revoke ends the session; revocation is irreversible and audited upstream.
func (session *Session) Revoke(now time.Time) {
	session.status = StatusRevoked
	session.updatedAt = now.UTC()
}

// CheckRefresh validates a presented refresh token against this session.
// A match on the rotated-out hash reports ErrRefreshReuse.
func (session Session) CheckRefresh(refreshHash string, now time.Time) error {
	if session.status != StatusActive {
		return ErrSessionNotActive
	}
	if refreshHash == session.refreshTokenHash {
		if session.RefreshExpired(now) {
			return ErrSessionExpired
		}
		return nil
	}
	if session.replacedRefreshHash != "" && refreshHash == session.replacedRefreshHash {
		return ErrRefreshReuse
	}
	return ErrRefreshTokenMismatch
}

// CheckAccess validates a presented access token hash.
func (session Session) CheckAccess(accessHash string, now time.Time) error {
	if session.status != StatusActive {
		return ErrSessionNotActive
	}
	if accessHash != session.accessTokenHash {
		return ErrRefreshTokenMismatch
	}
	if now.UTC().After(session.accessExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}

func (session Session) RefreshExpired(now time.Time) bool {
	return now.UTC().After(session.refreshExpiresAt)
}

func (session Session) ID() string                  { return session.id }
func (session Session) MemberID() string            { return session.memberID }
func (session Session) DeviceID() string            { return session.deviceID }
func (session Session) Status() Status              { return session.status }
func (session Session) AccessTokenHash() string     { return session.accessTokenHash }
func (session Session) AccessExpiresAt() time.Time  { return session.accessExpiresAt }
func (session Session) RefreshTokenHash() string    { return session.refreshTokenHash }
func (session Session) ReplacedRefreshHash() string { return session.replacedRefreshHash }
func (session Session) RefreshExpiresAt() time.Time { return session.refreshExpiresAt }
func (session Session) Version() int64              { return session.version }
func (session Session) CreatedAt() time.Time        { return session.createdAt }
func (session Session) UpdatedAt() time.Time        { return session.updatedAt }
