package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// MFA policy (E16-S01): 6-digit codes, 5-minute expiry, bounded attempts.
const (
	MfaLifetime     = 5 * time.Minute
	MfaMaxAttempts  = 5
	SessionLifetime = 30 * time.Minute
	mfaCodeDigits   = 6
)

var (
	ErrMfaExpired    = errors.New("mfa code is expired")
	ErrMfaMismatch   = errors.New("mfa code does not match")
	ErrMfaAttempts   = errors.New("mfa attempts exceeded")
	ErrMfaConsumed   = errors.New("mfa challenge already consumed")
	ErrSessionClosed = errors.New("admin session is not active")
)

// Challenge is one MFA code challenge (login or step-up).
type Challenge struct {
	id          string
	principalID string
	codeHash    string
	expiresAt   time.Time
	attempts    int
	createdAt   time.Time
	consumedAt  *time.Time
}

// NewChallenge builds a challenge for a freshly minted code.
func NewChallenge(id, principalID, plainCode string, now time.Time) Challenge {
	now = now.UTC()
	return Challenge{
		id:          id,
		principalID: principalID,
		codeHash:    HashCode(plainCode),
		expiresAt:   now.Add(MfaLifetime),
		createdAt:   now,
	}
}

// ReconstituteChallenge rebuilds a stored challenge without checks.
func ReconstituteChallenge(id, principalID, codeHash string, expiresAt time.Time, attempts int, createdAt time.Time, consumedAt *time.Time) Challenge {
	return Challenge{id: id, principalID: principalID, codeHash: codeHash, expiresAt: expiresAt, attempts: attempts, createdAt: createdAt, consumedAt: consumedAt}
}

// Verify checks a code, recording failed attempts.
func (challenge *Challenge) Verify(plainCode string, now time.Time) error {
	if challenge.consumedAt != nil {
		return ErrMfaConsumed
	}
	if challenge.attempts >= MfaMaxAttempts {
		return ErrMfaAttempts
	}
	if now.UTC().After(challenge.expiresAt) {
		return ErrMfaExpired
	}
	if HashCode(plainCode) != challenge.codeHash {
		challenge.attempts++
		return ErrMfaMismatch
	}
	consumed := now.UTC()
	challenge.consumedAt = &consumed
	return nil
}

// Session is a short-lived admin session with an optional step-up flag
// (step-up is required for sensitive desks and four-eyes actions).
type Session struct {
	id          string
	principalID string
	roles       []Role
	steppedUp   bool
	expiresAt   time.Time
	revoked     bool
	version     int64
	createdAt   time.Time
}

// NewSession issues an admin session.
func NewSession(id, principalID string, roles []Role, now time.Time) Session {
	now = now.UTC()
	return Session{
		id:          id,
		principalID: principalID,
		roles:       roles,
		expiresAt:   now.Add(SessionLifetime),
		version:     1,
		createdAt:   now,
	}
}

// ReconstituteSession rebuilds a stored session without checks.
func ReconstituteSession(id, principalID string, roles []Role, steppedUp bool, expiresAt time.Time, revoked bool, version int64, createdAt time.Time) Session {
	return Session{id: id, principalID: principalID, roles: roles, steppedUp: steppedUp, expiresAt: expiresAt, revoked: revoked, version: version, createdAt: createdAt}
}

// Active reports whether the session is usable now.
func (session Session) Active(now time.Time) bool {
	return !session.revoked && now.UTC().Before(session.expiresAt)
}

// MarkSteppedUp records successful step-up verification.
func (session *Session) MarkSteppedUp(now time.Time) error {
	if !session.Active(now) {
		return ErrSessionClosed
	}
	session.steppedUp = true
	session.version++
	return nil
}

// Revoke closes the session.
func (session *Session) Revoke() {
	session.revoked = true
	session.version++
}

func (session Session) ID() string           { return session.id }
func (session Session) PrincipalID() string  { return session.principalID }
func (session Session) Roles() []Role        { return session.roles }
func (session Session) SteppedUp() bool      { return session.steppedUp }
func (session Session) ExpiresAt() time.Time { return session.expiresAt }
func (session Session) Revoked() bool        { return session.revoked }
func (session Session) Version() int64       { return session.version }
func (session Session) CreatedAt() time.Time { return session.createdAt }

func (challenge Challenge) ID() string             { return challenge.id }
func (challenge Challenge) PrincipalID() string    { return challenge.principalID }
func (challenge Challenge) CodeHash() string       { return challenge.codeHash }
func (challenge Challenge) ExpiresAt() time.Time   { return challenge.expiresAt }
func (challenge Challenge) Attempts() int          { return challenge.attempts }
func (challenge Challenge) CreatedAt() time.Time   { return challenge.createdAt }
func (challenge Challenge) ConsumedAt() *time.Time { return challenge.consumedAt }

// GenerateCode mints a 6-digit MFA code.
func GenerateCode() (string, error) {
	code := make([]byte, mfaCodeDigits)
	for i := range code {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate mfa code: %w", err)
		}
		code[i] = byte('0' + digit.Int64())
	}
	return string(code), nil
}

// HashCode derives the stored form of a code.
func HashCode(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
